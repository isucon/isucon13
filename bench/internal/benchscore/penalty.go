package benchscore

import (
	"context"

	"github.com/isucon/isucandar/score"
)

var (
	penaltyScore *score.Score

	// ペナルティ

	// FIXME: ペナルティの内容を具体的に細分化する
	SystemError               score.ScoreTag = "system-error"
	InitializeError           score.ScoreTag = "intialize-error"
	PreTestError              score.ScoreTag = "pretest-error"
	BenchmarkCriticalError    score.ScoreTag = "benchmark-critical-error"
	BenchmarkApplicationError score.ScoreTag = "benchmark-application-error"
	BenchmarkTimeoutError     score.ScoreTag = "benchmark-timeout-error"
	BenchmarkTemporaryError   score.ScoreTag = "benchmark-temporary-error"
	FinalCheckError           score.ScoreTag = "finalcheck-error"
	DBInconsistencyError      score.ScoreTag = "db-inconsistency-error"
)

func initPenalty(ctx context.Context) {
	penaltyScore = score.NewScore(ctx)

	penaltyScore.Set(SystemError, 1)
	penaltyScore.Set(InitializeError, 1)
	penaltyScore.Set(PreTestError, 1)
	penaltyScore.Set(BenchmarkApplicationError, 1)
	penaltyScore.Set(BenchmarkTemporaryError, 1)
	penaltyScore.Set(FinalCheckError, 1)

	// スコアの計算には関与しないので0
	penaltyScore.Set(BenchmarkCriticalError, 0)
}

func AddPenalty(tag score.ScoreTag) {
	penaltyScore.Add(tag)
}

func GetCurrentPenalty() int64 {
	return penaltyScore.Sum()
}

func GetFinalPenalty() int64 {
	benchScore.Done()
	return benchScore.Sum()
}
