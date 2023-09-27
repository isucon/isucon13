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
	BenchmarkApplicationError failure.StringCode = "benchmark-application"
	// BenchmarkCriticalError は、ベンチマークを行う前提条件を満たしていない場合を表す
	// この場合、競技者サーバは失格扱いとし、得点を0にする
	BenchmarkCriticalError  failure.StringCode = "benchmark-critical"
	BenchmarkTimeoutError   failure.StringCode = "benchmark-timeout"
	BenchmarkTemporaryError failure.StringCode = "benchmark-temporary"
	FinalCheckError         failure.StringCode = "finalcheck"

	// InternalError はベンチマーカ内部のエラーが含まれる
	// このエラーが出力された場合、運営に連絡してもらう必要がある
	InternalError                 failure.StringCode = "internal"
	InvalidResponseFormatError    failure.StringCode = "invalid-response-format"
	DBInconsistencyError          failure.StringCode = "db-inconsistency"
	UnexpectedHTTPStatusCodeError failure.StringCode = "unexpected-https-status-code"
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
	return fmt.Errorf("[%s]: %s", code, err.Error())
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
	case InternalError:
		return benchscore.InternalError
	case DBInconsistencyError:
		return benchscore.DBInconsistencyError
	case InvalidResponseFormatError:
		return benchscore.InvalidResponseFormatError
	case UnexpectedHTTPStatusCodeError:
		return benchscore.UnexpectedHTTPStatusCodeError
	default:
		panic("unreachable")
	}
}
