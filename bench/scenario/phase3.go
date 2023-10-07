package scenario

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"golang.org/x/sync/errgroup"
)

var (
	ErrColdReservation = errors.New("coldな予約がこれ以上ありません")
)

// FIXME: 並列処理で同一時間帯にたくさん予約処理をかける
func runConcurrentReservation(ctx context.Context) error {
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

	// 配信確定
	var reserveGrp errgroup.Group
	var livestream *isupipe.Livestream
	// for i := 0; i < 10; i++ {
	reserveGrp.Go(func() error {
		// 予約を実施
		reservation, err := scheduler.Phase3ReservationScheduler.GetHotShortReservation()
		if err != nil {
			return err
		}

		ls, err := vtuberClient.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
			Tags:        []int{},
			Title:       reservation.Title,
			Description: reservation.Description,
			StartAt:     reservation.StartAt,
			EndAt:       reservation.EndAt,
		})
		if err != nil {
			// エラーが出る限り続ける
			scheduler.Phase3ReservationScheduler.AbortReservation(reservation)
			return nil
		}
		scheduler.Phase3ReservationScheduler.CommitReservation(reservation)

		livestream = ls

		return fmt.Errorf("stop reserve")
	})
	// }

	reserveGrp.Wait()
	if livestream == nil {
		return fmt.Errorf("failed to reserve livestream")
	}

	// 作成された予約について、視聴者を生成して投げ銭を稼がせてあげる

	// 視聴者を決定
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
}

func runDnsAttackScenario() {
	// FIXME: powerdns用意できてから
}

// スパムを投稿しまくる
func runSpamScenario(ctx context.Context) error {
	// 配信者を選定
	vtuber := scheduler.UserScheduler.SelectVTuber()

	livestream, err := scheduler.Phase3ReservationScheduler.GetStreamFor(vtuber)
	if err != nil {
		return err
	}

	// スパム
	client, err := isupipe.NewClient(agent.WithBaseURL(config.TargetBaseURL))
	if err != nil {
		return err
	}

	viewer := scheduler.UserScheduler.SelectViewer()
	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: viewer.Name,
		Password: viewer.RawPassword,
	}); err != nil {
		return err
	}

	comment := scheduler.LivecommentScheduler.GetNegativeComment()
	if _, err := client.PostLivecomment(ctx, livestream.Id, &isupipe.PostLivecommentRequest{
		Comment: comment.Comment,
		Tip:     0,
	}); err != nil {
		return err
	}

	return nil
}

// 重複した予約を投げる
// func runOverlapReserveScenario() {
// 	// FIXME: 予約済み一覧から適当に払い出し(rand)、それを使って同じ時間帯で予約すればいい
// }

func Phase3(ctx context.Context) error {
	var eg errgroup.Group

	for i := 0; i < 20; i++ {
		runRegisterScenario(ctx)
	}

	// 通常配信者
	// countは広告費用係数に合わせて増やす
	for i := 0; i < config.AdvertiseCost; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					runConcurrentReservation(ctx)
					runDnsAttackScenario()
					runSpamScenario(ctx)
				}
			}
		})
	}

	// 配信者のプロフィールを見に行く

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
