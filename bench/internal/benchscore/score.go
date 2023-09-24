package benchscore

import (
	"context"
	"fmt"
	"sync"

	"github.com/isucon/isucandar/score"
)

type ScoreTag string

const (
	// ユーザ
	SuccessRegister score.ScoreTag = "success-register"
	SuccessLogin    score.ScoreTag = "success-login"
	// ライブ配信
	// スパチャ
	SuccessPostSuperchat score.ScoreTag = "success-post-superchat"
	// リアクション
)

// FIXME: tipは１つのパッケージにまとめたほうがいいかも
const (
	TipLevel1 score.ScoreTag = "tip-level1"
	TipLevel2 score.ScoreTag = "tip-level2"
	TipLevel3 score.ScoreTag = "tip-level3"
	TipLevel4 score.ScoreTag = "tip-level4"
	TipLevel5 score.ScoreTag = "tip-level5"
)

var (
	benchScore *score.Score
	// FIXME: tipは１つのパッケージにまとめたほうがいいかも
	profit   *score.Score
	initOnce sync.Once
	doneOnce sync.Once
)

func InitScore(ctx context.Context) {
	initOnce.Do(func() {
		benchScore = score.NewScore(ctx)

		// FIXME: スコアの重み付けは後ほど考える
		// 登録、ログインは１点
		benchScore.Set(SuccessRegister, 1)
		benchScore.Set(SuccessLogin, 1)

		benchScore.Set(SuccessPostSuperchat, 1)

		profit = score.NewScore(ctx)
		profit.Set(TipLevel1, 1)
		profit.Set(TipLevel2, 2)
		profit.Set(TipLevel3, 3)
		profit.Set(TipLevel4, 4)
		profit.Set(TipLevel5, 5)
	})
}

func AddScore(tag score.ScoreTag) {
	benchScore.Add(tag)
}

// FIXME: tipは１つのパッケージにまとめたほうがいいかも
func AddTip(tip int) error {
	if tip >= 10 && tip < 100 {
		profit.Add(TipLevel1)
	} else if tip >= 100 && tip < 200 {
		profit.Add(TipLevel2)
	} else if tip >= 200 && tip < 300 {
		profit.Add(TipLevel3)
	} else if tip >= 300 && tip < 400 {
		profit.Add(TipLevel4)
	} else if tip >= 400 && tip < 500 {
		profit.Add(TipLevel5)
	} else {
		return fmt.Errorf("uncovered tip value specified: %d", tip)
	}

	return nil
}

func GetFinalScore() int64 {
	doneOnce.Do(func() {
		benchScore.Done()
	})
	return benchScore.Sum()
}

// GetProfit は、最終売上を返します
// FIXME: finalcheck後にprofitをスコアに加算しないと駄目
func GetProfit() int64 {
	doneOnce.Do(func() {
		profit.Done()
	})
	return profit.Sum()
}
