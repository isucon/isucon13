package main

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/logger"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
)

var assetDir string

type BenchResult struct {
	Pass          bool     `json:"pass"`
	Score         int64    `json:"score"`
	Messages      []string `json:"messages"`
	AvailableDays int      `json:"available_days"`
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

func benchmark(ctx context.Context) error {
	lgr := zap.S()

	// pretest, benchmarkにはこれら初期化が必要
	benchscore.InitScore(ctx)
	// bencherror.InitPenalty(ctx)
	bencherror.InitializeErrors(ctx)

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
			Value:       "http://127.0.0.1:12345",
			Destination: &config.TargetBaseURL,
			EnvVar:      "BENCH_TARGET_URL",
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
	},
	Action: func(cliCtx *cli.Context) error {
		ctx := context.Background()

		lgr, err := logger.InitZapLogger()
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		lgr.Infof("webapp: %s", config.TargetBaseURL)

		lgr.Info("===== Prepare benchmarker =====")
		// FIXME: アセット読み込み

		lgr.Info("===== Initialize webapp =====")
		initClient, err := isupipe.NewClient(
			agent.WithBaseURL(config.TargetBaseURL),
			agent.WithTimeout(20*time.Second),
		)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		initializeResp, err := initClient.Initialize(ctx)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		if initializeResp.AdvertiseLevel < 1 || 10 < initializeResp.AdvertiseLevel {
			return cli.NewExitError("不正な広告レベル", 1)
		}
		config.AdvertiseCost = initializeResp.AdvertiseLevel

		lgr.Info("===== Pretest webapp =====")
		pretestClient, err := isupipe.NewClient(
			agent.WithBaseURL(config.TargetBaseURL),
		)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		// pretest, benchmarkにはこれら初期化が必要
		benchscore.InitScore(ctx)
		// bencherror.InitPenalty(ctx)
		bencherror.InitializeErrors(ctx)
		if err := scenario.Pretest(ctx, pretestClient); err != nil {
			return cli.NewExitError(err, 1)
		}

		_ = benchmark(ctx)

		// lgr.Info("===== Final check =====")
		// paymentClient, err := isupipe.NewClient(
		// 	agent.WithBaseURL(config.TargetBaseURL),
		// )
		// if err != nil {
		// 	return cli.NewExitError(err, 1)
		// }

		// if err := scenario.FinalcheckScenario(ctx, paymentClient); err != nil {
		// 	return cli.NewExitError(err, 1)
		// }

		lgr.Info("===== System errors =====")
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

		lgr.Info("===== Calculate final score =====")
		var msgs []string

		score := benchscore.GetFinalScore()
		msgs = append(msgs, fmt.Sprintf("スコア: %d", score))

		profit := benchscore.GetFinalProfit()
		msgs = append(msgs, fmt.Sprintf("売上: %d", profit))
		lgr.Infof("売上: %d", profit)

		penalties := bencherror.GetFinalPenalties()
		for code, penalty := range penalties {
			message := fmt.Sprintf("ペナルティ[%s]: %d", code, penalty)
			msgs = append(msgs, message)
		}

		return nil
	},
}
