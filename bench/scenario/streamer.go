package scenario

import (
	"context"
	"errors"
	"math/rand"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

var basicStreamerScenarioRandSource = rand.New(rand.NewSource(18637418277836))

// 枠数1のタイミングで、複数クライアントから一斉に書き込み、１個だけ成立しない場合は失格判定
func BasicLongStreamerScenario(
	ctx context.Context,
	contestantLogger *zap.Logger,
	popularStreamerPool *isupipe.ClientPool,
) error {
	// FIXME: impl
	return nil
}

func BasicStreamerColdReserveScenario(
	ctx context.Context,
	contestantLogger *zap.Logger,
	streamerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	lgr := zap.S()
	n := basicStreamerScenarioRandSource.Int()

	client, err := streamerPool.Get(ctx)
	if err != nil {
		return err
	}
	streamerPool.Put(ctx, client) // 他のviewerが参入できるようにプールにすぐもどす

	username, err := client.Username()
	if err != nil {
		lgr.Warnf("reserve: failed to get username: %s\n", err.Error())
		return err
	}

	if n%10 == 0 { // NOTE: 一定数の配信者がアイコンを変更する
		lgr.Info("change icon")
		randomIcon := scheduler.IconSched.GetRandomIcon()
		if _, err := client.PostIcon(ctx, &isupipe.PostIconRequest{
			Image: randomIcon.Image,
		}); err != nil {
			return err
		}
	}

	var reservation *scheduler.Reservation
	if n%2 == 0 {
		r, err := scheduler.ReservationSched.GetColdShortReservation()
		if err != nil {
			lgr.Warnf("reserve: failed to get cold short reservation: %s\n", err.Error())
			return err
		}
		reservation = r
	} else {
		r, err := scheduler.ReservationSched.GetColdLongReservation()
		if err != nil {
			lgr.Warnf("reserve: failed to get cold long reservation: %s\n", err.Error())
			return err
		}
		reservation = r
	}

	tags, err := client.GetRandomLivestreamTags(ctx, 5)
	if err != nil {
		lgr.Warnf("reserve: failed to get random livestream tags: %s\n", err.Error())
		return err
	}

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
		lgr.Warnf("reserve: failed to reserve: %s\n", err.Error())
		return err
	}
	scheduler.ReservationSched.CommitReservation(reservation)

	livestreamPool.Put(ctx, livestream)
	// ログ削減
	// contestantLogger.Info("配信を予約しました", zap.String("streamer", livestream.Owner.Name), zap.String("title", livestream.Title), zap.Int("duration_hours", livestream.Hours()))

	return nil
}

// 人気VTuber同士を衝突させる？
func BasicLongStreamerHotScenario(
	ctx context.Context,
	contestantLogger *zap.Logger,
	streamerPool *isupipe.ClientPool,
	livestreamPool *isupipe.LivestreamPool,
) error {
	// FIXME: impl
	return nil
}

func BasicStreamerModerateScenario(
	ctx context.Context,
	contestantLogger *zap.Logger,
	streamerPool *isupipe.ClientPool,
) error {
	lgr := zap.S()

	client, err := streamerPool.Get(ctx)
	if err != nil {
		return err
	}
	defer streamerPool.Put(ctx, client)

	// 自分のライブ配信一覧取得
	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		lgr.Warnf("streamer_moderate: failed to get my livestream: %s\n", err.Error())
		return err
	}

	for _, livestream := range livestreams {
		client.GetIcon(ctx, livestream.Owner.Name, isupipe.WithETag(livestream.Owner.IconHash))
		// icon取得のエラーは無視

		if err := VisitLivestreamAdmin(ctx, contestantLogger, client, livestream); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			lgr.Warnf("streamer_moderate: failed to visit livestream admin: %s\n", err.Error())
			return err
		}

		reports, err := client.GetLivecommentReports(ctx, livestream.ID, livestream.Owner.Name)
		if err != nil {
			lgr.Warnf("streamer_moderate: failed to get livecomment reports: %s\n", err.Error())
			continue
		}

		for _, report := range reports {
			client.GetIcon(ctx, report.Livecomment.User.Name, isupipe.WithETag(report.Livecomment.User.IconHash))
			// icon取得のエラーは無視

			livestreamID := report.Livecomment.Livestream.ID
			ngword, err := scheduler.LivecommentScheduler.GetNgWord(report.Livecomment.Comment)
			if err != nil {
				lgr.Warnf("streamer_moderate: failed to get ngwords: %s\n", err.Error())
				return err
			}
			if err := client.Moderate(ctx, livestreamID, livestream.Owner.Name, ngword); err != nil {
				lgr.Warnf("streamer_moderate: failed to moderate: %s\n", err.Error())
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
	contestantLogger *zap.Logger,
	streamerPool *isupipe.ClientPool,
) error {
	lgr := zap.S()

	client, err := streamerPool.Get(ctx)
	if err != nil {
		lgr.Warnf("aggressive streamer moderate: failed to get streamer from pool: %s\n", err.Error())
		return err
	}
	defer streamerPool.Put(ctx, client)

	// 自分のライブ配信一覧取得
	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		lgr.Warnf("aggressive streamer moderate: failed to get my livestream: %s\n", err.Error())
		return err
	}

	for _, livestream := range livestreams {
		if err := VisitLivestreamAdmin(ctx, contestantLogger, client, livestream); err != nil && !errors.Is(err, bencherror.ErrTimeout) {
			lgr.Warnf("aggressive streamer moderate: failed to visit livestream admin: %s\n", err.Error())
			return err
		}

		ngWord := scheduler.LivecommentScheduler.GetDummyNgWord()
		if err := client.Moderate(ctx, livestream.ID, livestream.Owner.Name, ngWord.Word); err != nil {
			lgr.Warnf("aggressive streamer moderate: failed to moderate: %s\n", err.Error())
			continue
		}
		scheduler.LivecommentScheduler.ModerateNgWord(ngWord.Word)
	}

	return nil
}
