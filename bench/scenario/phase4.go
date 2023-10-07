package scenario

import (
	"context"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"golang.org/x/sync/errgroup"
)

// FIXME: coldで埋めていって、取れなくなったらhotで一斉に集中狙いしていく

// FIXME: 予約の衝突を起こしそうなトラフィックを流す

// ライブコメントを投稿
func runReserveScenario2(ctx context.Context) error {
	// 配信者決定
	vtuber := scheduler.UserScheduler.SelectVTuber()

	// ログイン
	vtuberClient, err := isupipe.NewClient(agent.WithBaseURL(config.TargetBaseURL))
	if err != nil {
		return err
	}

	if err := vtuberClient.Login(ctx, &isupipe.LoginRequest{
		UserName: vtuber.Name,
		Password: vtuber.RawPassword,
	}); err != nil {
		return err
	}

	// 予約を実施
	reservation, err := scheduler.Phase2ReservationScheduler.GetHotShortReservation()
	if err != nil {
		return err
	}

	// 配信確定
	livestream, err := vtuberClient.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		Tags:        []int{},
		Title:       reservation.Title,
		Description: reservation.Description,
		StartAt:     reservation.StartAt,
		EndAt:       reservation.EndAt,
	})
	if err != nil {
		return err
	}

	// 作成された予約について、視聴者を生成して投げ銭を稼がせてあげる

	// 視聴者を決定
	var viewersGrp errgroup.Group
	for i := 0; i < 10; i++ {
		viewersGrp.Go(func() error {
			viewer := scheduler.UserScheduler.SelectViewer()

			viewerClient, err := isupipe.NewClient(agent.WithBaseURL(config.TargetBaseURL))
			if err != nil {
				return err
			}

			if err := viewerClient.Login(ctx, &isupipe.LoginRequest{
				UserName: viewer.Name,
				Password: viewer.RawPassword,
			}); err != nil {
				return err
			}

			if _, err := viewerClient.GetLivestreamStatistics(ctx, livestream.Id); err != nil {
				return err
			}
			if _, err := viewerClient.GetUserStatistics(ctx, vtuber.UserId); err != nil {
				return err
			}

			// enter
			if err := viewerClient.EnterLivestream(ctx, livestream.Id); err != nil {
				return err
			}

			// FIXME: 時間枠の長さに合わせて投げ銭の機会を増やすため、forループの長さを変える
			for i := 0; i < 10; i++ {
				// ライブコメント決定
				comment := scheduler.LivecommentScheduler.GetShortPositiveComment()
				tip := scheduler.LivecommentScheduler.GetTipsForStream()

				// 視聴者からライブコメント投稿 (投げ銭)
				if _, err := viewerClient.PostLivecomment(ctx, livestream.Id, &isupipe.PostLivecommentRequest{
					Comment: comment.Comment,
					Tip:     tip,
				}); err != nil {
					return err
				}
			}

			// leave
			if err := viewerClient.LeaveLivestream(ctx, livestream.Id); err != nil {
				return err
			}

			return nil
		})
	}

	return viewersGrp.Wait()
}

func Phase4(ctx context.Context) error {
	var eg errgroup.Group

	// 通常配信者
	// countは広告費用係数に合わせて増やす
	for i := 0; i < config.AdvertiseCost; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					// 高い並列性で書き込みまくる
					// 合間にプロフィール閲覧、統計情報取得などを入れて、売上などのハードルにする
					go runReserveScenario(ctx)
				}
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
