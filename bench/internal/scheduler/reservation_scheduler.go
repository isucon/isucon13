package scheduler

import (
	"sync"
	"time"
)

var reservationScheduler = newReservationScheduler()

type ReservationTimeSlice struct {
	Id      int
	UserId  int
	StartAt time.Time
	EndAt   time.Time
}

type ReservationScheduler struct {
	mu sync.RWMutex
	// reservations []*Reservation
}

func newReservationScheduler() *ReservationScheduler {
	return nil
}

/*
func (r *ReservationScheduler) AddReservation(reservation *Reservation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// r.reservations = append(r.reservations, reservation)
}
*/

func (r *ReservationScheduler) GetHotLongReservation() {

}

func (r *ReservationScheduler) GetHotShortReservation() {

}

func (r *ReservationScheduler) GetColdReservation() {

}

// FIXME: 予約の突合処理
/*
func (r *ReservationScheduler) RangeReserved(fn func(*Reservation)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, reservationInterval := range r.reservations {
		fn(reservationInterval)
	}
}
*/
