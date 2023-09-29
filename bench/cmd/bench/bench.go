package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli"

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

var run = cli.Command{
	Name:  "run",
	Usage: "ベンチマーク実行",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "target",
			Value:       "http://localhost",
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

		lgr.Info("===== Prepare benchmarker =====")
		// FIXME: アセット読み込み

		lgr.Info("===== Initialize webapp =====")
		initClient, err := isupipe.NewClient(
			agent.WithBaseURL(config.TargetBaseURL),
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

		if err := scenario.Pretest(ctx, pretestClient); err != nil {
			return cli.NewExitError(err, 1)
		}

		benchCtx, cancelBench := context.WithTimeout(ctx, config.DefaultBenchmarkTimeout)
		defer cancelBench()

		benchmarker := newBenchmarker()
		benchscore.InitScore(ctx)
		bencherror.InitPenalty(ctx)
		bencherror.InitializeErrors(ctx)

		lgr.Info("===== Benchmark webapp - Season1 =====")
		lgr.Info("Season1の達成条件は以下のとおりです")
		lgr.Info("スコア >= 200000")
		if err := benchmarker.runSeason1(benchCtx); err != nil {
			return cli.NewExitError(err, 1)
		}

		lgr.Info("===== Benchmark webapp - Season2 =====")
		lgr.Info("Season2の達成条件は以下のとおりです")
		lgr.Info("スコア >= 400000")

		lgr.Info("===== Benchmark webapp - Season3 =====")
		lgr.Info("Season3の達成条件は以下のとおりです")
		lgr.Info("スコア >= 600000")

		lgr.Info("===== Benchmark webapp - Season4 =====")
		lgr.Info("Season4に達成条件はありません")

		lgr.Info("===== Final check =====")
		paymentClient, err := isupipe.NewClient(
			agent.WithBaseURL(config.TargetBaseURL),
		)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		if err := scenario.FinalcheckScenario(ctx, paymentClient); err != nil {
			return cli.NewExitError(err, 1)
		}

		lgr.Info("===== System errors =====")
		for _, msg := range bencherror.GetFinalErrorMessages() {
			lgr.Warn(msg)
		}

		lgr.Info("===== Calculate final score =====")
		var msgs []string

		score := benchscore.GetFinalScore()
		msgs = append(msgs, fmt.Sprintf("スコア: %d", score))

		profit := benchscore.GetFinalProfit()
		msgs = append(msgs, fmt.Sprintf("売上: %d", profit))

		penalty := bencherror.GetFinalPenalties()
		msgs = append(msgs, fmt.Sprintf("ペナルティ: %+v", penalty))

		lgr.Info(strings.Join(msgs, "\n"))

		return nil
	},
}
