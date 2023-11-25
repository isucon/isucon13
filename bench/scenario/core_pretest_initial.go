package scenario

import (
	"context"
	"fmt"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

// 初期データpretest

func normalInitialPaymentPretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	// 初期状態で0円であるか
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	result, err := client.GetPaymentResult(ctx)
	if err != nil {
		return err
	}

	if result.TotalTip != 0 {
		return fmt.Errorf("初期の売上は0ISUでなければなりません")
	}

	return nil
}
