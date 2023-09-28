package bencherror

import (
	"context"

	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
)

var (
	penaltyScore *score.Score

	// ペナルティ
	BenchmarkApplicationPenalty score.ScoreTag = "benchmark-application-error"
	BenchmarkTimeoutPenalty     score.ScoreTag = "benchmark-timeout-error"
)

func InitPenalty(ctx context.Context) {
	penaltyScore = score.NewScore(ctx)

	penaltyScore.Set(BenchmarkApplicationPenalty, 1)
	penaltyScore.Set(BenchmarkTimeoutPenalty, 1)
}

func AddPenalty(code failure.StringCode) {
	switch code {
	case BenchmarkApplicationError:
		penaltyScore.Add(BenchmarkApplicationPenalty)
	case BenchmarkTimeoutError:
		penaltyScore.Add(BenchmarkTimeoutPenalty)
	}
}

func GetCurrentPenalty() int64 {
	return penaltyScore.Sum()
}

func GetFinalPenalty() int64 {
	penaltyScore.Done()
	return penaltyScore.Sum()
}
