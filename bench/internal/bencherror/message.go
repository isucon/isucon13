package bencherror

import (
	"fmt"
)

func UnexpectedHTTPStatusCode(method string, endpoint string, expected int, actual int, err error) error {
	err = fmt.Errorf("%s %s: 期待されたHTTPステータスコードが確認できませんでした(expected:%d, actual:%d): %s", method, endpoint, expected, actual, err)
	return WrapError(UnexpectedHTTPStatusCodeError, err)
}

func Internal(err error) error {
	err = fmt.Errorf("ベンチマーカ側の内部エラーです、運営に連絡してください!: %s", err.Error())
	return WrapError(InternalError, err)
}

func BenchmarkTimeout(method string, endpoint string, err error) error {
	err = fmt.Errorf("%s %s: ベンチマーカのリクエストがタイムアウトしました: %s", method, endpoint, err.Error())
	return WrapError(BenchmarkTimeoutError, err)
}

func BenchmarkCritical(method string, endpoint string, err error) error {
	err = fmt.Errorf("ベンチマーカが継続不可能な致命的エラー(スコアが強制的に0となります): %s", err.Error())
	return WrapError(BenchmarkCriticalError, err)
}

func BenchmarkApplication(method string, endpoint string, err error) error {
	err = fmt.Errorf("%s", err.Error())
	return WrapError(BenchmarkApplicationError, err)
}

func InvalidResponseFormat(method string, endpoint string, err error) error {
	err = fmt.Errorf("レスポンスボディの形式が仕様に不一致です: %s", err)
	return WrapError(InvalidResponseFormatError, err)
}

func DBInconsistency(method string, endpoint string, err error) error {
	err = fmt.Errorf("DBの非一貫性を検出しました: %s", err)
	return WrapError(DBInconsistencyError, err)
}
