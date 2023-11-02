package scenario

import (
	"context"

	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/vendor/go.uber.org/zap"
)

func BasicViewerScenario(
	ctx context.Context,
	viewerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	lgr := zap.S()

	client, err := viewerPool.Get(ctx)
	if err != nil {
		lgr.Info("failed to get client from viewer pool", zap.Error(err))
		return err
	}
	defer viewerPool.Put(ctx, client)

	if err := VisitTop(ctx, client); err != nil {
		lgr.Info("failed to visit top", zap.Error(err))
		return err
	}

	livestream, err := livestreamPool.Get(ctx)
	if err != nil {
		lgr.Info("failed to get livestream from pool(as viewer)", zap.Error(err))
		return err
	}
	defer livestreamPool.Put(ctx, livestream)

	if err := VisitLivestream(ctx, client, livestream); err != nil {
		lgr.Info("failed to visit livestream", zap.Error(err))
		return err
	}

	if err := client.EnterLivestream(ctx, livestream.ID); err != nil {
		lgr.Info("failed to enter livestream", zap.Error(err))
		return err
	}

	for i := 0; i < config.AdvertiseCost*10; i++ {
		livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
		tip := scheduler.LivecommentScheduler.GetTipsForStream()
		if _, err := client.PostLivecomment(ctx, livestream.ID, &isupipe.PostLivecommentRequest{
			Comment: livecomment.Comment,
			Tip:     tip,
		}); err != nil {
			return err
		}

		emojiName := scheduler.GetReaction()
		if _, err := client.PostReaction(ctx, livestream.ID, &isupipe.PostReactionRequest{
			EmojiName: emojiName,
		}); err != nil {
			return err
		}
	}

	if err := client.LeaveLivestream(ctx, livestream.ID); err != nil {
		lgr.Info("failed to enter livestream", zap.Error(err))
		return err
	}

	return nil
}
