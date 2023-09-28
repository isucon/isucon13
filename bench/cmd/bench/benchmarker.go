package main

import (
	"context"
	"fmt"
	"time"

	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/scenario"
)

type benchmarker struct{}

func newBenchmarker() *benchmarker {
	return &benchmarker{}
}

const Season1PassConditionTips = 200000
const Season2PassConditionTips = 400000

// season1 はSeason1シナリオを実行する
// ctx には、context.WithTimeout()でタイムアウトが設定されたものが渡されることを想定
// 内部でSeason1条件が達成されたかどうかをチェックし、問題がなければnilが返る
func (b *benchmarker) season1(ctx context.Context, webappURL string) error {
	go scenario.Season1(ctx, webappURL)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("benchmarker timeout")
		default:
			currentProfit := benchscore.GetCurrentProfit()
			if currentProfit >= Season1PassConditionTips {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// season2 はSeason2シナリオを実行する
// ctx には、context.WithTimeout()でタイムアウトが設定されたものが渡されることを想定
// 内部でSeason2条件が達成されたかどうかをチェックし、問題がなければnilが返る
func (b *benchmarker) season2(ctx context.Context, webappURL string) error {
	go scenario.Season2(ctx, webappURL)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("benchmarker timeout")
		default:
			currentReservationCount := benchscore.GetCurrentScoreByTag(benchscore.SuccessReserveLivestream)
			if currentReservationCount < int64(len(scheduler.Season2LivestreamReservationPatterns)) {
				time.Sleep(1 * time.Second)
				continue
			}

			currentProfit := benchscore.GetCurrentProfit()
			if currentProfit >= Season2PassConditionTips {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}
