package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	benchmarker := newBenchmarker()

	benchCtx, cancel := context.WithTimeout(ctx, time.Second*defaultBenchmarkerTimeout)
	defer cancel()

	benchmarker.run(benchCtx, client)

	fmt.Printf("final score ==> %d\n", benchscore.GetFinalScore())
}
