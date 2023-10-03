package scheduler

import (
	"fmt"
	"log"
	"sync"

	"github.com/biogo/store/interval"
)

// NOTE: phase1は初期状態ですべての予約が埋まっている
// FIXME: 各所でシーズン判定ができないと、選べなさそう。使う側はシーズンがわかっているので取り出せるが、clientから予約のたびにココに突っ込むとなると、どれに突っ込めばいいかもたないといけない
// フェーズ切り替えの際、
// FIXME: baseAtちゃんとした値に直す
var (
	Phase2ReservationScheduler = mustNewReservationScheduler(1719759600, 24*365/4+10)
	Phase3ReservationScheduler = mustNewReservationScheduler(1719759600, 24*365/4+10)
	Phase4ReservationScheduler = mustNewReservationScheduler(1719759600, 24*365/4+10)
)

func init() {
	// スケジューラに初期データをロード
	Phase2ReservationScheduler.loadReservations(phase2ReservationPool)
}

// FIXME: 予約のタイトルと説明文はベンチ走行中にランダムに割り当てる
// 一定のランダム要素がないと、ベンチ走行データを再利用してチートするなど考えられるため

type ReservationScheduler struct {
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
}

func mustNewReservationScheduler(baseAt int64, hours int) *ReservationScheduler {
	// 同時配信枠数
	// FIXME: 最低限、枠数１、枠数２の場合はテストを十分に行っておきたい
	const numSlots = 1
	intervalTempertures, err := newIntervalTemperture(baseAt, numSlots, hours)
	if err != nil {
		log.Fatalln(err)
	}
	return &ReservationScheduler{
		intervalTempertures: intervalTempertures,
		intTreeStates:       make(map[int]CommitState),
		intervalTree:        &interval.IntTree{},
	}
}

func (r *ReservationScheduler) loadReservations(reservations []*Reservation) {
	const needFastInsertion = true
	for _, reservation := range reservations {
		r.intervalTree.Insert(reservation, needFastInsertion)
		r.intTreeStates[reservation.Id] = CommitState_None
	}
	// Get, DoMatching*が呼び出される前に必ずRangesで調整しておく
	// NOTE: 以後、区間木に対する挿入は行われないので逐次呼び出す必要はない
	// AdjustRangeはLLRBノードの範囲(Range)を更新する関数. AdjustRangesはツリーを再帰的にこれを各ノードについて実施していく.
	r.intervalTree.AdjustRanges()
}

// CommitReservation は、予約追加リクエストが通ったことをintervalTemperturesに記録します
func (r *ReservationScheduler) CommitReservation(reservation *Reservation) {
	r.reservationsMu.Lock()
	defer r.reservationsMu.Unlock()

	r.intervalTempertures.addInterval(
		reservation.StartAt,
		reservation.EndAt,
	)

	r.reservations = append(r.reservations, reservation)

	r.intTreeStates[int(reservation.Id)] = CommitState_Committed
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
		return nil, err
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
			continue
		}

		for i := len(reservations) - 1; i >= 0; i-- {
			id := reservations[i].Id
			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservations[i], nil
			}
		}
	}

	return nil, fmt.Errorf("no hot long reservation")
}

func (r *ReservationScheduler) GetHotShortReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	intervals, err := r.intervalTempertures.findHotIntervals()
	if err != nil {
		return nil, err
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
			continue
		}

		for i := 0; i < len(reservations); i++ {
			id := reservations[i].Id
			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservations[i], nil
			}
		}
	}

	return nil, fmt.Errorf("no hot short reservation")
}

func (r *ReservationScheduler) GetColdReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	intervals, err := r.intervalTempertures.findColdIntervals()
	if err != nil {
		return nil, err
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

	return nil, fmt.Errorf("no hot long reservation")
}

// 予約の突合処理に使う
func (r *ReservationScheduler) RangeReserved(fn func(*Reservation)) {
	r.reservationsMu.Lock()
	defer r.reservationsMu.Unlock()

	for _, reservationInterval := range r.reservations {
		fn(reservationInterval)
	}
}
