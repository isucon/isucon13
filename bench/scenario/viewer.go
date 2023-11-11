package scenario

import (
	"context"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
)

func BasicViewerScenario(
	ctx context.Context,
	viewerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {

	client, err := viewerPool.Get(ctx)
	if err != nil {
		return err
	}
	defer viewerPool.Put(ctx, client)

	if err := VisitTop(ctx, client); err != nil {
		return err
	}

	livestream, err := livestreamPool.Get(ctx)
	if err != nil {
		return err
	}
	defer livestreamPool.Put(ctx, livestream)

	if err := VisitLivestream(ctx, client, livestream); err != nil {
		return err
	}

	for i := 0; i < int(config.AdvertiseCost*10); i++ {
		livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
		tip := scheduler.LivecommentScheduler.GetTipsForStream()
		if _, _, err := client.PostLivecomment(ctx, livestream.ID, livecomment.Comment, tip); err != nil {
			return err
		}

		emojiName := scheduler.GetReaction()
		if _, err := client.PostReaction(ctx, livestream.ID, &isupipe.PostReactionRequest{
			EmojiName: emojiName,
		}); err != nil {
			return err
		}
	}

	if err := GoAwayFromLivestream(ctx, client, livestream); err != nil {
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
	defer livestreamPool.Put(ctx, livestream)

	comment := scheduler.LivecommentScheduler.GetNegativeComment()
	resp, _, err := viewer.PostLivecomment(ctx, livestream.ID, comment.Comment, &scheduler.Tip{}, isupipe.WithStatusCode(http.StatusBadRequest))
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
