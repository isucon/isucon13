package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/biogo/store/interval"
	"github.com/isucon/isucandar/pubsub"
)

// 同時配信枠数
const numSlots = 2

var ErrNoReservation = errors.New("条件を満たす予約がみつかりませんでした")

var (
	ReservationSched = mustNewReservationScheduler(1711897200, (24 * 365))
)

func init() {
	// スケジューラに初期データをロード
	ReservationSched.loadReservations(reservationPool)
}

// FIXME: 予約のタイトルと説明文はベンチ走行中にランダムに割り当てる
// 一定のランダム要素がないと、ベンチ走行データを再利用してチートするなど考えられるため

type ReservationScheduler struct {
	// 負荷フェーズごとに変わるので、プールを内部に保持
	reservationPool []*Reservation

	// 実施された予約リクエストに基づいて、予約が少ない区間、予約が多い区間を割り出すために温度(予約成功回数)を保持しておく
	// 予約リクエストに基づいて逐次的にカウントアップされる
	intervalTempertures *IntervalTemperatures

	// ある区間とオーバーラップする予約を列挙するために区間木を用いる
	// 起動時に予め決められた予約をすべてロードする
	intTreeMu     sync.Mutex
	intTreeStates map[int]CommitState
	intervalTree  *interval.IntTree

	// 成功した予約を覚えておく
	// 最終的にfinalcheckの突合に使う
	reservationsMu sync.Mutex
	reservations   []*Reservation

	// FIXME: 振ったタグを覚えておく
	// reservationid => []string{"tag1", "tag2", ...} という感じで持てばいいか

	// FIXME: 予約完了してからライブ配信を取り出したいケースがある
	// その場合、今予約完了していて利用可能な予約はあるのか判定しなければならない
	// そこでPubSubを用いて判定するようにする
	reservationPubSub        *pubsub.PubSub
	popularReservationPubSub *pubsub.PubSub
}

func mustNewReservationScheduler(baseAt int64, hours int) *ReservationScheduler {
	intervalTempertures, err := newIntervalTemperture(baseAt, numSlots, hours)
	if err != nil {
		log.Fatalln(err)
	}
	return &ReservationScheduler{
		reservationPool:     []*Reservation{},
		intervalTempertures: intervalTempertures,
		intTreeStates:       make(map[int]CommitState),
		intervalTree:        &interval.IntTree{},
		reservationPubSub:   pubsub.NewPubSub(),
	}
}

func (r *ReservationScheduler) loadReservations(reservations []*Reservation) {
	const needFastInsertion = true
	for _, reservation := range reservations {
		r.intervalTree.Insert(reservation, needFastInsertion)
		r.intTreeStates[reservation.Id] = CommitState_None
		r.reservationPool = append(r.reservationPool, reservation)
	}
	// Get, DoMatching*が呼び出される前に必ずRangesで調整しておく
	// NOTE: 以後、区間木に対する挿入は行われないので逐次呼び出す必要はない
	// AdjustRangeはLLRBノードの範囲(Range)を更新する関数. AdjustRangesはツリーを再帰的にこれを各ノードについて実施していく.
	r.intervalTree.AdjustRanges()
}

func (r *ReservationScheduler) GetPopularLivestream(ctx context.Context) (reservation *Reservation) {
	r.popularReservationPubSub.Subscribe(ctx, func(v interface{}) {
		reservation = v.(*Reservation)
	})
	return
}

func (r *ReservationScheduler) GetLivestream(ctx context.Context) (reservation *Reservation) {
	r.reservationPubSub.Subscribe(ctx, func(v interface{}) {
		reservation = v.(*Reservation)
	})
	return
}

// CommitReservation は、予約追加リクエストが通ったことをintervalTemperturesに記録します
func (r *ReservationScheduler) CommitReservation(reservation *Reservation) {
	r.reservationsMu.Lock()
	defer r.reservationsMu.Unlock()

	r.intervalTempertures.addInterval(
		reservation.StartAt,
		reservation.EndAt,
	)

	// FIXME: 炎上配信の払い出し数をチェックし、まだ払い出せるなら成立した予約から炎上配信として扱えるようにしていく

	r.reservations = append(r.reservations, reservation)

	r.intTreeStates[int(reservation.Id)] = CommitState_Committed

	if UserScheduler.IsPopularStreamer(reservation.UserId) {
		r.popularReservationPubSub.Publish(reservation)
	} else {
		r.reservationPubSub.Publish(reservation)
	}
}

func (r *ReservationScheduler) AbortReservation(reservation *Reservation) {
	r.reservationsMu.Lock()
	defer r.reservationsMu.Unlock()

	r.intTreeStates[int(reservation.Id)] = CommitState_None
}

func (r *ReservationScheduler) GetHotLongReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	intervals, err := r.intervalTempertures.findHotIntervals()
	if err != nil {
		return nil, ErrNoReservation
	}

	for i := len(intervals) - 1; i >= 0; i-- {
		interval := intervals[i]
		founds := r.intervalTree.Get(&Reservation{
			StartAt: interval.startAt.Unix(),
			EndAt:   interval.endAt.Unix(),
		})
		if len(founds) == 0 {
			continue
		}

		reservations, err := ConvertFromIntInterface(founds)
		if err != nil {
			return nil, err
		}

		for i := len(reservations) - 1; i >= 0; i-- {
			id := reservations[i].Id
			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservations[i], nil
			}
		}
	}

	return nil, ErrNoReservation
}

func (r *ReservationScheduler) GetHotShortReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	intervals, err := r.intervalTempertures.findHotIntervals()
	if err != nil {
		return nil, ErrNoReservation
	}

	for i := 0; i < len(intervals); i++ {
		interval := intervals[i]
		founds := r.intervalTree.Get(&Reservation{
			StartAt: interval.startAt.Unix(),
			EndAt:   interval.endAt.Unix(),
		})
		if len(founds) == 0 {
			continue
		}

		reservations, err := ConvertFromIntInterface(founds)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(reservations); i++ {
			id := reservations[i].Id
			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservations[i], nil
			}
		}
	}

	return nil, ErrNoReservation
}

func (r *ReservationScheduler) GetColdReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	intervals, err := r.intervalTempertures.findColdIntervals()
	if err != nil {
		return nil, ErrNoReservation
	}

	for i := 0; i < len(intervals); i++ {
		interval := intervals[i]
		founds := r.intervalTree.Get(&Reservation{
			StartAt: interval.startAt.Unix(),
			EndAt:   interval.endAt.Unix(),
		})
		if len(intervals) == 0 {
			return nil, fmt.Errorf("no hot long reservation")
		}

		reservations, err := ConvertFromIntInterface(founds)
		if err != nil {
			return nil, err
		}

		for _, reservation := range reservations {
			id := reservation.Id
			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservation, nil
			}
		}
	}

	return nil, ErrNoReservation
}

// 予約の突合処理に使う
func (r *ReservationScheduler) RangeReserved(fn func(*Reservation)) {
	r.reservationsMu.Lock()
	defer r.reservationsMu.Unlock()

	for _, reservationInterval := range r.reservations {
		fn(reservationInterval)
	}
}
