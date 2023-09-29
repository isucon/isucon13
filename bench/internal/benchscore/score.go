package benchscore

import (
	"context"
	"sync"

	"github.com/isucon/isucandar/score"
)

type ScoreTag string

const (
	SuccessGetTags score.ScoreTag = "success-get-tags"
	// ユーザ
	SuccessRegister     score.ScoreTag = "success-register"
	SuccessLogin        score.ScoreTag = "success-login"
	SuccessGetUser      score.ScoreTag = "success-get-user"
	SuccessGetUserTheme score.ScoreTag = "success-get-user-theme"
	// ライブ配信
	SuccessReserveLivestream  score.ScoreTag = "success-reserve-livestream"
	SuccessGetLivestream      score.ScoreTag = "success-get-livestream"
	SuccessGetLivestreamByTag score.ScoreTag = "success-get-livestream-by-tag"
	// ライブコメント
	SuccessGetLivecomments       score.ScoreTag = "success-get-livecomments"
	SuccessPostLivecomment       score.ScoreTag = "success-post-livecomment"
	SuccessReportLivecomment     score.ScoreTag = "success-report-livecomment"
	SuccessGetLivecommentReports score.ScoreTag = "success-get-livecomment-reports"
	// リアクション
	SuccessGetReactions score.ScoreTag = "success-get-reactions"
	SuccessPostReaction score.ScoreTag = "success-post-reaction"

	SuccessEnterLivestream score.ScoreTag = "success-enter-livestream"
	SuccessLeaveLivestream score.ScoreTag = "success-leave-livestream"
)

var (
	benchScore *score.Score
	// initOnce   sync.Once
	doneOnce sync.Once
)

func InitScore(ctx context.Context) {
	benchScore = score.NewScore(ctx)

	// FIXME: スコアの重み付けは後ほど考える
	benchScore.Set(SuccessGetTags, 1)
	benchScore.Set(SuccessRegister, 1)
	benchScore.Set(SuccessLogin, 1)
	benchScore.Set(SuccessGetUser, 1)
	benchScore.Set(SuccessGetUserTheme, 1)

	benchScore.Set(SuccessReserveLivestream, 1)
	benchScore.Set(SuccessGetLivestreamByTag, 1)

	benchScore.Set(SuccessGetLivecomments, 1)
	benchScore.Set(SuccessPostLivecomment, 1)
	benchScore.Set(SuccessReportLivecomment, 1)

	benchScore.Set(SuccessGetReactions, 1)
	benchScore.Set(SuccessPostReaction, 1)

	benchScore.Set(SuccessEnterLivestream, 1)
	benchScore.Set(SuccessLeaveLivestream, 1)

	initProfit(ctx)
}

func AddScore(tag score.ScoreTag) {
	benchScore.Add(tag)
}

func GetCurrentScore() int64 {
	return benchScore.Sum()
}

func GetCurrentScoreByTag(tag score.ScoreTag) int64 {
	return benchScore.Breakdown()[tag]
}

func GetFinalScore() int64 {
	doneOnce.Do(func() {
		benchScore.Done()
	})
	return benchScore.Sum()
}
