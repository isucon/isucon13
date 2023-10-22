package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/isucon/isucandar/pubsub"
)

func init() {
	// 人気配信者が走行ごと変わるようにシャッフルする
	rand.Shuffle(len(vtuberPool), func(i, j int) { vtuberPool[i], vtuberPool[j] = vtuberPool[j], vtuberPool[i] })
}

var UserScheduler = mustNewUserScheduler()

type User struct {
	UserId         int
	Name           string
	DisplayName    string
	Description    string
	RawPassword    string
	HashedPassword string
}

type userScheduler struct {
	loginPopularStreamerPubSub *pubsub.PubSub
	popularStreamerPoolMu      sync.Mutex
	popularStreamerPoolIdx     int
	popularStreamerPool        []*User

	loginStreamerPubSub *pubsub.PubSub
	streamerPoolMu      sync.Mutex
	streamerPoolIdx     int
	streamerPool        []*User

	loginViewerPubSub *pubsub.PubSub
	viewerPoolMu      sync.Mutex
	viewerPoolIdx     int
}

// 人気配信者制限
const popularLimit = 50

func mustNewUserScheduler() *userScheduler {
	sched := new(userScheduler)
	sched.popularStreamerPool = vtuberPool[:popularLimit]
	sched.streamerPool = vtuberPool[popularLimit:]

	sched.loginPopularStreamerPubSub = pubsub.NewPubSub()
	sched.loginPopularStreamerPubSub.Capacity = 1000 // no block
	sched.loginStreamerPubSub = pubsub.NewPubSub()
	sched.loginStreamerPubSub.Capacity = 1000 // no block
	sched.loginViewerPubSub = pubsub.NewPubSub()
	sched.loginViewerPubSub.Capacity = 1000 // no block

	return sched
}

func (s *userScheduler) IsPopularStreamer(userId int) bool {
	for _, streamer := range s.popularStreamerPool {
		if streamer.UserId == userId {
			return true
		}
	}
	return false
}

// 人気配信者
// PreparePopularStreamer は、プールから未ログイン状態の人気配信者を取得します
func (s *userScheduler) PreparePopularStreamer() (*User, error) {
	s.popularStreamerPoolMu.Lock()
	defer s.popularStreamerPoolMu.Unlock()

	if s.popularStreamerPoolIdx >= len(s.popularStreamerPool) {
		return nil, fmt.Errorf("there are no popular streamer in pool")
	}

	u := s.popularStreamerPool[s.popularStreamerPoolIdx]
	s.popularStreamerPoolIdx++

	return u, nil
}

// PublishStreamer は、ログイン済み配信者をpublishします
func (s *userScheduler) PublishPopularStreamer(u *User)   { s.loginPopularStreamerPubSub.Publish(u) }
func (s *userScheduler) RePublishPopularStreamer(u *User) { s.loginPopularStreamerPubSub.Publish(u) }

// SubscribeStreamer は、ログイン済み配信者をsubscribeします
func (s *userScheduler) SubscribePopularStreamer(ctx context.Context) (u *User) {
	s.loginPopularStreamerPubSub.Subscribe(ctx, func(v interface{}) {
		u = v.(*User)
	})
	return
}

// 通常配信者
// PrepareStreamer は、プールから未ログイン状態の配信者を取得します
func (s *userScheduler) PrepareStreamer() (*User, error) {
	s.streamerPoolMu.Lock()
	defer s.streamerPoolMu.Unlock()

	if s.streamerPoolIdx >= len(s.streamerPool) {
		return nil, fmt.Errorf("there are no popular streamer in pool")
	}

	u := s.streamerPool[s.streamerPoolIdx]
	s.streamerPoolIdx++

	return u, nil
}

// PublishStreamer は、ログイン済み配信者をpublishします
func (s *userScheduler) PublishStreamer(u *User)   { s.loginStreamerPubSub.Publish(u) }
func (s *userScheduler) RePublishStreamer(u *User) { s.loginStreamerPubSub.Publish(u) }

// SubscribeStreamer は、ログイン済み配信者をsubscribeします
func (s *userScheduler) SubscribeStreamer(ctx context.Context) (u *User) {
	s.loginStreamerPubSub.Subscribe(ctx, func(v interface{}) {
		u = v.(*User)
	})
	return
}

// 視聴者 (様々な動きをする視聴者を用意するが、どれも同じ視聴者として扱えるようにする)

func (s *userScheduler) PrepareViewer() (*User, error) {
	s.viewerPoolMu.Lock()
	defer s.viewerPoolMu.Unlock()

	if s.viewerPoolIdx >= len(viewerPool) {
		return nil, fmt.Errorf("there are no popular streamer in pool")
	}

	u := viewerPool[s.viewerPoolIdx]
	s.viewerPoolIdx++

	return u, nil
}

// PublishStreamer は、ログイン済み配信者をpublishします
func (s *userScheduler) PublishViewer(u *User)   { s.loginViewerPubSub.Publish(u) }
func (s *userScheduler) RePublishViewer(u *User) { s.loginViewerPubSub.Publish(u) }

// SubscribeStreamer は、ログイン済み配信者をsubscribeします
func (s *userScheduler) SubscribeViewer(ctx context.Context) (u *User) {
	s.loginViewerPubSub.Subscribe(ctx, func(v interface{}) {
		u = v.(*User)
	})
	return
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
func (s *userScheduler) SelectCollaborators(n int) []*User {
	//
	if n >= len(vtuberPool) {
		n = len(vtuberPool) - 1
	}
	return vtuberPool[:n]
}
