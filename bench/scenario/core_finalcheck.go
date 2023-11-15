package scenario

import (
	"context"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
)

func FinalcheckScenario(ctx context.Context, dnsResolver *resolver.DNSResolver) error {

	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.FinalcheckTimeout),
	)
	if err != nil {
		return err
	}

	// FIXME: ライブコメント存在チェック
	_ = client

	return nil
}
