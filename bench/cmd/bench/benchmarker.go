package main

import (
	"context"

	"github.com/isucon/isucon13/bench/internal/benchscore"
)

type benchmarker struct{}

func newBenchmarker() *benchmarker {
	return &benchmarker{}
}

const Season1PassConditionTips = 200000

func (b *benchmarker) checkSeason1Requirements(ctx context.Context, cancelSeason context.CancelFunc) chan struct{} {
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
// func (b *benchmarker) season1(ctx context.Context, webappURL string) error {
// 	seasonCtx, cancelSeason := context.WithCancel(ctx)

// 	go scenario.Season1(ctx, webappURL)

// 	checkCh := b.checkSeason1Requirements(ctx)
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return fmt.Errorf("benchmarker timeout")
// 		case <-checkCh:
// 			return nil
// 		}
// 	}
// }
