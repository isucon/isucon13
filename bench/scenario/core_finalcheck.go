package scenario

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

func FinalcheckScenario(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {

	// 3秒待つ
loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
			break loop
		}
	}

	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.FinalcheckTimeout),
	)
	if err != nil {
		return err
	}
	// タグ指定なし検索
	searchedStream, err := client.SearchLivestreams(ctx, isupipe.WithLimitQueryParam(config.NumSearchLivestreams))
	if err != nil {
		return err
	}
	lgr := zap.S()
	ids := []int64{}
	for _, stream := range searchedStream {
		ids = append(ids, stream.ID)
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	lgr.Info("Finalcheck SearchLivestreams IDs:", string(b))

	b2, err := json.Marshal(searchedStream)
	if err != nil {
		return err
	}

	if err := os.WriteFile(config.FinalcheckPath, b2, os.ModePerm); err != nil {
		return err
	}

	return nil
}
