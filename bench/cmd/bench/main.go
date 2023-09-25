package main

import (
	"context"
	"fmt"
	"log"
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
		log.Fatalln(err)
	}

	benchscore.InitScore(ctx)
	bencherror.InitializeErrors(ctx)
	benchmarker := newBenchmarker()

	benchCtx, cancel := context.WithTimeout(ctx, time.Second*defaultBenchmarkerTimeout)
	defer cancel()

	benchmarker.run(benchCtx, client)

	criticalErrors := bencherror.GetFinalErrorMessages()[bencherror.BenchmarkCriticalError.ErrorCode()]
	if len(criticalErrors) == 0 {
		for i, c := range criticalErrors {
			log.Printf("critical-error[%d]: %s\n", i, c)
		}

		log.Fatalln("final score ==> 0")
	}

	finalPenalty := 0
	for key, count := range bencherror.GetFinalPenalties() {
		if key == bencherror.BenchmarkCriticalError.ErrorCode() {
			continue
		}

		penalty := bencherror.PenaltyWeights[key]
		finalPenalty += penalty * int(count)
	}

	finalScore := int(benchscore.GetFinalScore())
	if finalScore < finalPenalty {
		fmt.Println("final score ==> 0")
	} else {
		fmt.Printf("final score ==> %d\n", finalScore-finalPenalty)
	}
}
