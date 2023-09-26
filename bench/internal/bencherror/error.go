package bencherror

import (
	"context"
	"fmt"
	"sync"

	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
	"github.com/isucon/isucon13/bench/internal/benchscore"
)

var (
	SystemError               failure.StringCode = "system"
	InitializeError           failure.StringCode = "initialize"
	PreTestError              failure.StringCode = "pretest"
	BenchmarkCriticalError    failure.StringCode = "benchmark-critical"
	BenchmarkApplicationError failure.StringCode = "benchmark-application"
	BenchmarkTimeoutError     failure.StringCode = "benchmark-timeout"
	BenchmarkTemporaryError   failure.StringCode = "benchmark-temporary"
	FinalCheckError           failure.StringCode = "finalcheck"

	DBInconsistencyError failure.StringCode = "db-inconsistency"
)

var (
	benchErrors *failure.Errors
	doneOnce    sync.Once
)

func InitializeErrors(ctx context.Context) {
	benchErrors = failure.NewErrors(ctx)
}

// FIXME: もうちょっと細分化して、エラーの一貫性を持たせたい
func WrapError(code failure.StringCode, err error) error {
	e := failure.NewError(code, err)
	benchErrors.Add(e)
	benchscore.AddPenalty(failureCodeToScoreTag(code))
	return fmt.Errorf("%s: %w", err.Error(), e)
}

func GetFinalErrorMessages() map[string][]string {
	doneOnce.Do(func() {
		benchErrors.Done()
	})
	return benchErrors.Messages()
}

func GetFinalPenalties() map[string]int64 {
	doneOnce.Do(func() {
		benchErrors.Done()
	})
	return benchErrors.Count()
}

func failureCodeToScoreTag(code failure.StringCode) score.ScoreTag {
	switch code {
	case SystemError:
		return benchscore.SystemError
	case BenchmarkApplicationError:
		return benchscore.BenchmarkApplicationError
	case InitializeError:
		return benchscore.InitializeError
	case PreTestError:
		return benchscore.PreTestError
	case BenchmarkCriticalError:
		return benchscore.BenchmarkCriticalError
	case BenchmarkTimeoutError:
		return benchscore.BenchmarkTimeoutError
	case BenchmarkTemporaryError:
		return benchscore.BenchmarkTemporaryError
	case FinalCheckError:
		return benchscore.FinalCheckError
	case DBInconsistencyError:
		return benchscore.DBInconsistencyError
	default:
		panic("unreachable")
	}
}
