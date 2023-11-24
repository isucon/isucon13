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
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
	"go.uber.org/zap"
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

type LoginCounter struct {
	sync.RWMutex
	cnt uint64
}

func (c *LoginCounter) Inc() {
	c.Lock()
	defer c.Unlock()
	c.cnt++
}

func (c *LoginCounter) Get() uint64 {
	c.RLock()
	defer c.RUnlock()
	return c.cnt
}

func (c *LoginCounter) WaitUntil(threshold uint64) chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		for c.Get() >= threshold {
		}
	}()
	return ch
}

type benchmarker struct {
	contestantLogger *zap.Logger

	streamerSem      *semaphore.Weighted
	moderatorSem     *semaphore.Weighted
	viewerSem        *semaphore.Weighted
	viewerReportSem  *semaphore.Weighted
	spammerSem       *semaphore.Weighted
	attackSem        *semaphore.Weighted
	attackParallelis int

	// login
	streamerLoginSem     *semaphore.Weighted
	streamerLoginCounter *LoginCounter
	viewerLoginSem       *semaphore.Weighted
	viewerLoginCounter   *LoginCounter

	longStreamerClientPool *isupipe.ClientPool
	streamerClientPool     *isupipe.ClientPool
	viewerClientPool       *isupipe.ClientPool

	livestreamPool *isupipe.LivestreamPool

	spamPool *isupipe.LivecommentPool

	scenarioCounter *score.Score

	startAt time.Time
}

func powWeightSize(m int) int64 {
	return int64(math.Pow(2, float64(m)))
}

func newBenchmarker(ctx context.Context, contestantLogger *zap.Logger) *benchmarker {
	var weight int64 = int64(config.BaseParallelism)
	// いま負荷レベルは固定値なので選手に見せる意味がない
	// contestantLogger.Info("負荷レベル", zap.Int64("level", weight))

	longStreamerClientPool := isupipe.NewClientPool(ctx)
	streamerClientPool := isupipe.NewClientPool(ctx)
	viewerClientPool := isupipe.NewClientPool(ctx)

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
		contestantLogger:       contestantLogger,
		streamerSem:            semaphore.NewWeighted(weight),
		moderatorSem:           semaphore.NewWeighted(weight),
		viewerSem:              semaphore.NewWeighted(weight * 10), // 配信者の10倍視聴者トラフィックがある
		viewerReportSem:        semaphore.NewWeighted(weight),
		spammerSem:             semaphore.NewWeighted(weight * 2), // 視聴者の２倍はスパム投稿者が潜んでいる
		attackSem:              semaphore.NewWeighted(512),        // 攻撃を段階的に大きくする最大値
		attackParallelis:       2,
		streamerLoginSem:       semaphore.NewWeighted(weight),
		streamerLoginCounter:   new(LoginCounter),
		viewerLoginSem:         semaphore.NewWeighted(weight),
		viewerLoginCounter:     new(LoginCounter),
		longStreamerClientPool: longStreamerClientPool,
		streamerClientPool:     streamerClientPool,
		viewerClientPool:       viewerClientPool,
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
	loginFn := func(p *isupipe.ClientPool, sem *semaphore.Weighted, cnt *LoginCounter) func(u *scheduler.User) {
		return func(u *scheduler.User) {
			go func() {
				if err := sem.Acquire(ctx, 1); err != nil {
					return
				}
				defer sem.Release(1)

				client, err := isupipe.NewClient(b.contestantLogger,
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

				icon := scheduler.IconSched.GetRandomIcon()
				if _, err := client.PostIcon(ctx, &isupipe.PostIconRequest{
					Image: icon.Image,
				}); err != nil {
					return
				}

				p.Put(ctx, client)
				cnt.Inc()
			}()
		}
	}

	scheduler.UserScheduler.RangeStreamer(loginFn(b.streamerClientPool, b.streamerLoginSem, b.streamerLoginCounter))
	scheduler.UserScheduler.RangeViewer(loginFn(b.viewerClientPool, b.viewerLoginSem, b.viewerLoginCounter))

	<-b.streamerLoginCounter.WaitUntil(config.NumMustTryLogins)
	<-b.viewerLoginCounter.WaitUntil(config.NumMustTryLogins)
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

	defer b.scenarioCounter.Add(DnsWaterTortureAttackScenario)
	if err := scenario.DnsWaterTortureAttackScenario(ctx, httpClient, loadLimiter); err != nil {
		return err
	}
	return nil
}

var prevNumResolved = int64(0)

func (b *benchmarker) loadAttackCoordinator(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
loop:
	for {
		select {
		case <-ticker.C:
			failRate := float64(benchscore.NumDNSFailed()) / float64(benchscore.NumResolves()+benchscore.NumDNSFailed()+1)
			avg := float64(benchscore.NumResolves()-prevNumResolved) / 2.0
			prevNumResolved = benchscore.NumResolves()
			if failRate < 0.01 && avg/float64(b.attackParallelis) > 50.0 {
				new := int(float64(b.attackParallelis) * 1.5)
				if new > 15 {
					new = 15
				}
				if new != b.attackParallelis {
					b.contestantLogger.Info("DNS水責め負荷が上昇します", zap.Int("parallelis", new))
					b.attackParallelis = new
				}
			}
		case <-ctx.Done():
			break loop
		}
	}
}

func (b *benchmarker) loadStreamer(ctx context.Context) error {
	defer b.streamerSem.Release(1)

	if err := scenario.BasicStreamerColdReserveScenario(ctx, b.contestantLogger, b.streamerClientPool, b.livestreamPool); err != nil {
		b.scenarioCounter.Add(BasicStreamerColdReserveFail)
		return err
	}
	b.scenarioCounter.Add(BasicStreamerColdReserve)

	return nil
}

// moderateが成功するなら可能な限り高速にmoderationしなければならない
func (b *benchmarker) loadModerator(ctx context.Context) error {
	defer b.moderatorSem.Release(1)

	if err := scenario.BasicStreamerModerateScenario(ctx, b.contestantLogger, b.streamerClientPool); err != nil {
		b.scenarioCounter.Add(BasicStreamerModerateScenarioFail)
		return err
	}
	b.scenarioCounter.Add(BasicStreamerModerateScenario)
	return nil
}

func (b *benchmarker) loadViewer(ctx context.Context) error {
	defer b.viewerSem.Release(1)

	if err := scenario.BasicViewerScenario(ctx, b.contestantLogger, b.viewerClientPool, b.livestreamPool); err != nil {
		b.scenarioCounter.Add(BasicViewerScenarioFail)
		return err
	}
	b.scenarioCounter.Add(BasicViewerScenario)
	return nil
}

func (b *benchmarker) loadViewerReport(ctx context.Context) error {
	defer b.viewerReportSem.Release(1)

	time.Sleep(1 * time.Second) // XXX: report回りすぎ抑止
	if err := scenario.BasicViewerReportScenario(ctx, b.contestantLogger, b.viewerClientPool, b.spamPool); err != nil {
		b.scenarioCounter.Add(BasicViewerReportScenarioFail)
		return err
	}
	b.scenarioCounter.Add(BasicViewerReportScenario)
	return nil
}

func (b *benchmarker) loadSpammer(ctx context.Context) error {
	defer b.spammerSem.Release(1)

	var spammerGrp sync.WaitGroup

	spammerGrp.Add(1)
	go func() {
		defer spammerGrp.Done()
		if err := scenario.ViewerSpamScenario(ctx, b.contestantLogger, b.viewerClientPool, b.livestreamPool, b.spamPool); err != nil {
			b.scenarioCounter.Add(ViewerSpamScenarioFail)
			return
		}
		b.scenarioCounter.Add(ViewerSpamScenario)
	}()

	spammerGrp.Add(1)
	go func() {
		defer spammerGrp.Done()
		if err := scenario.AggressiveStreamerModerateScenario(ctx, b.contestantLogger, b.streamerClientPool); err != nil {
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

	violateCh := make(chan error) // とめておく bencherror.RunViolationChecker(ctx)

	loadAttackHTTPClient := b.loadAttackHTTPClient()
	loadAttackLimiter := rate.NewLimiter(rate.Limit(3000), 1)
	go func() { b.loadAttackCoordinator(ctx) }()

	for {
		select {
		case <-ctx.Done():
			b.contestantLogger.Info("ベンチマーク走行を停止します")
			return nil
		case err := <-violateCh:
			b.contestantLogger.Warn("仕様違反が検出されたため、ベンチマーク走行を中断します")
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
			if ok := b.viewerSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadViewer(childCtx)
				}()
			}
			if ok := b.viewerReportSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadViewerReport(childCtx)
				}()
			}
			if ok := b.moderatorSem.TryAcquire(1); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.loadModerator(childCtx)
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
