package scheduler

import (
	"errors"
	"fmt"
	"sync"

	"github.com/biogo/store/interval"
	"github.com/isucon/isucon13/bench/internal/config"
	"go.uber.org/zap"
)

var ErrNoReservation = errors.New("条件を満たす予約がみつかりませんでした")

var (
	ReservationSched = mustNewReservationScheduler(config.BaseAt, config.NumSlots, config.NumHours+10)
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
}

func mustNewReservationScheduler(baseAt int64, numSlots int64, hours int) *ReservationScheduler {
	lgr := zap.S()

	intervalTempertures, err := newIntervalTemperture(baseAt, numSlots, hours)
	if err != nil {
		lgr.Warn("failed to initiate interval temperture: %s\n", err.Error())
	}
	return &ReservationScheduler{
		reservationPool:     []*Reservation{},
		intervalTempertures: intervalTempertures,
		intTreeStates:       make(map[int]CommitState),
		intervalTree:        &interval.IntTree{},
	}
}

func (r *ReservationScheduler) loadReservations(reservations []*Reservation) {
	const needFastInsertion = true
	for _, reservation := range reservations {
		r.intervalTree.Insert(reservation, needFastInsertion)
		r.intTreeStates[reservation.id] = CommitState_None
		r.reservationPool = append(r.reservationPool, reservation)
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

	// FIXME: 炎上配信の払い出し数をチェックし、まだ払い出せるなら成立した予約から炎上配信として扱えるようにしていく

	r.reservations = append(r.reservations, reservation)

	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()
	r.intTreeStates[int(reservation.id)] = CommitState_Committed
}

func (r *ReservationScheduler) AbortReservation(reservation *Reservation) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	r.intTreeStates[int(reservation.id)] = CommitState_None
}

func (r *ReservationScheduler) GetHotShortReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	lgr := zap.S()

	intervals, err := r.intervalTempertures.findHotIntervals()
	if err != nil {
		lgr.Warnf("GetHotShortReservation: failed to find hot intervals: %s\n", err.Error())
		return nil, ErrNoReservation
	}

	for i := 0; i < len(intervals); i++ {
		interval := intervals[i]
		founds := r.intervalTree.Get(&Reservation{
			StartAt: interval.startAt.Unix(),
			EndAt:   interval.endAt.Unix(),
		})
		if len(founds) == 0 {
			lgr.Warnf("GetHotShortReservation: failed to get reservation from interval tree (founds=0)")
			continue
		}

		reservations, err := ConvertFromIntInterface(founds)
		if err != nil {
			lgr.Warnf("GetHotShortReservation: failed to convert reservation: %s\n", err.Error())
			return nil, err
		}

		for _, reservation := range reservations {
			id := reservation.id
			if reservation.Hours() >= config.LongHourThreshold {
				continue
			}
			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservation, nil
			}
		}
	}

	lgr.Warn("GetHotShortReservation: failed to get reservation (not found)")
	return nil, ErrNoReservation
}

func (r *ReservationScheduler) GetHotLongReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	lgr := zap.S()

	intervals, err := r.intervalTempertures.findHotIntervals()
	if err != nil {
		lgr.Warnf("GetHotLongReservation: failed to find hot intervals: %s\n", err.Error())
		return nil, ErrNoReservation
	}

	for i := 0; i < len(intervals); i++ {
		interval := intervals[i]
		founds := r.intervalTree.Get(&Reservation{
			StartAt: interval.startAt.Unix(),
			EndAt:   interval.endAt.Unix(),
		})
		if len(founds) == 0 {
			lgr.Warnf("GetHotLongReservation: failed to get reservation from interval tree (founds=0)")
			continue
		}

		reservations, err := ConvertFromIntInterface(founds)
		if err != nil {
			lgr.Warnf("GetHotLongReservation: failed to convert reservation: %s\n", err.Error())
			return nil, err
		}

		for _, reservation := range reservations {
			id := reservation.id
			if reservation.Hours() < config.LongHourThreshold {
				continue
			}
			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservation, nil
			}
		}
	}

	lgr.Warn("GetHotLongReservation: failed to get reservation (not found)")
	return nil, ErrNoReservation
}

func (r *ReservationScheduler) GetColdShortReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	lgr := zap.S()

	intervals, err := r.intervalTempertures.findColdIntervals()
	if err != nil {
		lgr.Warnf("GetColdShortReservation: failed to find hot intervals: %s\n", err.Error())
		return nil, ErrNoReservation
	}

	for i := 0; i < len(intervals); i++ {
		interval := intervals[i]
		founds := r.intervalTree.Get(&Reservation{
			StartAt: interval.startAt.Unix(),
			EndAt:   interval.endAt.Unix(),
		})
		if len(intervals) == 0 {
			lgr.Warnf("GetColdShortReservation: failed to get reservation from interval tree (founds=0)")
			return nil, fmt.Errorf("no cold short reservation")
		}

		reservations, err := ConvertFromIntInterface(founds)
		if err != nil {
			lgr.Warnf("GetColdShortReservation: failed to convert reservation: %s\n", err.Error())
			return nil, err
		}

		for _, reservation := range reservations {
			id := reservation.id
			if reservation.Hours() >= config.LongHourThreshold {
				continue
			}

			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservation, nil
			}
		}
	}

	lgr.Warn("GetColdShortReservation: failed to get reservation (not found)")
	return nil, ErrNoReservation
}

func (r *ReservationScheduler) GetColdLongReservation() (*Reservation, error) {
	r.intTreeMu.Lock()
	defer r.intTreeMu.Unlock()

	lgr := zap.S()

	intervals, err := r.intervalTempertures.findColdIntervals()
	if err != nil {
		lgr.Warnf("GetColdLongReservation: failed to find hot intervals: %s\n", err.Error())
		return nil, ErrNoReservation
	}

	for i := 0; i < len(intervals); i++ {
		interval := intervals[i]
		founds := r.intervalTree.Get(&Reservation{
			StartAt: interval.startAt.Unix(),
			EndAt:   interval.endAt.Unix(),
		})
		if len(intervals) == 0 {
			lgr.Warnf("GetColdLongReservation: failed to get reservation from interval tree (founds=0)")
			return nil, fmt.Errorf("no cold long reservation")
		}

		reservations, err := ConvertFromIntInterface(founds)
		if err != nil {
			lgr.Warnf("GetColdLongReservation: failed to convert reservation: %s\n", err.Error())
			return nil, err
		}

		for _, reservation := range reservations {
			id := reservation.id
			if reservation.Hours() < config.LongHourThreshold {
				continue
			}

			if state, ok := r.intTreeStates[id]; ok && state == CommitState_None {
				r.intTreeStates[id] = CommitState_Inflight
				return reservation, nil
			}
		}
	}

	lgr.Warn("GetColdLongReservation: failed to get reservation (not found)")
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
