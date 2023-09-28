package scheduler

import (
	"time"
)

type LivestreamPattern struct {
	UserID      int
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
}

func newLivestreamPattern(userID int, title string, description string, startAtS string, endAtS string) LivestreamPattern {
	startAt, _ := time.Parse("2006-01-02 15:04:05", startAtS)
	endAt, _ := time.Parse("2006-01-02 15:04:05", endAtS)
	return LivestreamPattern{
		UserID:      userID,
		Title:       title,
		Description: description,
		StartAt:     startAt,
		EndAt:       endAt,
	}
}
