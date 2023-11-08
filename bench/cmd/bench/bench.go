package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/logger"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
)

var assetDir string

type BenchResult struct {
	Pass          bool     `json:"pass"`
	Score         int64    `json:"score"`
	Messages      []string `json:"messages"`
	AdvertiseCost int      `json:"advertise_cost"`
	Language      string   `json:"language"`
}

// UniqueMsgs は重複除去したメッセージ配列を返します
func uniqueMsgs(msgs []string) (uniqMsgs []string) {
	dedup := map[string]struct{}{}
	for _, msg := range msgs {
		if _, ok := dedup[msg]; ok {
			continue
		}
		dedup[msg] = struct{}{}
		uniqMsgs = append(uniqMsgs, msg)
	}
	return
}

func dumpFailedResult(msgs []string) {
	lgr := zap.S()

	b, err := json.Marshal(&BenchResult{
		Pass:          false,
		Score:         0,
		Messages:      msgs,
		AdvertiseCost: int(config.AdvertiseCost),
		Language:      config.Language,
	})
	if err != nil {
		lgr.Warnf("失格判定結果書き出しに失敗. 運営に連絡してください: messages=%+v, err=%+v", msgs, err)
		fmt.Println(fmt.Sprintf(`{"pass": false, "score": 0, "messages": ["%s"]}`, string(b)))
		return
	}

	fmt.Println(string(b))
}

func benchmark(ctx context.Context) error {
	lgr := zap.S()

	// pretest, benchmarkにはこれら初期化が必要
	benchscore.InitCounter(ctx)
	benchscore.InitProfit(ctx)
	bencherror.InitErrors(ctx)

	benchCtx, cancelBench := context.WithTimeout(ctx, config.DefaultBenchmarkTimeout)
	defer cancelBench()

	benchmarker := newBenchmarker(benchCtx)
	if err := benchmarker.run(benchCtx); err != nil {
		lgr.Warnf("ベンチマーク走行エラー", zap.Error(err))
		return err
	}

	return nil
}

var run = cli.Command{
	Name:  "run",
	Usage: "ベンチマーク実行",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "target",
			Value:       fmt.Sprintf("http://pipe.u.isucon.dev:%d", config.TargetPort),
			Destination: &config.TargetBaseURL,
			EnvVar:      "BENCH_TARGET_URL",
		},
		cli.StringFlag{
			Name:        "nameserver",
			Value:       "127.0.0.1",
			Destination: &config.TargetNameserver,
			EnvVar:      "BENCH_NAMESERVER",
		},
		cli.IntFlag{
			Name:        "dns-port",
			Value:       53,
			Destination: &config.DNSPort,
			EnvVar:      "BENCH_DNS_PORT",
		},
		cli.StringFlag{
			Name:        "assetdir",
			Value:       "assets/testdata",
			Destination: &assetDir,
			EnvVar:      "BENCH_ASSETDIR",
		},
		cli.StringFlag{
			Name:        "webhookurl",
			Destination: &config.SlackWebhookURL,
			EnvVar:      "BENCH_SLACK_WEBHOOK_URL",
		},
		cli.StringFlag{
			Name:        "logpath",
			Destination: &config.LogPath,
			EnvVar:      "BENCH_LOG_PATH",
			Value:       "/tmp/isupipe-benchmarker.log",
		},
	},
	Action: func(cliCtx *cli.Context) error {
		ctx := context.Background()

		lgr, err := logger.InitZapLogger()
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		lgr.Infof("webapp: %s", config.TargetBaseURL)
		lgr.Infof("nameserver: %s", net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort)))

		lgr.Info("===== Prepare benchmarker =====")
		// FIXME: アセット読み込み

		lgr.Info("webappの初期化を行います")
		initClient, err := isupipe.NewClient(
			resolver.NewDNSResolver(),
			agent.WithBaseURL(config.TargetBaseURL),
			agent.WithTimeout(1*time.Minute),
		)
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		// FIXME: initialize以後のdumpFailedResult、ポータル報告への書き出しを実装
		// Actionsの結果にも乗ってしまうが、サイズ的に問題ないか
		// ベンチの出力変動が落ち着いてから実装する

		initializeResp, err := initClient.Initialize(ctx)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		if initializeResp.AdvertiseLevel < 1 {
			return cli.NewExitError("不正な広告レベル", 1)
		}
		config.AdvertiseCost = initializeResp.AdvertiseLevel
		config.Language = initializeResp.Language

		lgr.Info("ベンチマーク走行前のデータ整合性チェックを行います")
		pretestDNSResolver := resolver.NewDNSResolver()
		pretestDNSResolver.ResolveAttempts = 10
		pretestClient, err := isupipe.NewClient(
			pretestDNSResolver,
			agent.WithBaseURL(config.TargetBaseURL),
			agent.WithTimeout(10*time.Second),
		)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		// pretest, benchmarkにはこれら初期化が必要
		benchscore.InitCounter(ctx)
		benchscore.InitProfit(ctx)
		bencherror.InitErrors(ctx)
		if err := scenario.Pretest(ctx, pretestClient); err != nil {
			return cli.NewExitError(err, 1)
		}

		lgr.Info("ベンチマーク走行を開始します")
		_ = benchmark(ctx)

		benchscore.DoneCounter()
		benchscore.DoneProfit()
		bencherror.Done()
		lgr.Info("ベンチマーク走行終了")

		lgr.Info("===== 最終チェック =====")
		finalcheckDNSResolver := resolver.NewDNSResolver()
		finalcheckDNSResolver.ResolveAttempts = 10
		finalcheckClient, err := isupipe.NewClient(
			finalcheckDNSResolver,
			agent.WithBaseURL(config.TargetBaseURL),
			agent.WithTimeout(10*time.Second),
		)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		if err := scenario.FinalcheckScenario(ctx, finalcheckClient); err != nil {
			return cli.NewExitError(err, 1)
		}

		lgr.Info("===== ベンチ走行中エラー (重複排除済み) =====")
		var systemErrors []string
		for _, msgs := range bencherror.GetFinalErrorMessages() {
			for _, msg := range msgs {
				systemErrors = append(systemErrors, msg)
			}
		}
		systemErrors = uniqueMsgs(systemErrors)
		for _, systemError := range systemErrors {
			lgr.Warn(systemError)
		}

		lgr.Info("===== ベンチ走行結果 =====")
		var msgs []string

		var (
			tooManySlows = benchscore.GetByTag(benchscore.TooSlow)
			tooManySpams = benchscore.GetByTag(benchscore.TooManySpam)
		)
		msgs = append(msgs, fmt.Sprintf("遅延による離脱: %d", tooManySlows))
		msgs = append(msgs, fmt.Sprintf("スパムによる離脱: %d", tooManySpams))
		lgr.Infof("遅延離脱=%d, スパム離脱=%d", tooManySlows, tooManySpams)

		numResolves := benchscore.GetByTag(benchscore.DNSResolve)
		numDNSFailed := benchscore.GetByTag(benchscore.DNSFailed)
		msgs = append(msgs, fmt.Sprintf("名前解決成功数 %d", numResolves))
		msgs = append(msgs, fmt.Sprintf("名前解決失敗数 %d", numDNSFailed))
		lgr.Infof("名前解決成功数: %d", numResolves)
		lgr.Infof("名前解決失敗数: %d", numDNSFailed)

		profit := benchscore.GetTotalProfit()
		msgs = append(msgs, fmt.Sprintf("売上: %d", profit))
		lgr.Infof("売上: %d", profit)

		return nil
	},
}
