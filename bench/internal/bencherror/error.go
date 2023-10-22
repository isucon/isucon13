package bencherror

import (
	"context"
	"fmt"
	"sync"

	"github.com/isucon/isucandar/failure"
	"go.uber.org/zap"
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
	// AddPenalty(code)
	return fmt.Errorf("%s: %w", code, err)
}

func GetFinalErrorMessages() map[string][]string {
	lgr := zap.S()

	doneOnce.Do(func() {
		benchErrors.Done()
	})

	// メッセージを整形した上でコード種別ごと詰め直して返す
	m := make(map[string][]string)
	for _, e := range benchErrors.All() {
		code := failure.GetErrorCode(e)

		failureErr, ok := e.(*failure.Error)
		if !ok {
			lgr.Warnf("ベンチマーカーが制御できないエラーが発生しました: %+v", e)
			continue
		}

		err := failureErr.Unwrap()
		if err == nil {
			lgr.Warnf("ベンチマーカーが制御できないエラーが発生しました: %+v", e)
		}

		if _, ok := m[code]; !ok {
			m[code] = []string{err.Error()}
		} else {
			m[code] = append(m[code], err.Error())
		}
	}

	return m
}

func GetFinalPenalties() map[string]int64 {
	doneOnce.Do(func() {
		benchErrors.Done()
	})

	return benchErrors.Count()
}
