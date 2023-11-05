package benchscore

import (
	"context"
	"sync"

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
	profit         *score.Score
	doneProfitOnce sync.Once
)

func InitProfit(ctx context.Context) {
	profit = score.NewScore(ctx)
	profit.Set(TipProfitLevel1, 1)
	profit.Set(TipProfitLevel2, 2)
	profit.Set(TipProfitLevel3, 3)
	profit.Set(TipProfitLevel4, 4)
	profit.Set(TipProfitLevel5, 5)
}

func AddTipLevel(level int64) {
	switch level {
	case 1:
		profit.Add(TipProfitLevel1)
	case 2:
		profit.Add(TipProfitLevel2)
	case 3:
		profit.Add(TipProfitLevel3)
	case 4:
		profit.Add(TipProfitLevel4)
	case 5:
		profit.Add(TipProfitLevel5)
	}
}

// GetFinalProfit は、最終売上を返します
// FIXME: finalcheck後にprofitをスコアに加算しないと駄目
func GetTotalProfit() int64 {
	return profit.Sum()
}

func DoneProfit() {
	doneProfitOnce.Do(func() {
		profit.Done()
	})
}
