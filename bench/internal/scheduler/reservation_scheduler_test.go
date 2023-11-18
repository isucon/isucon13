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
		hours          = 30
	)

	// 基本的にColdShortを取る
	// short < 10hour
	// ColdShort ... いくつか予約を取る。時間が異なる枠もちゃんと用意して、全て残り枠数1まで取れるのをチェックする必要がある
	// ColdShort用にとりあえず3つx2枠の予約を取ってく.
	// 0 ~ 1
	// 1 ~ 5
	// 5 ~ 10
	// でいいか

	// ColdLong ... Shortがまだ残ってる中で、長い物をちゃんと取れるのを確認する必要がある
	// 10 ~ 20
	// 20 ~ 35
	// 35 ~ 59

	// Short -> Long -> Short -> Long -> ...と交互に取っていく

	// Hot ... 初期状態で取れないことを確認。ColdShort/ColdLongが終わったら、Hotが取れることを確認する必要がある
	// 長時間配信者を意識するなら、Hotの方も長時間配信を区別して取るべきか？でないと, 現状の実装ではチップが取れなくなる
	// とりあえず今の実装でやり切る
	// それでとりあえずColdの方は片付くので. ColdとHotで長い短いの区別の処理は全く同じになる.
	// なので、一旦HotShortで扱うようにテストを書く

	sched := mustNewReservationScheduler(baseUnix, 2, hours)
	sched.loadReservations([]*Reservation{
		// short
		{id: 1, StartAt: baseAt.Add(0 * time.Hour).Unix(), EndAt: baseAt.Add(1 * time.Hour).Unix()},
		{id: 2, StartAt: baseAt.Add(0 * time.Hour).Unix(), EndAt: baseAt.Add(1 * time.Hour).Unix()},
		{id: 3, StartAt: baseAt.Add(1 * time.Hour).Unix(), EndAt: baseAt.Add(5 * time.Hour).Unix()},
		{id: 4, StartAt: baseAt.Add(1 * time.Hour).Unix(), EndAt: baseAt.Add(5 * time.Hour).Unix()},
		// long
		{id: 5, StartAt: baseAt.Add(5 * time.Hour).Unix(), EndAt: baseAt.Add(15 * time.Hour).Unix()},
		{id: 6, StartAt: baseAt.Add(5 * time.Hour).Unix(), EndAt: baseAt.Add(15 * time.Hour).Unix()},
		{id: 7, StartAt: baseAt.Add(15 * time.Hour).Unix(), EndAt: baseAt.Add(30 * time.Hour).Unix()},
		{id: 8, StartAt: baseAt.Add(15 * time.Hour).Unix(), EndAt: baseAt.Add(30 * time.Hour).Unix()},
	})

	reservation, err := sched.GetHotShortReservation()
	assert.Error(t, err)
	assert.Nil(t, reservation)
	reservation, err = sched.GetHotLongReservation()
	assert.Error(t, err)
	assert.Nil(t, reservation)

	log.Println("===== cold short1 =====")
	reservation, err = sched.GetColdShortReservation()
	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.Equal(t, 1, reservation.id)
	assert.Equal(t, baseAt.Add(0*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(1*time.Hour).Unix(), reservation.EndAt)
	sched.CommitReservation(reservation)

	log.Println("===== cold long1 =====")
	reservation, err = sched.GetColdLongReservation()
	assert.NoError(t, err)
	assert.Equal(t, 5, reservation.id)
	assert.Equal(t, baseAt.Add(5*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(15*time.Hour).Unix(), reservation.EndAt)
	sched.CommitReservation(reservation)

	log.Println("===== cold short2 =====")
	reservation, err = sched.GetColdShortReservation()
	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.Equal(t, 3, reservation.id)
	assert.Equal(t, baseAt.Add(1*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(5*time.Hour).Unix(), reservation.EndAt)
	sched.CommitReservation(reservation)

	log.Println("===== cold long2 =====")
	reservation, err = sched.GetColdLongReservation()
	assert.NoError(t, err)
	assert.Equal(t, 7, reservation.id)
	assert.Equal(t, baseAt.Add(15*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(30*time.Hour).Unix(), reservation.EndAt)
	sched.CommitReservation(reservation)

	log.Println("===== hot short1 =====")
	reservation, err = sched.GetHotShortReservation()
	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.Equal(t, 2, reservation.id)
	assert.Equal(t, baseAt.Add(0*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(1*time.Hour).Unix(), reservation.EndAt)
	assert.NotNil(t, reservation)
	sched.CommitReservation(reservation)

	log.Println("===== hot long1 =====")
	reservation, err = sched.GetHotLongReservation()
	assert.NoError(t, err)
	assert.Equal(t, 6, reservation.id)
	assert.Equal(t, baseAt.Add(5*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(15*time.Hour).Unix(), reservation.EndAt)
	sched.CommitReservation(reservation)

	log.Println("===== hot short2 =====")
	reservation, err = sched.GetHotShortReservation()
	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.Equal(t, 4, reservation.id)
	assert.Equal(t, baseAt.Add(1*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(5*time.Hour).Unix(), reservation.EndAt)
	assert.NotNil(t, reservation)
	sched.CommitReservation(reservation)

	log.Println("===== hot long2 =====")
	reservation, err = sched.GetHotLongReservation()
	assert.NoError(t, err)
	assert.Equal(t, 8, reservation.id)
	assert.Equal(t, baseAt.Add(15*time.Hour).Unix(), reservation.StartAt)
	assert.Equal(t, baseAt.Add(30*time.Hour).Unix(), reservation.EndAt)
	sched.CommitReservation(reservation)
}

// AbortReservationした際のテスト
// commitした後、再利用できないか
// abortした後、再利用できるか
func TestReservationScheduler_Abort(t *testing.T) {

}

// membenchを実行して、リソース消費について簡単に見ておく
