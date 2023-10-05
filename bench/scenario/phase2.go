package scenario

import (
	"context"
	"log"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"golang.org/x/sync/errgroup"
)

// FIXME: Hot/Cold問わないで予約を取るので、GetReservationみたいなの生やす

// ライブコメントを投稿
func runReserveScenario(ctx context.Context) error {
	log.Println("run scenario")
	// 配信者決定
	vtuber := scheduler.UserScheduler.SelectVTuber()

	// ログイン
	log.Println("login")
	vtuberClient, err := isupipe.NewClient(agent.WithBaseURL(config.TargetBaseURL))
	if err != nil {
		log.Println(err)
		return err
	}

	if err := vtuberClient.Login(ctx, &isupipe.LoginRequest{
		UserName: vtuber.Name,
		Password: vtuber.RawPassword,
	}); err != nil {
		log.Println(err)
		return err
	}

	// 予約を実施
	log.Println("reserve")
	reservation, err := scheduler.Phase2ReservationScheduler.GetHotShortReservation()
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		return err
	}

	// 作成された予約について、視聴者を生成して投げ銭を稼がせてあげる

	// 視聴者を決定
	log.Println("setup viewer")
	viewer := scheduler.UserScheduler.SelectViewer()

	viewerClient, err := isupipe.NewClient(agent.WithBaseURL(config.TargetBaseURL))
	if err != nil {
		log.Println(err)
		return err
	}

	if err := viewerClient.Login(ctx, &isupipe.LoginRequest{
		UserName: viewer.Name,
		Password: viewer.RawPassword,
	}); err != nil {
		log.Println(err)
		return err
	}

	// enter
	log.Println("enter")
	if err := viewerClient.EnterLivestream(ctx, livestream.Id); err != nil {
		log.Println(err)
		return err
	}

	// FIXME: 時間枠の長さに合わせて投げ銭の機会を増やすため、forループの長さを変える
	for i := 0; i < 10; i++ {
		// ライブコメント決定
		comment := scheduler.LivecommentScheduler.GetShortPositiveComment()
		tip := scheduler.LivecommentScheduler.GetTipsForStream()

		// 視聴者からライブコメント投稿 (投げ銭)
		log.Println("post livecomment tips")
		if _, err := viewerClient.PostLivecomment(ctx, livestream.Id, &isupipe.PostLivecommentRequest{
			Comment: comment.Comment,
			Tip:     tip,
		}); err != nil {
			return err
		}
	}

	// leave
	log.Println("leave")
	if err := viewerClient.LeaveLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	return nil
}

func Phase2(ctx context.Context) error {
	var eg errgroup.Group

	// 通常配信者
	// countは広告費用係数に合わせて増やす
	count := 30
	for i := 0; i < count; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					runReserveScenario(ctx)
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
