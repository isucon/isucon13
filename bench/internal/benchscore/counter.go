package benchscore

import (
	"context"
	"sync"

	"github.com/isucon/isucandar/score"
)

const (
	DNSResolve score.ScoreTag = "dns-resolve"
	DNSFailed  score.ScoreTag = "dns-failed"

	TooSlow     score.ScoreTag = "too-slow-left"
	TooManySpam score.ScoreTag = "too-many-spam"
)

var (
	counter         *score.Score
	doneCounterOnce sync.Once
)

func InitCounter(ctx context.Context) {
	counter = score.NewScore(ctx)
	counter.Set(DNSResolve, 1)
	counter.Set(DNSFailed, 1)
	counter.Set(TooSlow, 1)
	counter.Set(TooManySpam, 1)
}

func IncResolves() {
	counter.Add(DNSResolve)
}

func NumResolves() int64 {
	table := counter.Breakdown()
	return table[DNSResolve]
}

func IncDNSFailed() {
	counter.Add(DNSFailed)
}

func NumDNSFailed() int64 {
	table := counter.Breakdown()
	return table[DNSFailed]
}

func GetByTag(tag score.ScoreTag) int64 {
	return counter.Breakdown()[tag]
}

func DoneCounter() {
	doneCounterOnce.Do(func() {
		counter.Close()
	})
}
