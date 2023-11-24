package scheduler

import (
	"fmt"
	"math/rand"
)

var UserScheduler = mustNewUserScheduler()

type User struct {
	Name           string
	DisplayName    string
	Description    string
	RawPassword    string
	HashedPassword string
	DarkMode       bool
}

type userScheduler struct {
	streamerPool []*User
}

func mustNewUserScheduler() *userScheduler {
	sched := new(userScheduler)
	sched.streamerPool = streamerPool

	return sched
}

// 通常配信者
func (s *userScheduler) RangeStreamer(fn func(streamer *User)) {
	for _, streamer := range s.streamerPool {
		fn(streamer)
	}
}

// テスト用
func (s *userScheduler) GetRandomStreamer() *User {
	idx := rand.Intn(len(s.streamerPool))
	return s.streamerPool[idx]
}

// 視聴者 (様々な動きをする視聴者を用意するが、どれも同じ視聴者として扱えるようにする)
func (s *userScheduler) RangeViewer(fn func(viewer *User)) {
	for _, viewer := range viewerPool {
		fn(viewer)
	}
}

func (s *userScheduler) GetInitialUserForPretest(id int64) (*User, error) {
	idx := max(id-1, 1)
	if idx > int64(len(initialUserPool)-1) {
		return nil, fmt.Errorf("想定しない初期ユーザが利用されました (idx=%d)", idx)
	}

	return initialUserPool[idx], nil
}
