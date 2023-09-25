package benchscore

import (
	"context"
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

var (
	benchScore *score.Score
	initOnce   sync.Once
	doneOnce   sync.Once
)

func InitScore(ctx context.Context) {
	initOnce.Do(func() {
		benchScore = score.NewScore(ctx)

		// FIXME: スコアの重み付けは後ほど考える
		// 登録、ログインは１点
		benchScore.Set(SuccessRegister, 1)
		benchScore.Set(SuccessLogin, 1)

		benchScore.Set(SuccessPostSuperchat, 1)

		initProfit(ctx)
	})
}

func AddScore(tag score.ScoreTag) {
	benchScore.Add(tag)
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
