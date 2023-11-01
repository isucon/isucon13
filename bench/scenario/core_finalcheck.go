package scenario

import (
	"context"

	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

func FinalcheckScenario(ctx context.Context, client *isupipe.Client) error {
	lgr := zap.S()

	result, err := client.GetPaymentResult(ctx)
	if err != nil {
		return err
	}

	// 金額チェック
	// total := scheduler.GetTotal()
	// if result.Total

	lgr.Infof("result = %+v\n", result)

	return nil
}
