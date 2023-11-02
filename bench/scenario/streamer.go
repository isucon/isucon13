package scenario

import (
	"context"

	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
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
	lgr := zap.S()

	client, err := streamerPool.Get(ctx)
	if err != nil {
		lgr.Info("failed to get client from streamer pool")
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
		Tags:         tags,
		Title:        reservation.Title,
		Description:  reservation.Description,
		PlaylistUrl:  reservation.PlaylistUrl,
		ThumbnailUrl: reservation.ThumbnailUrl,
		StartAt:      reservation.StartAt,
		EndAt:        reservation.EndAt,
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

func BasicStreamerHotScenario(
	ctx context.Context,
	streamerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	// FIXME: impl
	return nil
}
