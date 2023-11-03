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
	profit    *score.Score
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

func AddTipProfit(tip int64) error {
	tag := tipToProfitLevel(tip)
	if tag == TipProfitLevel0 {
		return nil
	}

	profit.Add(tag)

	return nil
}

// FIXME: チップ額がフロントエンドから見て違和感ないように
func tipToProfitLevel(tip int64) score.ScoreTag {
	switch tip {
	case 0:
		return TipProfitLevel0
	case 1:
		return TipProfitLevel1
	case 2:
		return TipProfitLevel2
	case 3:
		return TipProfitLevel3
	case 4:
		return TipProfitLevel4
	case 5:
		return TipProfitLevel5
	default:
		return TipProfitLevel0
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
