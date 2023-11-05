package scenario

import (
	"context"

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

	if err := client.EnterLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	for i := 0; i < int(config.AdvertiseCost*10); i++ {
		livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
		tip := scheduler.LivecommentScheduler.GetTipsForStream()
		if _, err := client.PostLivecomment(ctx, livestream.ID, livecomment.Comment, tip); err != nil {
			return err
		}

		emojiName := scheduler.GetReaction()
		if _, err := client.PostReaction(ctx, livestream.ID, &isupipe.PostReactionRequest{
			EmojiName: emojiName,
		}); err != nil {
			return err
		}
	}

	if err := client.ExitLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	return nil
}
