package scenario

import (
	"context"

	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
)

func BasicPopularStreamerScenario(
	ctx context.Context,
	popularStreamerPool *isupipe.ClientPool,
) error {
	// FIXME: impl
	return nil
}

func BasicStreamerColdReserveScenario(
	ctx context.Context,
	streamerPool *isupipe.ClientPool,
	popularLivestreamPool *isupipe.LivestreamPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	client, err := streamerPool.Get(ctx)
	if err != nil {
		return err
	}
	defer streamerPool.Put(ctx, client) // 使い終わったらお片付け

	username, err := client.LoginUserName()
	if err != nil {
		return err
	}

	if err := VisitLivestreamAdmin(ctx, client); err != nil {
		return err
	}

	reservation, err := scheduler.ReservationSched.GetColdReservation()
	if err != nil {
		return err
	}

	tags, err := client.GetRandomTags(ctx, 5)
	if err != nil {
		return err
	}

	livestream, err := client.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		Tags:        tags,
		Title:       reservation.Title,
		Description: reservation.Description,
		StartAt:     reservation.StartAt,
		EndAt:       reservation.EndAt,
	})
	if err != nil {
		scheduler.ReservationSched.AbortReservation(reservation)
		return err
	}
	scheduler.ReservationSched.CommitReservation(reservation)

	if scheduler.UserScheduler.IsPopularStreamer(username) {
		popularLivestreamPool.Put(ctx, livestream)
	} else {
		livestreamPool.Put(ctx, livestream)
	}

	return nil
}

func StreamerHotScenario(
	ctx context.Context,
	streamerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	// FIXME: impl
	return nil
}

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

	if err := client.EnterLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	// FIXME: とりあえず固定値でやってるが、広告費用係数合わせる
	for i := 0; i < config.AdvertiseCost*10; i++ {
		livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
		tip := scheduler.LivecommentScheduler.GetTipsForStream()
		if _, err := client.PostLivecomment(ctx, livestream.Id, &isupipe.PostLivecommentRequest{
			Comment: livecomment.Comment,
			Tip:     tip,
		}); err != nil {
			return err
		}

		emojiName := scheduler.GetReaction()
		if _, err := client.PostReaction(ctx, livestream.Id, &isupipe.PostReactionRequest{
			EmojiName: emojiName,
		}); err != nil {
			return err
		}
	}

	if err := client.LeaveLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	return nil
}
