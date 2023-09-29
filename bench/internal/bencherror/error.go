package bencherror

import (
	"context"
	"fmt"
	"sync"

	"github.com/isucon/isucandar/failure"
)

var (
	// SystemError は、ベンチマーカ内部のエラー (継続不可。FIXME: スコアについての扱いどうするか)
	SystemError failure.StringCode = "system"
	// InitializeError は、webapp初期化のエラー (即fail)
	InitializeError failure.StringCode = "initialize"
	// PreTestError は、pretest実行中のエラー (即fail)
	PreTestError failure.StringCode = "pretest"
	// BenchmarkApplicationError は、ベンチ走行中の一般的なエラー (減点)
	BenchmarkApplicationError failure.StringCode = "benchmark-application"
	// BenchmarkCriticalError は、仕様違反エラー (fail)
	BenchmarkViolationError failure.StringCode = "benchmark-critical"
	// BenchmarkTimeoutError は、タイムアウトによるエラー (減点)
	BenchmarkTimeoutError failure.StringCode = "benchmark-timeout"
	// FinalCheckError は、売上金額突合処理(finalcheck)でのエラー (即fail)
	FinalCheckError failure.StringCode = "finalcheck"
)

var (
	benchErrors *failure.Errors
	doneOnce    sync.Once
)

func InitializeErrors(ctx context.Context) {
	benchErrors = failure.NewErrors(ctx)
}

func WrapError(code failure.StringCode, err error) error {
	benchErrors.Add(failure.NewError(code, err))
	AddPenalty(code)
	return fmt.Errorf("%s: %w", code, err)
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
