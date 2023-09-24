package main

import (
	"context"

	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
)

type benchmarker struct{}

func newBenchmarker() *benchmarker {
	return &benchmarker{}
}

// run はベンチマークシナリオを実行する
// ctx には、context.WithTimeout()でタイムアウトが設定されたものが渡されることを想定
func (b *benchmarker) run(ctx context.Context, client *isupipe.Client) {
	go scenario.Normal(ctx, client)
}
