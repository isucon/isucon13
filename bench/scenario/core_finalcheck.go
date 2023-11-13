package scenario

import (
	"context"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

func FinalcheckScenario(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	lgr := zap.S()

	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.FinalcheckTimeout),
	)
	if err != nil {
		return err
	}

	result, err := client.GetPaymentResult(ctx)
	if err != nil {
		return err
	}

	// FIXME: 統計情報の検証

	// 金額チェック
	// total := scheduler.GetTotal()
	// if result.Total

	lgr.Infof("result = %+v\n", result)

	return nil
}
