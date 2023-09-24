package bencherror

import (
	"fmt"
	"sync"

	"github.com/isucon/isucandar/failure"
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
)

var (
	benchErrors *failure.Errors
	doneOnce    sync.Once
)

// FIXME: もうちょっと細分化して、エラーの一貫性を持たせたい
func WrapError(code failure.StringCode, err error) error {
	e := failure.NewError(code, err)
	benchErrors.Add(e)
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
