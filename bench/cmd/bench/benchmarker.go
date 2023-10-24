package main

import (
	"context"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type benchmarker struct {
	sem *semaphore.Weighted

	popularStreamerClientPool *isupipe.ClientPool
	streamerClientPool        *isupipe.ClientPool
	viewerClientPool          *isupipe.ClientPool

	popularLivestreamPool *isupipe.LivestreamPool
	livestreamPool        *isupipe.LivestreamPool
}

func newBenchmarker(ctx context.Context) *benchmarker {
	lgr := zap.S()

	// FIXME: 広告費用から重さを計算する
	// いったん固定値で設定しておく
	var weight int64 = 10
	lgr.Infof("負荷レベル: %d", weight)

	popularStreamerClientPool := isupipe.NewClientPool(ctx)
	streamerClientPool := isupipe.NewClientPool(ctx)
	viewerClientPool := isupipe.NewClientPool(ctx)

	popularLivestreamPool := isupipe.NewLivestreamPool(ctx)
	livestreamPool := isupipe.NewLivestreamPool(ctx)

	return &benchmarker{
		sem:                       semaphore.NewWeighted(weight),
		popularStreamerClientPool: popularStreamerClientPool,
		streamerClientPool:        streamerClientPool,
		viewerClientPool:          viewerClientPool,
		popularLivestreamPool:     popularLivestreamPool,
		livestreamPool:            livestreamPool,
	}
}

func (b *benchmarker) runLoginWorkers(ctx context.Context) {
	loginFn := func(p *isupipe.ClientPool) func(u *scheduler.User) {
		return func(u *scheduler.User) {
			go func() {
				time.Sleep(100 * time.Millisecond)
				client, err := isupipe.NewClient(
					agent.WithBaseURL(config.TargetBaseURL),
				)
				if err != nil {
					return
				}

				if err := client.Login(ctx, &isupipe.LoginRequest{
					UserName: u.Name,
					Password: u.RawPassword,
				}); err != nil {
					return
				}

				p.Put(ctx, client)
			}()
		}
	}

	scheduler.UserScheduler.RangePopularStreamer(loginFn(b.popularStreamerClientPool))
	scheduler.UserScheduler.RangeStreamer(loginFn(b.streamerClientPool))
	scheduler.UserScheduler.RangeViewer(loginFn(b.viewerClientPool))
}

func (b *benchmarker) load(ctx context.Context) error {
	defer b.sem.Release(1)

	if err := scenario.BasicStreamerColdReserveScenario(ctx, b.streamerClientPool, b.popularLivestreamPool, b.livestreamPool); err != nil {
		return err
	}

	if err := scenario.BasicViewerScenario(ctx, b.viewerClientPool, b.livestreamPool); err != nil {
		return err
	}

	return nil
}

func (b *benchmarker) run(ctx context.Context) error {
	b.runLoginWorkers(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if ok := b.sem.TryAcquire(1); ok {
				go b.load(ctx)
			}
		}
	}
}
