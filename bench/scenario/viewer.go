package scenario

import (
	"context"
	"errors"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

func BasicViewerScenario(
	ctx context.Context,
	viewerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	lgr := zap.S()

	client, err := viewerPool.Get(ctx)
	if err != nil {
		return err
	}
	defer viewerPool.Put(ctx, client)

	if err := VisitTop(ctx, client); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
		return err
	}

	livestream, err := livestreamPool.Get(ctx)
	if err != nil {
		return err
	}
	defer livestreamPool.Put(ctx, livestream)

	if err := VisitLivestream(ctx, client, livestream); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
		return err
	}

	for hour := 0; hour < livestream.Hours(); hour++ {
		if _, err := client.GetLivestreamStatistics(ctx, livestream.ID); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			continue
		}

		if _, err := client.GetLivecomments(ctx, livestream.ID); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			continue
		}

		livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
		tip := scheduler.LivecommentScheduler.GetTipsForStream(livestream.Hours(), hour)
		if _, _, err := client.PostLivecomment(ctx, livestream.ID, livecomment.Comment, tip); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			// FIXME: 真面目にログを書く
			lgr.Info("離脱: %s", err.Error())
			return err
		}

		if _, err := client.GetReactions(ctx, livestream.ID); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			continue
		}

		emojiName := scheduler.GetReaction()
		if _, err := client.PostReaction(ctx, livestream.ID, &isupipe.PostReactionRequest{
			EmojiName: emojiName,
		}); err != nil {
			continue
		}
	}

	if err := LeaveFromLivestream(ctx, client, livestream); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
		return err
	}

	return nil
}

func ViewerSpamScenario(
	ctx context.Context,
	clientPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
	livecommentPool *isupipe.LivecommentPool,
) error {
	// ここがmoderate数を左右する
	// pubsubに供給できるように、スパムをたくさん投げる
	viewer, err := clientPool.Get(ctx)
	if err != nil {
		return err
	}
	defer clientPool.Put(ctx, viewer)

	livestream, err := livestreamPool.Get(ctx)
	if err != nil {
		return err
	}
	livestreamPool.Put(ctx, livestream) // 他の視聴者、スパム投稿者が入れるようにプールにすぐ戻す

	comment, isModerated := scheduler.LivecommentScheduler.GetNegativeComment()
	if isModerated {
		_, _, err := viewer.PostLivecomment(ctx, livestream.ID, comment.Comment, &scheduler.Tip{}, isupipe.WithStatusCode(http.StatusBadRequest))
		if err != nil {
			return err
		}
	} else {
		resp, _, err := viewer.PostLivecomment(ctx, livestream.ID, comment.Comment, &scheduler.Tip{})
		if err != nil {
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

	if err := viewer.ReportLivecomment(ctx, spam.Livestream.ID, spam.ID); err != nil {
		return err
	}

	return nil
}
