package scheduler

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

// フェーズに応じて、指定された種別のユーザを返す
// こういうユーザがほしいという使い方

// 基本的に、予約に関しては配信者が多いほうがよく
//           投げ銭などスコアの天井を上げる場合は視聴者が多いほうがいい
// とりあえずイーブンに半分に分割する
// シードデータの時点で分けておこう

var UserScheduler = mustNewUserScheduler()

type UserType int

const (
	UserType_Normal  UserType = iota
	UserType_Popular          // 人気
)

type User struct {
	UserId         int
	Name           string
	DisplayName    string
	Description    string
	RawPassword    string
	HashedPassword string

	Type UserType
}

type userScheduler struct {
	PopularLimit int

	vtuberCursorMu sync.Mutex
	vtuberCursor   int

	popularVtuberCursorMu sync.Mutex
	popularVtuberCursor   int

	viewerCursorMu sync.Mutex
	viewerCursor   int

	negativeCountsMu sync.RWMutex
	negativeCounts   []int
}

func mustNewUserScheduler() *userScheduler {
	sched := new(userScheduler)
	// 人気配信者制限
	sched.PopularLimit = 10

	// negative
	sched.negativeCounts = make([]int, len(vtuberPool)+10)

	// 人気配信者を設定
	offset := rand.Intn(len(vtuberPool) - sched.PopularLimit)
	for i := offset; i < offset+sched.PopularLimit; i++ {
		vtuberPool[i].Type = UserType_Popular
	}

	return sched
}

// 負荷レベルを上げる
// 負荷フェーズの切替時、mainからこれを呼び出して負荷レベルを上昇させる
func IncreaseWorkloadLevel(populars int) {
	for i := 0; i < len(vtuberPool); i++ {
		if vtuberPool[i].Type == UserType_Normal {
			if populars > 0 {
				vtuberPool[i].Type = UserType_Popular
				populars--
			} else {
				return
			}
		}
	}
}

// 特定のユーザがトラブルメーカーとして振る舞うべきか判定する
func (u *userScheduler) BehaveTroubleMaker(viewer *User) bool {
	u.negativeCountsMu.RLock()
	defer u.negativeCountsMu.RUnlock()

	const maxNegativeCount = 100

	if viewer.UserId <= 0 || viewer.UserId >= len(u.negativeCounts) {
		return false
	}
	negativeCount := u.negativeCounts[viewer.UserId]

	// 100程度のリクエスト失敗以降は同等に扱う
	// 0 ~ 10の値を取るようになるので、負数を除いて2割程度は最低限正常な振る舞いをするように残しておく
	negativeCount = int(math.Min(float64(negativeCount), maxNegativeCount))
	negativeValue := math.Sqrt(float64(negativeCount))

	r := rand.Intn(int(math.Sqrt(maxNegativeCount)))
	return r >= int(math.Max(negativeValue-2, 0))
}

// 配信者を人気と決定づける要因はなにか？
//   - ライブコメントが集まるところ
//   - 投げ銭がたくさん投げられてるところ
//
// 人気に仕立て上げるかどうかはすべてこちらの采配次第
// 実際に人気であるか (投稿数、スパム数などをもとに判断)を判定して返す
func (u *userScheduler) IsPopular(user *User) bool {
	return false
}

// 人気になる候補を取得。人気に仕立てていく
func (s *userScheduler) SelectPopularCandidate() (*User, error) {
	s.popularVtuberCursorMu.Lock()
	defer s.popularVtuberCursorMu.Unlock()

	for i := s.popularVtuberCursor; i < len(vtuberPool); i++ {
		if vtuberPool[i].Type == UserType_Popular {
			s.popularVtuberCursor = i
			return vtuberPool[i], nil
		}
	}

	for i := 0; i < len(vtuberPool); i++ {
		if vtuberPool[i].Type == UserType_Popular {
			s.popularVtuberCursor = i
			return vtuberPool[i], nil
		}
	}

	return nil, fmt.Errorf("人気VTuber候補を発見できませんでした")
}

// 普通の配信者でいいなら、Normalなものを探せばいい
func (s *userScheduler) SelectVTuber() *User {
	s.vtuberCursorMu.Lock()
	defer s.vtuberCursorMu.Unlock()

	vtuber := vtuberPool[s.vtuberCursor]
	s.vtuberCursor = (s.vtuberCursor + 1) % len(vtuberPool)
	return vtuber
}

// viewerは、可能な限り何もしてない人から払い出していく
func (s *userScheduler) SelectViewer() *User {
	s.viewerCursorMu.Lock()
	defer s.viewerCursorMu.Unlock()

	viewer := viewerPool[s.viewerCursor]
	s.viewerCursor = (s.viewerCursor + 1) % len(viewerPool)
	return viewer
}

// 予約時のコラボ配信者候補を出す
// FIXME: なるべく重くしたいので、人気配信者や、投稿数が多い配信者を狙う
func (s *userScheduler) SelectCollaborators(n int) []*User {
	//
	if n >= len(vtuberPool) {
		n = len(vtuberPool) - 1
	}
	return vtuberPool[:n]
}
