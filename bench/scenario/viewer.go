package scenario

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

var basicViewerScenarioRandSource = rand.New(rand.NewSource(63877281473681))

func BasicViewerScenario(
	ctx context.Context,
	contestantLogger *zap.Logger,
	viewerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	lgr := zap.S()
	n := basicViewerScenarioRandSource.Int()

	lgr.Info("basic viewer scenario")
	client, err := viewerPool.Get(ctx)
	if err != nil {
		lgr.Warnf("view: failed to get viewer from pool: %s\n", err.Error())
		return err
	}
	defer viewerPool.Put(ctx, client)

	username, err := client.Username()
	if err != nil {
		lgr.Warnf("view: failed to get client username: %s\n", err.Error())
	}

	// NOTE: 配信リンクを直に叩いて視聴開始する人が一定数いる
	lgr.Info("visit top")
	if n%10 == 0 {
		if err := VisitTop(ctx, contestantLogger, client); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			lgr.Warnf("view: failed to visit top page: %s\n", err.Error())
			return err
		}
	}

	lgr.Info("get livestream")
	livestream, err := livestreamPool.Get(ctx)
	if err != nil {
		lgr.Warnf("view: failed to get livestream from pool: %s\n", err.Error())
		return err
	}
	defer livestreamPool.Put(ctx, livestream)

	// NOTE: 配信者のプロフィールが気になる人が一定数いる
	if n%10 == 0 {
		contestantLogger.Info("視聴者が配信者のプロフィールに関心を持ち、訪問しようとしています", zap.String("viewer", username), zap.String("streamer", livestream.Owner.Name))
		lgr.Info("visit user profile")
		if err := VisitUserProfile(ctx, contestantLogger, client, &livestream.Owner); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			lgr.Warnf("view: failed to visit user profile: %s\n", err.Error())
			return err
		}
	}

	lgr.Info("visit livestream")
	if err := VisitLivestream(ctx, contestantLogger, client, livestream); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
		lgr.Warnf("view: failed to visit livestream: %s\n", err.Error())
		return err
	}

	lgr.Info("get livestream stats")
	if _, err := client.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
		lgr.Warnf("view: failed to get livestream stats: %s\n", err.Error())
		return err
	}

	contestantLogger.Info("視聴を開始しました", zap.String("username", username), zap.Int("duration_hours", livestream.Hours()))
	for hour := 1; hour <= livestream.Hours(); hour++ {
		if _, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			lgr.Warnf("view: failed to get livecomments: %s\n", err.Error())
			continue
		}

		livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
		tip, err := scheduler.LivecommentScheduler.GetTipsForStream(livestream.Hours(), hour)
		if err != nil {
			lgr.Warnf("view: failed to get tips for stream: %s\n", err.Error())
			return err
		}
		if _, _, err := client.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecomment.Comment, tip); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			// FIXME: 離脱関連のハンドリング
			lgr.Warnf("view: failed to post livecomment: %s\n", err.Error())
			return err
		}

		if _, err := client.GetReactions(ctx, livestream.ID, livestream.Owner.Name); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			lgr.Warnf("view: failed to get reactions: %s\n", err.Error())
			continue
		}

		emojiName := scheduler.GetReaction()
		if _, err := client.PostReaction(ctx, livestream.ID, livestream.Owner.Name, &isupipe.PostReactionRequest{
			EmojiName: emojiName,
		}); err != nil {
			lgr.Warnf("view: failed to post reactions: %s\n", err.Error())
			continue
		}
	}

	if err := LeaveFromLivestream(ctx, contestantLogger, client, livestream); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
		lgr.Warnf("view: failed to leave from livestream: %s\n", err.Error())
		return err
	}

	return nil
}

func ViewerSpamScenario(
	ctx context.Context,
	contestantLogger *zap.Logger,
	clientPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
	livecommentPool *isupipe.LivecommentPool,
) error {
	lgr := zap.S()

	// ここがmoderate数を左右する
	// pubsubに供給できるように、スパムをたくさん投げる
	viewer, err := clientPool.Get(ctx)
	if err != nil {
		lgr.Warnf("viewer_spam: failed to get client from pool: %s\n", err.Error())
		return err
	}
	defer clientPool.Put(ctx, viewer)

	livestream, err := livestreamPool.Get(ctx)
	if err != nil {
		lgr.Warnf("viewer_spam: failed to get livesteram from pool: %s\n", err.Error())
		return err
	}
	livestreamPool.Put(ctx, livestream) // 他の視聴者、スパム投稿者が入れるようにプールにすぐ戻す

	comment, isModerated := scheduler.LivecommentScheduler.GetNegativeComment()
	if isModerated {
		_, _, err := viewer.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, comment.Comment, &scheduler.Tip{}, isupipe.WithStatusCode(http.StatusBadRequest))
		if err != nil {
			lgr.Warnf("viewer_spam: failed to post livecomment (moderated spam): %s\n", err.Error())
			return err
		}
	} else {
		resp, _, err := viewer.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, comment.Comment, &scheduler.Tip{})
		if err != nil {
			lgr.Warnf("viewer_spam: failed to post livecomment (non-moderated spam): %s\n", err.Error())
			return err
		}

		livecommentPool.Put(ctx, &isupipe.Livecomment{
			ID:         resp.ID,
			User:       resp.User,
			Livestream: resp.Livestream,
			Comment:    resp.Comment,
			Tip:        int(resp.Tip),
			CreatedAt:  int(resp.CreatedAt),
		})
	}

	return nil
}

func BasicViewerReportScenario(
	ctx context.Context,
	contestantLogger *zap.Logger,
	clientPool *isupipe.ClientPool,
	livecommentPool *isupipe.LivecommentPool,
) error {
	viewer, err := clientPool.Get(ctx)
	if err != nil {
		return err
	}
	defer clientPool.Put(ctx, viewer)

	spam, err := livecommentPool.Get(ctx)
	if err != nil {
		return err
	}
	defer livecommentPool.Put(ctx, spam)

	if err := viewer.ReportLivecomment(ctx, spam.Livestream.ID, spam.Livestream.Owner.Name, spam.ID); err != nil {
		return err
	}

	return nil
}
