package main

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/scenario"
)

type benchmarker struct{}

func newBenchmarker() *benchmarker {
	return &benchmarker{}
}

const Season1PassConditionTips = 200000

// season1 はSeason1シナリオを実行する
// ctx には、context.WithTimeout()でタイムアウトが設定されたものが渡されることを想定
// 内部でSeason1条件が達成されたかどうかをチェックし、問題がなければnilが返る
func (b *benchmarker) runSeason1(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	benchscore.SetAchivementGoal(Season1PassConditionTips)

	// NOTE: config.TargetBaseURLがあるので、いちいち引数で引き回さなくて良い
	go scenario.Season1(ctx, "")

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("benchmarker timeout")
		case <-benchscore.Achieve():
			// 目標達成
			return nil
		}
	}
}

func (b *benchmarker) runSeason2(ctx context.Context) error {

	return nil
}
