package benchscore

import (
	"context"
	"sync"

	"github.com/isucon/isucandar/score"
)

const (
	TipProfitLevel0 score.ScoreTag = "tip-level0" // 投げ銭が含まれないライブコメント
	TipProfitLevel1 score.ScoreTag = "tip-level1"
	TipProfitLevel2 score.ScoreTag = "tip-level2"
	TipProfitLevel3 score.ScoreTag = "tip-level3"
	TipProfitLevel4 score.ScoreTag = "tip-level4"
	TipProfitLevel5 score.ScoreTag = "tip-level5"
)

var (
	profit *score.Score

	achieveCh chan struct{}
	goalSum   int
	closeOnce sync.Once
)

func initProfit(ctx context.Context) {
	profit = score.NewScore(ctx)
	profit.Set(TipProfitLevel1, 1)
	profit.Set(TipProfitLevel2, 2)
	profit.Set(TipProfitLevel3, 3)
	profit.Set(TipProfitLevel4, 4)
	profit.Set(TipProfitLevel5, 5)
}

func SetAchivementGoal(goal int) {
	achieveCh = make(chan struct{})
	goalSum = goal
}

func Achieve() chan struct{} {
	return achieveCh
}

func AddTipProfit(tip int) error {
	tag := tipToProfitLevel(tip)
	if tag == TipProfitLevel0 {
		return nil
	}

	profit.Add(tag)
	if profit.Sum() >= int64(goalSum) {
		closeOnce.Do(func() {
			close(achieveCh)
		})
	}

	return nil
}

func tipToProfitLevel(tip int) score.ScoreTag {
	if tip == 0 {
		return TipProfitLevel0
	} else if tip >= 1 && tip <= 500 {
		return TipProfitLevel1
	} else if tip >= 500 && tip < 1000 {
		return TipProfitLevel2
	} else if tip >= 1000 && tip < 5000 {
		return TipProfitLevel3
	} else if tip >= 5000 && tip < 10000 {
		return TipProfitLevel4
	} else if tip >= 10000 && tip <= 20000 {
		return TipProfitLevel5
	} else {
		// APIサーバが正しくtipsの下限と上限をバリデーションできているかチェックするロジックはここではない
		panic("UNREACHABLE: uncovered tip value specified")
	}
}

// GetFinalProfit は、最終売上を返します
// FIXME: finalcheck後にprofitをスコアに加算しないと駄目
func GetFinalProfit() int64 {
	doneOnce.Do(func() {
		profit.Done()
	})
	return profit.Sum()
}

func GetCurrentProfit() int64 {
	return profit.Sum()
}
