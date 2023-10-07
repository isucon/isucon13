package main

import (
	"context"
	"sync"

	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/scenario"
	"go.uber.org/zap"
)

type benchmarker struct{}

func newBenchmarker() *benchmarker {
	return &benchmarker{}
}

const Season1PassConditionTips = 70000
const Season2PassConditionTips = 90000
const Season3PassConditionTips = 110000

// season1 はSeason1シナリオを実行する
// ctx には、context.WithTimeout()でタイムアウトが設定されたものが渡されることを想定
// 内部でSeason1条件が達成されたかどうかをチェックし、問題がなければnilが返る
func (b *benchmarker) runSeason1(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)

	lgr := zap.S()

	benchscore.SetAchivementGoal(Season1PassConditionTips)
	lgr.Infof("シーズン1の達成条件は以下のとおりです")
	lgr.Infof("投げ銭売上: %d", Season1PassConditionTips)

	// NOTE: config.TargetBaseURLがあるので、いちいち引数で引き回さなくて良い
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scenario.Phase1(ctx)
	}()

	// 時間切れか達成条件を満たすまで待ち、goroutineの終了を待ち合わせる
	for {
		select {
		case <-ctx.Done():
			cancel()
			wg.Wait()
			return ctx.Err()
		case <-benchscore.Achieve():
			// 目標達成
			cancel()
			wg.Wait()
			return nil
		}
	}
}

func (b *benchmarker) runSeason2(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)

	lgr := zap.S()

	benchscore.SetAchivementGoal(Season2PassConditionTips)
	lgr.Infof("シーズン2の達成条件は以下のとおりです")
	lgr.Infof("投げ銭売上: %d", Season2PassConditionTips)

	// NOTE: config.TargetBaseURLがあるので、いちいち引数で引き回さなくて良い
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scenario.Phase2(ctx)
	}()

	// 時間切れか達成条件を満たすまで待ち、goroutineの終了を待ち合わせる
	for {
		select {
		case <-ctx.Done():
			cancel()
			wg.Wait()
			return ctx.Err()
		case <-benchscore.Achieve():
			// 目標達成
			cancel()
			wg.Wait()
			return nil
		}
	}
}

func (b *benchmarker) runSeason3(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)

	lgr := zap.S()

	benchscore.SetAchivementGoal(Season3PassConditionTips)
	lgr.Infof("シーズン3の達成条件は以下のとおりです")
	lgr.Infof("投げ銭売上: %d", Season3PassConditionTips)

	// NOTE: config.TargetBaseURLがあるので、いちいち引数で引き回さなくて良い
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scenario.Phase3(ctx)
	}()

	// 時間切れか達成条件を満たすまで待ち、goroutineの終了を待ち合わせる
	for {
		select {
		case <-ctx.Done():
			cancel()
			wg.Wait()
			return ctx.Err()
		case <-benchscore.Achieve():
			// 目標達成
			cancel()
			wg.Wait()
			return nil
		}
	}
}

func (b *benchmarker) runSeason4(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)

	lgr := zap.S()

	benchscore.SetAchivementGoal(Season3PassConditionTips)
	lgr.Infof("シーズン4の達成条件はありません")

	// NOTE: config.TargetBaseURLがあるので、いちいち引数で引き回さなくて良い
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scenario.Phase4(ctx)
	}()

	// 時間切れか達成条件を満たすまで待ち、goroutineの終了を待ち合わせる
	for {
		select {
		case <-ctx.Done():
			cancel()
			wg.Wait()
			return ctx.Err()
		case <-benchscore.Achieve():
			// 目標達成
			cancel()
			wg.Wait()
			return nil
		}
	}
}
