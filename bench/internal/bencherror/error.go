package bencherror

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/isucon/isucandar/failure"
	"go.uber.org/zap"
)

var (
	ErrViolation = errors.New("仕様違反が発生しました")
	ErrSystem    = errors.New("ベンチマーカーに問題が発生しました")
)

var (
	// SystemError は、ベンチマーカ内部のエラー (継続不可。fail)
	SystemError failure.StringCode = "system"
	// BenchmarkApplicationError は、ベンチ走行中の一般的なエラー (減点)
	BenchmarkApplicationError failure.StringCode = "benchmark-application"
	// BenchmarkCriticalError は、仕様違反エラー (fail)
	BenchmarkViolationError failure.StringCode = "benchmark-critical"
	// BenchmarkTimeoutError は、タイムアウトによるエラー (減点)
	BenchmarkTimeoutError failure.StringCode = "benchmark-timeout"
)

var (
	benchErrors *failure.Errors
	doneOnce    sync.Once
)

func InitErrors(ctx context.Context) {
	benchErrors = failure.NewErrors(ctx)
}

func WrapError(code failure.StringCode, err error) error {
	benchErrors.Add(failure.NewError(code, err))
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

func Done() {
	doneOnce.Do(func() {
		benchErrors.Close()
	})
}

func CheckViolation() error {
	counts := benchErrors.Count()

	systemErrorCount, ok := counts[string(SystemError)]
	if !ok {
		systemErrorCount = 0
	}
	if systemErrorCount > 0 {
		return ErrSystem
	}

	violationCount, ok := counts[string(BenchmarkViolationError)]
	if !ok {
		violationCount = 0
	}
	if violationCount > 0 {
		return ErrViolation
	}

	return nil
}

func RunViolationChecker(ctx context.Context) chan error {
	violate := make(chan error, 1)
	go func() {
		defer close(violate)
		for {
			select {
			case <-ctx.Done():
			default:
				if err := CheckViolation(); err != nil {
					violate <- err
					return
				}
			}
		}
	}()
	return violate
}
