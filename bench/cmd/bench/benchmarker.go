package main

import (
	"context"
	"errors"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type benchmarker struct {
	streamerSem *semaphore.Weighted
	viewerSem   *semaphore.Weighted
	attackSem   *semaphore.Weighted

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
	var weight int64 = int64(config.AdvertiseCost)
	lgr.Infof("負荷レベル: %d", weight)

	popularStreamerClientPool := isupipe.NewClientPool(ctx)
	streamerClientPool := isupipe.NewClientPool(ctx)
	viewerClientPool := isupipe.NewClientPool(ctx)

	popularLivestreamPool := isupipe.NewLivestreamPool(ctx)
	livestreamPool := isupipe.NewLivestreamPool(ctx)

	return &benchmarker{
		streamerSem:               semaphore.NewWeighted(weight),
		viewerSem:                 semaphore.NewWeighted(weight * 10), // 配信者の10倍視聴者トラフィックがある
		attackSem:                 semaphore.NewWeighted(weight),
		popularStreamerClientPool: popularStreamerClientPool,
		streamerClientPool:        streamerClientPool,
		viewerClientPool:          viewerClientPool,
		popularLivestreamPool:     popularLivestreamPool,
		livestreamPool:            livestreamPool,
	}
}

func (b *benchmarker) runClientProviders(ctx context.Context) {
	loginFn := func(p *isupipe.ClientPool) func(u *scheduler.User) {
		return func(u *scheduler.User) {
			go func() {
				client, err := isupipe.NewClient(
					agent.WithBaseURL(config.TargetBaseURL),
				)
				if err != nil {
					return
				}

				if _, err := client.Register(ctx, &isupipe.RegisterRequest{
					Name:        u.Name,
					DisplayName: u.DisplayName,
					Description: u.Description,
					Password:    u.RawPassword,
					Theme: isupipe.Theme{
						DarkMode: true,
					},
				}); err != nil {
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

	// FIXME: Rangeをctxで中断できるように
	scheduler.UserScheduler.RangePopularStreamer(loginFn(b.popularStreamerClientPool))
	scheduler.UserScheduler.RangeStreamer(loginFn(b.streamerClientPool))
	scheduler.UserScheduler.RangeViewer(loginFn(b.viewerClientPool))
}

func (b *benchmarker) loadAttack(ctx context.Context) error {
	defer b.attackSem.Release(1)

	if err := scenario.DnsWaterTortureAttackScenario(ctx); err != nil {
		return err
	}

	return nil
}

func (b *benchmarker) loadStreamer(ctx context.Context) error {
	defer b.streamerSem.Release(1)

	if err := scenario.BasicStreamerColdReserveScenario(ctx, b.streamerClientPool, b.popularLivestreamPool, b.livestreamPool); err != nil {
		return err
	}

	return nil
}

func (b *benchmarker) loadViewer(ctx context.Context) error {
	defer b.viewerSem.Release(1)

	if err := scenario.BasicViewerScenario(ctx, b.viewerClientPool, b.livestreamPool); err != nil {
		return err
	}

	return nil
}

func (b *benchmarker) run(ctx context.Context) error {
	lgr := zap.S()

	b.runClientProviders(ctx)
	for {
		select {
		case <-ctx.Done():
			lgr.Info("ベンチマーク走行の停止中です")
			return nil
		default:
			if err := bencherror.CheckViolation(); err != nil && errors.Is(err, bencherror.ErrViolation) {
				return err
			}
			if ok := b.streamerSem.TryAcquire(1); ok {
				go b.loadStreamer(ctx)
			}
			if ok := b.viewerSem.TryAcquire(1); ok {
				go b.loadViewer(ctx)
			}
			if ok := b.attackSem.TryAcquire(1); ok {
				go b.loadAttack(ctx)
			}
		}
	}
}
