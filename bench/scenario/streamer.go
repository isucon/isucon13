package scenario

import (
	"context"

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

// 人気VTuber同士を衝突させる？
func BasicStreamerHotScenario(
	ctx context.Context,
	streamerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	// FIXME: impl
	return nil
}

func BasicStreamerModerateScenario(
	ctx context.Context,
	streamerPool *isupipe.ClientPool,
) error {
	client, err := streamerPool.Get(ctx)
	if err != nil {
		return err
	}
	defer streamerPool.Put(ctx, client)

	if err := VisitLivestreamAdmin(ctx, client); err != nil {
		return err
	}

	// 自分のライブ配信一覧取得
	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	for _, livestream := range livestreams {
		reports, err := client.GetLivecommentReports(ctx, livestream.ID)
		if err != nil {
			return err
		}

		for _, report := range reports {
			livestreamID := report.Livecomment.Livestream.ID
			ngWord := scheduler.LivecommentScheduler.GetDummyNgWord()
			if err := client.Moderate(ctx, livestreamID, ngWord.Word); err != nil {
				continue
			}
			scheduler.LivecommentScheduler.Moderate(report.Livecomment.Comment)
		}
	}

	return err
}

// 攻め気にmoderateを行う配信者シナリオ
// スパムが怖くて焦りからこのような行動に出る
// インターネット上のデマ記事を鵜呑みにしてしまう
// 投機的なので、間違った単語を突っ込んだりする
func AggressiveStreamerModerateScenario(
	ctx context.Context,
) error {
	return nil
}
