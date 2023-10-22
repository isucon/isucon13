package main

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type benchmarker struct {
	sem *semaphore.Weighted
}

func newBenchmarker() *benchmarker {
	lgr := zap.S()

	// FIXME: 広告費用から重さを計算する
	// いったん固定値で設定しておく
	var weight int64 = 10
	lgr.Infof("負荷レベル: %d", weight)

	return &benchmarker{sem: semaphore.NewWeighted(weight)}
}

func (b *benchmarker) load(ctx context.Context) error {
	defer b.sem.Release(1)

	// FIXME: impl

	return nil
}

func (b *benchmarker) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if ok := b.sem.TryAcquire(1); ok {
				go b.load(ctx)
			}
		}
	}
}
