package benchscore

import (
	"context"
	"fmt"

	"github.com/isucon/isucandar/score"
)

const (
	TipProfitLevel1 score.ScoreTag = "tip-level1"
	TipProfitLevel2 score.ScoreTag = "tip-level2"
	TipProfitLevel3 score.ScoreTag = "tip-level3"
	TipProfitLevel4 score.ScoreTag = "tip-level4"
	TipProfitLevel5 score.ScoreTag = "tip-level5"
)

var (
	profit *score.Score
)

func initProfit(ctx context.Context) {
	profit = score.NewScore(ctx)
	profit.Set(TipProfitLevel1, 1)
	profit.Set(TipProfitLevel2, 2)
	profit.Set(TipProfitLevel3, 3)
	profit.Set(TipProfitLevel4, 4)
	profit.Set(TipProfitLevel5, 5)
}

func AddTipProfit(tip int) error {
	tag, err := tipToProfitLevel(tip)
	if err != nil {
		return err
	}
	profit.Add(tag)
	return nil
}

func tipToProfitLevel(tip int) (score.ScoreTag, error) {
	if tip >= 1 && tip <= 500 {
		return TipProfitLevel1, nil
	} else if tip >= 500 && tip < 1000 {
		return TipProfitLevel2, nil
	} else if tip >= 1000 && tip < 5000 {
		return TipProfitLevel3, nil
	} else if tip >= 5000 && tip < 10000 {
		return TipProfitLevel4, nil
	} else if tip >= 10000 && tip < 50000 {
		// FIXME: 50000という上限は、APIサーバ側で定めた一回のtip上限によって変える
		return TipProfitLevel5, nil
	} else {
		return TipProfitLevel5, fmt.Errorf("uncovered tip value specified: %d", tip)
	}
}
