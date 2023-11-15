package main

import (
	"context"
	"sync"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/score"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

var (
	DnsWaterTortureAttackScenario      score.ScoreTag = "dns-watertorture-attack"
	BasicStreamerColdReserve           score.ScoreTag = "streamer-cold-reserve"
	BasicStreamerModerateScenario      score.ScoreTag = "streamer-moderate"
	BasicViewerScenario                score.ScoreTag = "viewer"
	BasicViewerReportScenario          score.ScoreTag = "viewer-report"
	ViewerSpamScenario                 score.ScoreTag = "viewer-spam"
	AggressiveStreamerModerateScenario score.ScoreTag = "aggressive-streamer-moderate"
)

type benchmarker struct {
	streamerSem     *semaphore.Weighted
	longStreamerSem *semaphore.Weighted
	moderatorSem    *semaphore.Weighted
	viewerSem       *semaphore.Weighted
	spammerSem      *semaphore.Weighted
	attackSem       *semaphore.Weighted

	popularStreamerClientPool *isupipe.ClientPool
	streamerClientPool        *isupipe.ClientPool
	viewerClientPool          *isupipe.ClientPool

	popularLivestreamPool *isupipe.LivestreamPool
	livestreamPool        *isupipe.LivestreamPool

	spamPool *isupipe.LivecommentPool

	scenarioCounter *score.Score

	startAt time.Time
}

func newBenchmarker(ctx context.Context) *benchmarker {
	lgr := zap.S()

	var weight int64 = int64(config.BaseParallelism)
	lgr.Infof("負荷レベル: %d", weight)

	popularStreamerClientPool := isupipe.NewClientPool(ctx)
	streamerClientPool := isupipe.NewClientPool(ctx)
	viewerClientPool := isupipe.NewClientPool(ctx)

	popularLivestreamPool := isupipe.NewLivestreamPool(ctx)
	livestreamPool := isupipe.NewLivestreamPool(ctx)

	spamPool := isupipe.NewLivecommentPool(ctx)

	counter := score.NewScore(ctx)
	counter.Set(DnsWaterTortureAttackScenario, 1)
	counter.Set(BasicStreamerColdReserve, 1)
	counter.Set(BasicStreamerModerateScenario, 1)
	counter.Set(BasicViewerScenario, 1)
	counter.Set(BasicViewerReportScenario, 1)
	counter.Set(ViewerSpamScenario, 1)
	counter.Set(AggressiveStreamerModerateScenario, 1)

	return &benchmarker{
		streamerSem:               semaphore.NewWeighted(weight),
		longStreamerSem:           semaphore.NewWeighted(weight),
		moderatorSem:              semaphore.NewWeighted(weight),
		viewerSem:                 semaphore.NewWeighted(weight * 10), // 配信者の10倍視聴者トラフィックがある
		spammerSem:                semaphore.NewWeighted(weight * 20), // 視聴者の２倍はスパム投稿者が潜んでいる
		attackSem:                 semaphore.NewWeighted(weight * 20), // 視聴者と同程度、攻撃を仕掛ける輩がいる
		popularStreamerClientPool: popularStreamerClientPool,
		streamerClientPool:        streamerClientPool,
		viewerClientPool:          viewerClientPool,
		popularLivestreamPool:     popularLivestreamPool,
		livestreamPool:            livestreamPool,
		spamPool:                  spamPool,
		startAt:                   time.Now(),
		scenarioCounter:           score.NewScore(ctx),
	}
}

func (b *benchmarker) ScenarioCounter() score.ScoreTable {
	return b.scenarioCounter.Breakdown()
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
					Username: u.Name,
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

var loadAttackPerSecond int64 = 5000

func (b *benchmarker) loadAttack(ctx context.Context) error {
	defer b.attackSem.Release(1)

	factor := float64(loadAttackPerSecond) / float64(config.AdvertiseCost) * 20.0
	failRate := float64(benchscore.NumDNSFailed()) / float64(benchscore.NumResolves()+benchscore.NumDNSFailed())
	if failRate < 0.01 {
		now := time.Now()
		d := now.Sub(b.startAt) / time.Second
		factor = factor * (1.0 + float64(d)/16.0)
	}
	loadLimiter := rate.NewLimiter(rate.Limit(factor), 1)

	defer b.scenarioCounter.Add(DnsWaterTortureAttackScenario)
	if err := scenario.DnsWaterTortureAttackScenario(ctx, loadLimiter); err != nil {
		return err
	}

	return nil
}

func (b *benchmarker) loadStreamer(ctx context.Context) error {
	defer b.streamerSem.Release(1)
	eg, childCtx := errgroup.WithContext(ctx)

	defer b.scenarioCounter.Add(BasicStreamerColdReserve)
	if err := scenario.BasicStreamerColdReserveScenario(childCtx, b.streamerClientPool, b.popularLivestreamPool, b.livestreamPool); err != nil {
		return err
	}

	return eg.Wait()
}

func (b *benchmarker) loadLongStreamer(ctx context.Context) error {
	defer b.longStreamerSem.Release(1)

	return nil
}

// moderateが成功するなら可能な限り高速にmoderationしなければならない
func (b *benchmarker) loadModerator(ctx context.Context) error {
	defer b.moderatorSem.Release(1)

	eg, childCtx := errgroup.WithContext(ctx)

	defer b.scenarioCounter.Add(BasicStreamerModerateScenario)
	eg.Go(func() error {
		return scenario.BasicStreamerModerateScenario(childCtx, b.popularStreamerClientPool)
	})

	defer b.scenarioCounter.Add(BasicStreamerModerateScenario)
	eg.Go(func() error {
		return scenario.BasicStreamerModerateScenario(childCtx, b.streamerClientPool)
	})

	return eg.Wait()
}

func (b *benchmarker) loadViewer(ctx context.Context) error {
	defer b.viewerSem.Release(1)
	eg, childCtx := errgroup.WithContext(ctx)

	defer b.scenarioCounter.Add(BasicViewerScenario)
	eg.Go(func() error {
		return scenario.BasicViewerScenario(childCtx, b.viewerClientPool, b.livestreamPool)
	})

	defer b.scenarioCounter.Add(BasicViewerReportScenario)
	eg.Go(func() error {
		return scenario.BasicViewerReportScenario(childCtx, b.viewerClientPool, b.spamPool)
	})

	return eg.Wait()
}

func (b *benchmarker) loadSpammer(ctx context.Context) error {
	defer b.spammerSem.Release(1)

	var spammerGrp sync.WaitGroup

	spammerGrp.Add(1)
	go func() {
		defer spammerGrp.Done()
		defer b.scenarioCounter.Add(ViewerSpamScenario)
		scenario.ViewerSpamScenario(ctx, b.viewerClientPool, b.livestreamPool, b.spamPool)
	}()

	spammerGrp.Add(1)
	go func() {
		defer spammerGrp.Done()
		defer b.scenarioCounter.Add(AggressiveStreamerModerateScenario)
		scenario.AggressiveStreamerModerateScenario(ctx, b.streamerClientPool)
	}()

	spammerGrp.Wait()

	return nil
}

func (b *benchmarker) run(ctx context.Context) error {
	lgr := zap.S()

	var wg sync.WaitGroup
	defer wg.Wait()

	childCtx, cancelChildCtx := context.WithCancel(ctx)
	defer cancelChildCtx()

	b.runClientProviders(ctx)
	violateCh := bencherror.RunViolationChecker(ctx)

	for {
		select {
		case <-ctx.Done():
			lgr.Info("ベンチマーク走行を停止します")
			return nil
		case err := <-violateCh:
			lgr.Warn("仕様違反が検出されたため、ベンチマーク走行を中断します")
			lgr.Warnf("仕様違反エラー: %s", err.Error())
			return err
		default:
			if ok := b.streamerSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadStreamer(childCtx)
				}()
			}
			if ok := b.longStreamerSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadLongStreamer(childCtx)
				}()
			}
			if ok := b.moderatorSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadModerator(childCtx)
				}()
			}
			if ok := b.viewerSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadViewer(childCtx)
				}()
			}
			if ok := b.spammerSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadSpammer(childCtx)
				}()
			}
			if ok := b.attackSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadAttack(childCtx)
				}()
			}
		}
	}
}
