package scheduler

import "sync"

type ReserveTimeSlice struct {
}

type ReservationScheduler struct {
	mu sync.RWMutex
}

func newReservationScheduler() *ReservationScheduler {
	// FIXME: impl
	return nil
}

var reservationScheduler = newReservationScheduler()
