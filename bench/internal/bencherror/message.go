package bencherror

import (
	"errors"
	"fmt"
	"net/http"
)

// NOTE: Goのhttp.Clientがcontext.DeadlineExceededをラップして返してくれないので、暫定対応
var ErrTimeout = errors.New("タイムアウトによりリクエスト失敗")

func NewInternalError(err error) error {
	err = fmt.Errorf("[ベンチ本体のエラー] 運営に連絡してください: %w", err)
	return WrapError(SystemError, err)
}

func NewTimeoutError(err error, msg string, args ...interface{}) error {
	message := fmt.Sprintf(msg, args...)
	err = fmt.Errorf("%s: %w", err.Error(), ErrTimeout)
	err = fmt.Errorf("[リクエストタイムアウト] %s: %w", message, err)
	return WrapError(BenchmarkTimeoutError, err)
}

func NewViolationError(err error, msg string, args ...interface{}) error {
	message := fmt.Sprintf(msg, args...)
	err = fmt.Errorf("[仕様違反] %s: %w", message, err)
	return WrapError(BenchmarkViolationError, err)
}

func NewApplicationError(err error, msg string, args ...interface{}) error {
	message := fmt.Sprintf(msg, args...)
	err = fmt.Errorf("[一般エラー] %s: %w", message, err)
	return WrapError(BenchmarkApplicationError, err)
}

func NewHttpError(err error, req *http.Request, msg string, args ...interface{}) error {
	endpoint := fmt.Sprintf("%s %s", req.Method, req.URL.EscapedPath())
	message := fmt.Sprintf(msg, args...)
	err = fmt.Errorf("[一般エラー] %sへのリクエストに対して、%s: %w", endpoint, message, err)
	return WrapError(BenchmarkApplicationError, err)
}

func NewHttpStatusError(req *http.Request, expected int, actual int) error {
	endpoint := fmt.Sprintf("%s %s", req.Method, req.URL.EscapedPath())
	err := fmt.Errorf("[一般エラー] %s へのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d, actual:%d)", endpoint, expected, actual)
	return WrapError(BenchmarkApplicationError, err)
}

func NewHttpResponseError(err error, req *http.Request) error {
	endpoint := fmt.Sprintf("%s %s", req.Method, req.URL.EscapedPath())
	err = fmt.Errorf("[一般エラー] %s へのリクエストに対して、レスポンスボディの形式が不正です: %w", endpoint, err)
	return WrapError(BenchmarkApplicationError, err)
}

func NewAssertionError(err error, msg string, args ...interface{}) error {
	message := fmt.Sprintf(msg, args...)
	err = fmt.Errorf("[仕様違反] %s: %w", message, err)
	return WrapError(BenchmarkViolationError, err)
}
