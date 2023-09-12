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
	// リアクション
)

var (
	benchScore *score.Score
	doneOnce   sync.Once
)

func InitScore(ctx context.Context) (*score.Score, error) {
	if benchScore != nil {
		return nil, fmt.Errorf("benchmark score is already set")
	}
	benchScore = score.NewScore(ctx)

	// 登録、ログインは１点
	benchScore.Set(SuccessRegister, 1)
	benchScore.Set(SuccessLogin, 1)

	return benchScore, nil
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
