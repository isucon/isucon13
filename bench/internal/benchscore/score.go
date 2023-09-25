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
	SuccessPostReaction score.ScoreTag = "success-post-reaction"
)

var (
	benchScore *score.Score
	// initOnce   sync.Once
	doneOnce sync.Once
)

func InitScore(ctx context.Context) {
	benchScore = score.NewScore(ctx)

	// FIXME: スコアの重み付けは後ほど考える
	// 登録、ログインは１点
	benchScore.Set(SuccessRegister, 1)
	benchScore.Set(SuccessLogin, 1)

	benchScore.Set(SuccessPostSuperchat, 1)
	benchScore.Set(SuccessPostReaction, 1)

	initProfit(ctx)
}

func AddScore(tag score.ScoreTag) {
	benchScore.Add(tag)
}

func GetCurrentScore() int64 {
	return benchScore.Sum()
}

func GetFinalScore() int64 {
	doneOnce.Do(func() {
		benchScore.Done()
	})
	return benchScore.Sum()
}
