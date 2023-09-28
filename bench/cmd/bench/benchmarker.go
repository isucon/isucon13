package main

import (
	"context"
	"fmt"

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

func (b *benchmarker) checkSeason1Requirements(ctx context.Context) chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				current := benchscore.GetCurrentProfit()
				if current >= Season1PassConditionTips {
					return
				}
			}
		}
	}()
	return done
}

// season1 はSeason1シナリオを実行する
// ctx には、context.WithTimeout()でタイムアウトが設定されたものが渡されることを想定
// 内部でSeason1条件が達成されたかどうかをチェックし、問題がなければnilが返る
func (b *benchmarker) season1(ctx context.Context, webappURL string) error {
	seasonCtx, cancelSeason := context.WithCancel(ctx)

	go scenario.Season1(ctx, webappURL)

	checkCh := b.checkSeason1Requirements(ctx)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("benchmarker timeout")
		case <-checkCh:
			return nil
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
