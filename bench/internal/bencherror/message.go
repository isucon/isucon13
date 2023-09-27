package bencherror

import (
	"fmt"
)

func UnexpectedHTTPStatusCode(expected int, actual int, err error) error {
	err = fmt.Errorf("期待されたHTTPステータスコードが確認できませんでした(expected:%d, actual:%d): %s", expected, actual, err)
	return WrapError(UnexpectedHTTPStatusCodeError, err)
}

func Internal(err error) error {
	err = fmt.Errorf("ベンチマーカ側の内部エラーです、運営に連絡してください!: %s", err.Error())
	return WrapError(InternalError, err)
}

func BenchmarkTimeout(err error) error {
	err = fmt.Errorf("ベンチマーカのリクエストがタイムアウトしました: %s", err.Error())
	return WrapError(BenchmarkTimeoutError, err)
}

func BenchmarkCritical(err error) error {
	err = fmt.Errorf("ベンチマーカが継続不可能な致命的エラー: %s", err.Error())
	return WrapError(BenchmarkCriticalError, err)
}

func BenchmarkApplication(err error) error {
	err = fmt.Errorf("%s", err.Error())
	return WrapError(BenchmarkApplicationError, err)
}

func InvalidResponseFormat(err error) error {
	err = fmt.Errorf("レスポンスボディの形式が仕様に不一致です: %s", err)
	return WrapError(InvalidResponseFormatError, err)
}

func DBInconsistency(err error) error {
	err = fmt.Errorf("DBの非一貫性を検出しました: %s", err)
	return WrapError(DBInconsistencyError, err)
}
