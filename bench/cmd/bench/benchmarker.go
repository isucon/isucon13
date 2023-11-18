package main

import (
	"context"
	"crypto/tls"
	"math"
	"net"
	"net/http"
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
	DnsWaterTortureAttackScenario          score.ScoreTag = "dns-watertorture-attack"
	BasicStreamerColdReserve               score.ScoreTag = "streamer-cold-reserve"
	BasicStreamerColdReserveFail           score.ScoreTag = "streamer-cold-reserve-fail"
	BasicStreamerModerateScenario          score.ScoreTag = "streamer-moderate"
	BasicStreamerModerateScenarioFail      score.ScoreTag = "streamer-moderate-fail"
	BasicViewerScenario                    score.ScoreTag = "viewer"
	BasicViewerScenarioFail                score.ScoreTag = "viewer-fail"
	BasicViewerReportScenario              score.ScoreTag = "viewer-report"
	BasicViewerReportScenarioFail          score.ScoreTag = "viewer-report-fail"
	ViewerSpamScenario                     score.ScoreTag = "viewer-spam"
	ViewerSpamScenarioFail                 score.ScoreTag = "viewer-spam-fail"
	AggressiveStreamerModerateScenario     score.ScoreTag = "aggressive-streamer-moderate"
	AggressiveStreamerModerateScenarioFail score.ScoreTag = "aggressive-streamer-moderate-fail"
)

type benchmarker struct {
	streamerSem      *semaphore.Weighted
	longStreamerSem  *semaphore.Weighted
	moderatorSem     *semaphore.Weighted
	viewerSem        *semaphore.Weighted
	spammerSem       *semaphore.Weighted
	attackSem        *semaphore.Weighted
	attackParallelis int

	longStreamerClientPool *isupipe.ClientPool
	streamerClientPool     *isupipe.ClientPool
	viewerClientPool       *isupipe.ClientPool

	popularLivestreamPool *isupipe.LivestreamPool
	livestreamPool        *isupipe.LivestreamPool

	spamPool *isupipe.LivecommentPool

	scenarioCounter *score.Score

	startAt time.Time
}

func powWeightSize(m int) int64 {
	return int64(math.Pow(2, float64(m)))
}

func newBenchmarker(ctx context.Context) *benchmarker {
	lgr := zap.S()

	var weight int64 = int64(config.BaseParallelism)
	lgr.Infof("負荷レベル: %d", weight)

	longStreamerClientPool := isupipe.NewClientPool(ctx)
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
		streamerSem:            semaphore.NewWeighted(weight),
		longStreamerSem:        semaphore.NewWeighted(weight),
		moderatorSem:           semaphore.NewWeighted(weight),
		viewerSem:              semaphore.NewWeighted(weight * 1), // 配信者の10倍視聴者トラフィックがある
		spammerSem:             semaphore.NewWeighted(weight * 2), // 視聴者の２倍はスパム投稿者が潜んでいる
		attackSem:              semaphore.NewWeighted(512),        // 攻撃を段階的に大きくする最大値
		attackParallelis:       2,
		longStreamerClientPool: longStreamerClientPool,
		streamerClientPool:     streamerClientPool,
		viewerClientPool:       viewerClientPool,
		popularLivestreamPool:  popularLivestreamPool,
		livestreamPool:         livestreamPool,
		spamPool:               spamPool,
		startAt:                time.Now(),
		scenarioCounter:        score.NewScore(ctx),
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

	scheduler.UserScheduler.RangeStreamer(loginFn(b.streamerClientPool))
	scheduler.UserScheduler.RangeViewer(loginFn(b.viewerClientPool))
}

// dns水責め攻撃につかうhttp client
func (b *benchmarker) loadAttackHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if ip, ok := ctx.Value(config.AttackHTTPClientContextKey).(string); ok {
				addr = ip
			}
			return dialer.DialContext(ctx, network, addr)
		},
		TLSHandshakeTimeout: 5 * time.Second,
		// 複数labelがあるのでskip verify
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: false},
		ExpectContinueTimeout: 5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (b *benchmarker) loadAttack(ctx context.Context, asize int64, httpClient *http.Client, loadLimiter *rate.Limiter) error {
	defer b.attackSem.Release(asize)

	failRate := float64(benchscore.NumDNSFailed()) / float64(benchscore.NumResolves()+benchscore.NumDNSFailed())
	if failRate < 0.01 {
		now := time.Now()
		d := now.Sub(b.startAt) / time.Second
		b.attackParallelis = int(2.0 * (1.0 + float64(d)/12.0))
		if b.attackParallelis > 10 {
			b.attackParallelis = 10
		}
	}

	defer b.scenarioCounter.Add(DnsWaterTortureAttackScenario)
	if err := scenario.DnsWaterTortureAttackScenario(ctx, httpClient, loadLimiter); err != nil {
		return err
	}
	return nil
}

func (b *benchmarker) loadStreamer(ctx context.Context) error {
	defer b.streamerSem.Release(1)

	if err := scenario.BasicStreamerColdReserveScenario(ctx, b.streamerClientPool, b.popularLivestreamPool, b.livestreamPool); err != nil {
		b.scenarioCounter.Add(BasicStreamerColdReserveFail)
		return err
	}
	b.scenarioCounter.Add(BasicStreamerColdReserve)

	return nil
}

func (b *benchmarker) loadLongStreamer(ctx context.Context) error {
	defer b.longStreamerSem.Release(1)

	return nil
}

// moderateが成功するなら可能な限り高速にmoderationしなければならない
func (b *benchmarker) loadModerator(ctx context.Context) error {
	defer b.moderatorSem.Release(1)

	eg, childCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := scenario.BasicStreamerModerateScenario(childCtx, b.streamerClientPool); err != nil {
			b.scenarioCounter.Add(BasicStreamerModerateScenarioFail)
			return err
		}
		b.scenarioCounter.Add(BasicStreamerModerateScenario)
		return nil
	})

	return eg.Wait()
}

func (b *benchmarker) loadViewer(ctx context.Context) error {
	defer b.viewerSem.Release(1)
	eg, childCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := scenario.BasicViewerScenario(childCtx, b.viewerClientPool, b.livestreamPool); err != nil {
			b.scenarioCounter.Add(BasicViewerScenarioFail)
			return err
		}
		b.scenarioCounter.Add(BasicViewerScenario)
		return nil
	})

	eg.Go(func() error {
		if err := scenario.BasicViewerReportScenario(childCtx, b.viewerClientPool, b.spamPool); err != nil {
			b.scenarioCounter.Add(BasicViewerReportScenarioFail)
			return err
		}
		b.scenarioCounter.Add(BasicViewerReportScenario)
		return nil
	})

	return eg.Wait()
}

func (b *benchmarker) loadSpammer(ctx context.Context) error {
	defer b.spammerSem.Release(1)

	var spammerGrp sync.WaitGroup

	spammerGrp.Add(1)
	go func() {
		defer spammerGrp.Done()
		if err := scenario.ViewerSpamScenario(ctx, b.viewerClientPool, b.livestreamPool, b.spamPool); err != nil {
			b.scenarioCounter.Add(ViewerSpamScenarioFail)
			return
		}
		b.scenarioCounter.Add(ViewerSpamScenario)
	}()

	spammerGrp.Add(1)
	go func() {
		defer spammerGrp.Done()
		if err := scenario.AggressiveStreamerModerateScenario(ctx, b.streamerClientPool); err != nil {
			b.scenarioCounter.Add(AggressiveStreamerModerateScenarioFail)
			return
		}
		b.scenarioCounter.Add(AggressiveStreamerModerateScenario)
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

	loadAttackHTTPClient := b.loadAttackHTTPClient()
	loadAttackLimiter := rate.NewLimiter(rate.Limit(900), 1)

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
			asize := int64(512.0 / float64(b.attackParallelis))
			if ok := b.attackSem.TryAcquire(asize); ok {
				wg.Add(1)
				asize := asize
				go func() {
					defer wg.Done()
					b.loadAttack(childCtx, asize, loadAttackHTTPClient, loadAttackLimiter)
				}()
			}
		}
	}
}
