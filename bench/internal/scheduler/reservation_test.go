package scheduler

import (
	"testing"
	"time"

	"github.com/biogo/store/interval"
	"github.com/stretchr/testify/assert"
)

func TestReservation_Overlap(t *testing.T) {
	var (
		baseUnix int64 = 1711897200
		baseAt         = time.Unix(baseUnix, 0)
	)
	tests := []struct {
		name          string
		reservation   *Reservation
		interval      interval.IntRange
		wantIsOverlap bool
	}{
		{
			name:          "intervalが覆ってる",
			reservation:   &Reservation{id: 1, StartAt: baseAt.Add(1 * time.Hour).Unix(), EndAt: baseAt.Add(2 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseUnix), End: int(baseAt.Add(3 * time.Hour).Unix())},
			wantIsOverlap: true,
		},
		{
			name:          "予約が覆ってる",
			reservation:   &Reservation{id: 1, StartAt: baseUnix, EndAt: baseAt.Add(3 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseAt.Add(1 * time.Hour).Unix()), End: int(baseAt.Add(2 * time.Hour).Unix())},
			wantIsOverlap: true,
		},
		{
			name:          "intervalと予約区間が同一",
			reservation:   &Reservation{id: 1, StartAt: baseUnix, EndAt: baseAt.Add(1 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseUnix), End: int(baseAt.Add(1 * time.Hour).Unix())},
			wantIsOverlap: true,
		},
		{
			name:          "intervalと予約区間が同一かつサイズが1",
			reservation:   &Reservation{id: 1, StartAt: baseUnix, EndAt: baseUnix},
			interval:      interval.IntRange{Start: int(baseUnix), End: int(baseUnix)},
			wantIsOverlap: true,
		},
		{
			name:          "予約が左隣にある",
			reservation:   &Reservation{id: 1, StartAt: baseUnix, EndAt: baseAt.Add(1 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseAt.Add(1 * time.Hour).Unix()), End: int(baseAt.Add(3 * time.Hour).Unix())},
			wantIsOverlap: false,
		},
		{
			name:          "予約が左にある",
			reservation:   &Reservation{id: 1, StartAt: baseUnix, EndAt: baseAt.Add(1 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseAt.Add(2 * time.Hour).Unix()), End: int(baseAt.Add(3 * time.Hour).Unix())},
			wantIsOverlap: false,
		},
		{
			name:          "予約が右隣にある",
			reservation:   &Reservation{id: 1, StartAt: baseAt.Add(3 * time.Hour).Unix(), EndAt: baseAt.Add(5 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseUnix), End: int(baseAt.Add(3 * time.Hour).Unix())},
			wantIsOverlap: false,
		},
		{
			name:          "予約が右にある",
			reservation:   &Reservation{id: 1, StartAt: baseAt.Add(4 * time.Hour).Unix(), EndAt: baseAt.Add(5 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseUnix), End: int(baseAt.Add(3 * time.Hour).Unix())},
			wantIsOverlap: false,
		},

		{
			name:          "same start-end interval & left-wide reservation",
			reservation:   &Reservation{id: 1, StartAt: baseUnix, EndAt: baseAt.Add(1 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseAt.Add(1 * time.Hour).Unix()), End: int(baseAt.Add(1 * time.Hour).Unix())},
			wantIsOverlap: true,
		},
		{
			name:          "same start-end interval & right-wide reservation",
			reservation:   &Reservation{id: 1, StartAt: baseAt.Add(1 * time.Hour).Unix(), EndAt: baseAt.Add(2 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseAt.Add(1 * time.Hour).Unix()), End: int(baseAt.Add(1 * time.Hour).Unix())},
			wantIsOverlap: true,
		},
		{
			name:          "same start-end interval & wide reservation",
			reservation:   &Reservation{id: 1, StartAt: baseUnix, EndAt: baseAt.Add(2 * time.Hour).Unix()},
			interval:      interval.IntRange{Start: int(baseAt.Add(1 * time.Hour).Unix()), End: int(baseAt.Add(1 * time.Hour).Unix())},
			wantIsOverlap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.reservation.Overlap(tt.interval)
			if tt.wantIsOverlap {
				assert.True(t, got)
			} else {
				assert.False(t, got)
			}
		})
	}
}
