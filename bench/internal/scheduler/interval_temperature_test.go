package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//

func TestFindIntervals(t *testing.T) {
	var (
		// 2024/04/01がテスト開始日
		testStartUnix int64 = 1711897200
		testStartAt         = time.Unix(testStartUnix, 0)
	)

	// 5時間分取っておく
	it, err := newIntervalTemperture(testStartUnix, 1, 5)
	assert.NoError(t, err)

	var (
		intervalTestStart1 = testStartAt
		intervalTestEnd1   = testStartAt.Add(1 * time.Hour)
	)
	it.addInterval(intervalTestStart1.Unix(), intervalTestEnd1.Unix())
	intervals := it.findIntervals(func(i uint64) bool {
		return i == 1
	})
	assert.Len(t, intervals, 1)

	var (
		intervalTestStart2 = testStartAt.Add(3 * time.Hour)
		intervalTestEnd2   = testStartAt.Add(4 * time.Hour)
	)
	it.addInterval(intervalTestStart2.Unix(), intervalTestEnd2.Unix())
	intervals = it.findIntervals(func(i uint64) bool {
		return i == 1
	})
	assert.Len(t, intervals, 2)

	intervals = it.findIntervals(func(i uint64) bool {
		return i == 0
	})
	assert.Len(t, intervals, 1)
	assert.Equal(t, testStartAt.Add(2*time.Hour).Unix(), intervals[0].startAt.Unix())
	assert.Equal(t, testStartAt.Add(2*time.Hour).Unix(), intervals[0].endAt.Unix())

	it.addInterval(testStartAt.Add(2*time.Hour).Unix(), testStartAt.Add(2*time.Hour).Unix())
	intervals = it.findIntervals(func(i uint64) bool {
		return i == 0
	})
	assert.Len(t, intervals, 0)
	intervals = it.findIntervals(func(i uint64) bool {
		return i == 1
	})
	assert.Len(t, intervals, 1)
	assert.Equal(t, testStartAt.Unix(), intervals[0].startAt.Unix())
	assert.Equal(t, testStartAt.Add(4*time.Hour).Unix(), intervals[0].endAt.Unix())
}

// startAt = endAtの場合の扱いに関するテストを書く

// FIXME: 枠数2 (maxTemperature>=2)のテストを書く

// FIXME: ちらばって区間を追加したあと、Coldな区間を取得するテスト
//        もれなく区間を得られることをチェック
