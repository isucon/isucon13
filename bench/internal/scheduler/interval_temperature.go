package scheduler

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Intervals は、区間長でソート可能な区間列です
type Intervals []*Interval

func (i Intervals) Len() int           { return len(i) }
func (i Intervals) Swap(j, k int)      { i[j], i[k] = i[k], i[j] }
func (i Intervals) Less(j, k int) bool { return i[j].Hours() < i[k].Hours() }

type Interval struct {
	startAt time.Time
	endAt   time.Time
}

// Hours は、当該区間が時間単位で何時間の区間であるかを返します
func (i *Interval) Hours() int {
	return int(i.endAt.Sub(i.startAt) / time.Hour)
}

// HACK: ロックが著しくパフォーマンスを損なうようなら、24hなど利用区間最大長のレンジでロックを取る構造を用意する
//		 一旦は大きなロックで対処

// IntervalTemperatures は、区間の温度を管理します
type IntervalTemperatures struct {
	mu                   sync.RWMutex
	baseAt               int64
	maxTemperature       int64
	intervalTemperatures []uint64
}

func newIntervalTemperture(baseAt int64, maxTemperature int64, length int) (*IntervalTemperatures, error) {
	// NOTE: webappの予約数がこれより少なくなることが無くなったのでチェックを入れておく
	// findColdShortInterval, findColdLongInterval, findHotIntervalなどがこの数を前提に組まれてる
	//
	if maxTemperature < 2 {
		return nil, fmt.Errorf("maxTemperture must be larger than 2")
	}

	// 1h単位のカウンタを初期化
	tempertures := make([]uint64, length)
	return &IntervalTemperatures{
		baseAt:               baseAt,
		maxTemperature:       maxTemperature,
		intervalTemperatures: tempertures,
	}, nil
}

func (t *IntervalTemperatures) addInterval(startAtUnix int64, endAtUnix int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var (
		baseAt     = time.Unix(t.baseAt, 0)
		startAt    = time.Unix(startAtUnix, 0)
		baseOffset = startAt.Sub(baseAt) / time.Hour
		length     = time.Unix(endAtUnix, 0).Sub(startAt) / time.Hour
	)
	for i := baseOffset; i <= baseOffset+length && int(i) < len(t.intervalTemperatures); i++ {
		t.intervalTemperatures[i]++
	}
}

// NOTE: 1インデックス進むごと1hour進むことに注意
func (t *IntervalTemperatures) findIntervals(fn func(uint64) bool) Intervals {
	var intervals Intervals
	var (
		left, right int
		turn        int
	)
	for left < len(t.intervalTemperatures) && right <= len(t.intervalTemperatures) {
		if turn%2 == 0 { // 左 (turnは偶数)
			// 条件を満たす要素を発見するまで、左を進める
			for left < len(t.intervalTemperatures) {
				leftElem := t.intervalTemperatures[left]
				if fn(leftElem) {
					break
				}
				left++
			}
		} else { // 右 (turnは奇数)
			// 条件を満たさない要素を発見するまで、右を進める
			// NOTE: 半開区間なので、場合によってrightが配列長と等しくなることに注意
			for right <= len(t.intervalTemperatures) {
				if right == len(t.intervalTemperatures) {
					break
				}
				rightElem := t.intervalTemperatures[right]
				if !fn(rightElem) {
					break
				}
				right++
			}
		}

		if turn%2 == 1 {
			// 右を進め終わったら、[left, right)の半開区間確定
			leftHour := time.Duration(left) * time.Hour
			var rightHour time.Duration
			if left == 0 && right-left == 1 {
				rightHour = time.Duration(right) * time.Hour
			} else {
				rightHour = time.Duration(right-1) * time.Hour
			}
			var (
				startAt = time.Unix(t.baseAt, 0).Add(leftHour)
				endAt   = time.Unix(t.baseAt, 0).Add(rightHour)
			)
			intervals = append(intervals, &Interval{
				startAt: startAt,
				endAt:   endAt,
			})
		}

		// 片方をもう一方に追いつかせる
		if turn%2 == 0 {
			// 右を追いつかせる
			right = left
		} else {
			// 左を追いつかせる
			left = right
		}

		turn++
	}

	sort.Sort(intervals)
	return intervals
}

// FIXME: hotについて、interval temperatureにおいてはshort/longの区別をつけず、見つかった配列を返すだけにする
// 受け手側で、先頭を取るか末尾を取るかでshort/longを選べば良い
// 候補枯渇するとダメなので、こちら側で絞ってはいけない

// findHotShortInterval は、残り枠数が１である枠のうち時間の長いほうから消費していきます
func (t *IntervalTemperatures) findHotIntervals() ([]*Interval, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	intervals := t.findIntervals(func(temperature uint64) bool {
		// 予約枠がのこり１つしか無いものを探す
		return temperature == uint64(t.maxTemperature-1)
	})
	if len(intervals) == 0 {
		return nil, fmt.Errorf("no hot long interval")
	}

	return intervals, nil
}

// findColdInterval は、残り枠数が１になるようにまんべんなく予約区間を探し出します
func (t *IntervalTemperatures) findColdIntervals() ([]*Interval, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// NOTE: 同時配信枠が１の場合何もしない考慮を入れた実装
	// ref. https://github.com/isucon/isucon13/blob/18609cfa37c978cb3b6499a20d211cc28f5e7097/bench/internal/scheduler/interval_temperature.go#L155
	intervals := t.findIntervals(func(temperature uint64) bool {
		return temperature < uint64(t.maxTemperature-1)
	})
	if len(intervals) == 0 {
		return nil, fmt.Errorf("no cold interval")
	}

	return intervals, nil
}
