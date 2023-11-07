package main

import (
	"context"
	"errors"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type benchmarker struct {
	streamerSem        *semaphore.Weighted
	popularStreamerSem *semaphore.Weighted
	moderatorSem       *semaphore.Weighted
	viewerSem          *semaphore.Weighted
	spammerSem         *semaphore.Weighted
	attackSem          *semaphore.Weighted

	popularStreamerClientPool *isupipe.ClientPool
	streamerClientPool        *isupipe.ClientPool
	viewerClientPool          *isupipe.ClientPool

	popularLivestreamPool *isupipe.LivestreamPool
	livestreamPool        *isupipe.LivestreamPool

	spamPool *isupipe.LivecommentPool

	startAt time.Time
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

	spamPool := isupipe.NewLivecommentPool(ctx)

	return &benchmarker{
		streamerSem:               semaphore.NewWeighted(weight),
		popularStreamerSem:        semaphore.NewWeighted(weight),
		moderatorSem:              semaphore.NewWeighted(weight),
		viewerSem:                 semaphore.NewWeighted(weight * 10), // 配信者の10倍視聴者トラフィックがある
		spammerSem:                semaphore.NewWeighted(weight),
		attackSem:                 semaphore.NewWeighted(weight),
		popularStreamerClientPool: popularStreamerClientPool,
		streamerClientPool:        streamerClientPool,
		viewerClientPool:          viewerClientPool,
		popularLivestreamPool:     popularLivestreamPool,
		livestreamPool:            livestreamPool,
		spamPool:                  spamPool,
		startAt:                   time.Now(),
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

	scheduler.UserScheduler.RangePopularStreamer(loginFn(b.popularStreamerClientPool))
	scheduler.UserScheduler.RangeStreamer(loginFn(b.streamerClientPool))
	scheduler.UserScheduler.RangeViewer(loginFn(b.viewerClientPool))
}

func (b *benchmarker) loadAttack(ctx context.Context) error {
	defer b.attackSem.Release(1)

	now := time.Now()
	parallelism := 5 + (now.Sub(b.startAt) / time.Second / 10)

	if err := scenario.DnsWaterTortureAttackScenario(ctx, int(parallelism)); err != nil {
		return err
	}

	return nil
}

func (b *benchmarker) loadStreamer(ctx context.Context) error {
	defer b.streamerSem.Release(1)
	eg, childCtx := errgroup.WithContext(ctx)

	if err := scenario.BasicStreamerColdReserveScenario(childCtx, b.streamerClientPool, b.popularLivestreamPool, b.livestreamPool); err != nil {
		return err
	}

	return eg.Wait()
}

func (b *benchmarker) loadPopularStreamer(ctx context.Context) error {
	defer b.popularStreamerSem.Release(1)

	return nil
}

// moderateが成功するなら可能な限り高速にmoderationしなければならない
func (b *benchmarker) loadModerator(ctx context.Context) error {
	defer b.moderatorSem.Release(1)

	eg, childCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return scenario.BasicStreamerModerateScenario(childCtx, b.popularStreamerClientPool)
	})

	eg.Go(func() error {
		return scenario.BasicStreamerModerateScenario(childCtx, b.streamerClientPool)
	})

	// FIXME: aggressiveは正常系の影響受けたくない
	// eg.Go(func() error {
	// 	return scenario.AggressiveStreamerModerateScenario(childCtx)
	// })

	return eg.Wait()
}

func (b *benchmarker) loadViewer(ctx context.Context) error {
	defer b.viewerSem.Release(1)
	eg, childCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return scenario.BasicViewerScenario(childCtx, b.viewerClientPool, b.livestreamPool)
	})

	eg.Go(func() error {
		return scenario.BasicViewerReportScenario(childCtx, b.viewerClientPool, b.spamPool)
	})

	return eg.Wait()
}

func (b *benchmarker) loadSpammer(ctx context.Context) error {
	defer b.spammerSem.Release(1)

	if err := scenario.ViewerSpamScenario(ctx, b.viewerClientPool, b.livestreamPool, b.spamPool); err != nil {
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
			if ok := b.popularStreamerSem.TryAcquire(1); ok {
				go b.loadPopularStreamer(ctx)
			}
			if ok := b.moderatorSem.TryAcquire(1); ok {
				go b.loadModerator(ctx)
			}
			if ok := b.viewerSem.TryAcquire(1); ok {
				go b.loadViewer(ctx)
			}
			if ok := b.spammerSem.TryAcquire(1); ok {
				go b.loadSpammer(ctx)
			}
			if ok := b.attackSem.TryAcquire(1); ok {
				go b.loadAttack(ctx)
			}
		}
	}
}
