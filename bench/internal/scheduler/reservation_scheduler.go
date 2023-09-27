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
	mu                    sync.RWMutex
	reservationTimeSlices []*ReservationTimeSlice
}

func newReservationScheduler() *ReservationScheduler {
	return nil
}

// FIXME: シード生成したシーズン毎のスケジュールをもとにしてデータ生成を行う

func (r *ReservationScheduler) AddReservation(timeSlice *ReservationTimeSlice) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reservationTimeSlices = append(r.reservationTimeSlices, timeSlice)
}
