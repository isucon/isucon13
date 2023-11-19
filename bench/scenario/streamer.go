package scenario

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
)

// 枠数1のタイミングで、複数クライアントから一斉に書き込み、１個だけ成立しない場合は失格判定
func BasicLongStreamerScenario(
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
	streamerPool.Put(ctx, client) // 他のviewerが参入できるようにプールにすぐもどす

	username, err := client.Username()
	if err != nil {
		return err
	}

	reservation, err := scheduler.ReservationSched.GetColdShortReservation()
	if err != nil {
		return err
	}

	tags, err := client.GetRandomLivestreamTags(ctx, 5)
	if err != nil {
		return err
	}

	log.Printf("before reserve livestream start=%s, end=%s\n", time.Unix(reservation.StartAt, 0).String(), time.Unix(reservation.EndAt, 0).String())
	livestream, err := client.ReserveLivestream(ctx, username, &isupipe.ReserveLivestreamRequest{
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
	log.Printf("after reserve livestream start=%s, end=%s\n", time.Unix(livestream.StartAt, 0).String(), time.Unix(livestream.EndAt, 0).String())
	scheduler.ReservationSched.CommitReservation(reservation)

	if scheduler.UserScheduler.IsPopularStreamer(username) {
		popularLivestreamPool.Put(ctx, livestream)
	} else {
		livestreamPool.Put(ctx, livestream)
	}

	return nil
}

// 人気VTuber同士を衝突させる？
func BasicLongStreamerHotScenario(
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

	// 自分のライブ配信一覧取得
	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	for _, livestream := range livestreams {
		if err := VisitLivestreamAdmin(ctx, client, livestream); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			return err
		}

		reports, err := client.GetLivecommentReports(ctx, livestream.ID)
		if err != nil {
			continue
		}

		for _, report := range reports {
			livestreamID := report.Livecomment.Livestream.ID
			ngword, err := scheduler.LivecommentScheduler.GetNgWord(report.Livecomment.Comment)
			if err != nil {
				return err
			}
			if err := client.Moderate(ctx, livestreamID, ngword); err != nil {
				continue
			}
			scheduler.LivecommentScheduler.Moderate(report.Livecomment.Comment)
		}
	}

	return err
}

// 攻め気にmoderateを行う配信者シナリオ
// 基本的なmoderateの流れから外れており、livecomment_reportsに存在しないNGワードを入れようとするので
// ng_wordsテーブルが嵩む要因になる
// livecomment_reportsを見てmoderateを弾く実装は初期ではないので、ng_wordsが異常に嵩んだり重複レコードが多かったりすることに気づいたタイミングで
// 対策を打ってもらう想定
func AggressiveStreamerModerateScenario(
	ctx context.Context,
	streamerPool *isupipe.ClientPool,
) error {
	client, err := streamerPool.Get(ctx)
	if err != nil {
		return err
	}
	defer streamerPool.Put(ctx, client)

	// 自分のライブ配信一覧取得
	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	for _, livestream := range livestreams {
		if err := VisitLivestreamAdmin(ctx, client, livestream); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			return err
		}

		ngWord := scheduler.LivecommentScheduler.GetDummyNgWord()
		if err := client.Moderate(ctx, livestream.ID, ngWord.Word); err != nil {
			continue
		}
		scheduler.LivecommentScheduler.ModerateNgWord(ngWord.Word)
	}

	return nil
}
