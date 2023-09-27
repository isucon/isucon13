package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
)

const (
	defaultBenchmarkerTimeout = 5 // seconds
)

func main() {
	ctx := context.Background()

	client, err := isupipe.NewClient()
	if err != nil {
		log.Fatalln(err)
	}

	if err := scenario.Pretest(ctx, client); err != nil {
		log.Fatalf("Pretest: ベンチマーカの初期テストに失敗しました: %s", err.Error())
	}

	benchscore.InitScore(ctx)
	bencherror.InitializeErrors(ctx)
	benchmarker := newBenchmarker()

	benchCtx, cancel := context.WithTimeout(ctx, time.Second*defaultBenchmarkerTimeout)
	defer cancel()

	// season1シナリオが無事達成された際には、一度cancel()を実行してこれ以上season1シナリオが進行しないようにするため、
	// seasonごとにctxを切る
	season1Ctx, season1Cancel := context.WithCancel(benchCtx)
	defer season1Cancel()

	log.Println("Season1シナリオ走行開始")
	if err := benchmarker.season1(season1Ctx, isupipe.DefaultClientBaseURL); err != nil {
		// 単なるエラーではなく、season1を達成できずに終わっただけなので、スコアを表示する
		log.Println("Season1シナリオの達成条件を満たせませんでした")
		printBenchmarkResult()
		os.Exit(0)
	}
	log.Println("Season1シナリオの達成条件: 200,000の利益を獲得を満たしました")
	season1Cancel()

	printBenchmarkResult()
}

func printBenchmarkResult() {
	criticalErrors, ok := bencherror.GetFinalErrorMessages()[bencherror.BenchmarkCriticalError.ErrorCode()]
	if ok && len(criticalErrors) == 0 {
		for i, c := range criticalErrors {
			log.Printf("critical-error[%d]: %s\n", i, c)
		}

		log.Println("final score ==> 0 (denied)")
		os.Exit(0)
	}

	for key, messages := range bencherror.GetFinalErrorMessages() {
		if key == bencherror.BenchmarkCriticalError.ErrorCode() {
			continue
		}

		for i, message := range messages {
			log.Printf("%s[%d]: %s\n", key, i, message)
		}
	}

	finalScore := benchscore.GetFinalScore()
	finalProfit := benchscore.GetFinalProfit()
	finalPenalty := benchscore.GetFinalPenalty()

	if finalScore+finalProfit < finalPenalty {
		log.Println("final score ==> 0")
	} else {
		log.Printf("final score ==> %d\n", finalScore+finalProfit-finalPenalty)
	}
}
