package scheduler

import (
	"math/rand"
)

func init() {
	// 人気配信者が走行ごと変わるようにシャッフルする
	rand.Shuffle(len(streamerPool), func(i, j int) { streamerPool[i], streamerPool[j] = streamerPool[j], streamerPool[i] })
}

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
	popularStreamerPool []*User
	streamerPool        []*User
}

// 人気配信者制限
const popularLimit = 50

func mustNewUserScheduler() *userScheduler {
	sched := new(userScheduler)
	sched.popularStreamerPool = streamerPool[:popularLimit]
	sched.streamerPool = streamerPool[popularLimit:]

	return sched
}

func (s *userScheduler) IsPopularStreamer(name string) bool {
	for _, streamer := range s.popularStreamerPool {
		if streamer.Name == name {
			return true
		}
	}
	return false
}

// 人気配信者
func (s *userScheduler) RangePopularStreamer(fn func(streamer *User)) {
	for _, streamer := range s.popularStreamerPool {
		fn(streamer)
	}
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

// FIXME: 予約スケジューラとの連携
// ログイン不要
// 配信者から取り出す
// FIXME: ただ、予約時に衝突するコラボ配信者を意図的に取り出したい場合がある (異常時シナリオ)
//
//	時間枠を指定したとき、完了した予約から予約者を割り出す必要がある
//
// 予約スケジューラに、当該時刻の予約済みライブ配信を列挙せよと命令し、予約者を割り出す必要がある.
// 予約構造にユーザIDは含まれるので、それをもとにユーザを割り出すことが可能.
// FIXME: 完了した予約を、ユーザ情報と合わせて保持する区間木を新たに定義し、そこから取り出す
// ベンチマーカーだけで完結するので、取り急ぎ保留
func (s *userScheduler) SelectCollaborators(n int) []*User {
	//
	if n >= len(streamerPool) {
		n = len(streamerPool) - 1
	}
	return streamerPool[:n]
}
