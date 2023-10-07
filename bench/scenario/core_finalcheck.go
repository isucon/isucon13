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

	// payments := result.Payments

	// var found bool

	// 金額チェック
	// total := scheduler.GetTotal()
	// if result.Total

	// 予約の整合性チェック
	for _, payment := range result.Payments {
		_ = payment
		// payment.ReservationId
	}

	lgr.Infof("result = %+v\n", result)

	return nil
}
