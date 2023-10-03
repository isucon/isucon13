package scheduler

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReservationScheduler_Edgecase(t *testing.T) {
	var (
		baseUnix int64 = 1711897200
		baseAt         = time.Unix(baseUnix, 0)
		hours          = 24
	)

	sched := mustNewReservationScheduler(baseUnix, hours)
	log.Println("load")
	sched.loadReservations([]*Reservation{
		{Id: 1, StartAt: baseAt.Add(0 * time.Hour).Unix(), EndAt: baseAt.Add(1 * time.Hour).Unix()},
		{Id: 2, StartAt: baseAt.Add(3 * time.Hour).Unix(), EndAt: baseAt.Add(4 * time.Hour).Unix()},
		{Id: 3, StartAt: baseAt.Add(6 * time.Hour).Unix(), EndAt: baseAt.Add(10 * time.Hour).Unix()},
		{Id: 4, StartAt: baseAt.Add(10 * time.Hour).Unix(), EndAt: baseAt.Add(15 * time.Hour).Unix()},
		{Id: 5, StartAt: baseAt.Add(15 * time.Hour).Unix(), EndAt: baseAt.Add(20 * time.Hour).Unix()},
		{Id: 6, StartAt: baseAt.Add(21 * time.Hour).Unix(), EndAt: baseAt.Add(23 * time.Hour).Unix()},
	})
	log.Println("===== test1 =====")
	reservation, err := sched.GetHotShortReservation()
	assert.NoError(t, err)
	assert.Equal(t, 1, reservation.Id)
	assert.Equal(t, baseAt.Add(0*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(1*time.Hour).Unix(), reservation.EndAt)

	sched.CommitReservation(reservation)

	log.Println("===== test2 =====")
	reservation, err = sched.GetHotShortReservation()
	assert.NoError(t, err)
	assert.Equal(t, 2, reservation.Id)
	assert.Equal(t, baseAt.Add(3*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(4*time.Hour).Unix(), reservation.EndAt)

	sched.CommitReservation(reservation)

	log.Println("===== test3 =====")
	reservation, err = sched.GetHotLongReservation()
	log.Printf("[Test] id=%d, [%s,%s)\n", reservation.Id, time.Unix(reservation.StartAt, 0), time.Unix(reservation.EndAt, 0))
	assert.NoError(t, err)
	assert.Equal(t, 6, reservation.Id)
	assert.Equal(t, baseAt.Add(21*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(23*time.Hour).Unix(), reservation.EndAt)

	sched.CommitReservation(reservation)

	log.Println("===== test4 ====")
	reservation, err = sched.GetHotShortReservation()
	assert.NoError(t, err)
	assert.Equal(t, 3, reservation.Id)
	assert.Equal(t, baseAt.Add(6*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(10*time.Hour).Unix(), reservation.EndAt)

	sched.CommitReservation(reservation)

	reservation, err = sched.GetHotLongReservation()
	assert.NoError(t, err)
	assert.Equal(t, 5, reservation.Id)
	assert.Equal(t, baseAt.Add(15*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(20*time.Hour).Unix(), reservation.EndAt)

	sched.CommitReservation(reservation)

	reservation, err = sched.GetHotShortReservation()
	assert.NoError(t, err)
	assert.Equal(t, 4, reservation.Id)
	assert.Equal(t, baseAt.Add(10*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(15*time.Hour).Unix(), reservation.EndAt)

	sched.CommitReservation(reservation)

	reservation, err = sched.GetHotLongReservation()
	assert.Error(t, err)

}

// AbortReservationした際のテスト
// commitした後、再利用できないか
// abortした後、再利用できるか
func TestReservationScheduler_Abort(t *testing.T) {

}

func TestReservationScheduler_Cold(t *testing.T) {
	// 枠数2以上のケースをテスト
}

// membenchを実行して、リソース消費について簡単に見ておく
