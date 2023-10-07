package scenario

import (
	"context"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"golang.org/x/sync/errgroup"
)

// ライブコメントを投稿
func runPostLivecommentScenario(ctx context.Context) error {
	// 配信者決定
	vtuber := scheduler.UserScheduler.SelectVTuberForSeason1()

	// 配信を決定
	livestream, err := scheduler.Phase1ReservationScheduler.GetStreamFor(vtuber)
	if err != nil {
		return err
	}

	// 視聴者を決定
	viewer := scheduler.UserScheduler.SelectViewerForSeason1()

	// ログイン
	client, err := isupipe.NewClient(agent.WithBaseURL(config.TargetBaseURL))
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: viewer.Name,
		Password: viewer.RawPassword,
	}); err != nil {
		return err
	}

	// enter
	if err := client.EnterLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	// FIXME: 時間枠の長さに合わせて投げ銭の機会を増やすため、forループの長さを変える
	for i := 0; i < 10; i++ {
		// ライブコメント決定
		comment := scheduler.LivecommentScheduler.GetShortPositiveComment()
		tip := scheduler.LivecommentScheduler.GetTipsForStream()

		// 視聴者からライブコメント投稿 (投げ銭)
		if _, err := client.PostLivecomment(ctx, livestream.Id, &isupipe.PostLivecommentRequest{
			Comment: comment.Comment,
			Tip:     tip,
		}); err != nil {
			return err
		}
	}

	// leave
	if err := client.LeaveLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	return nil
}

func Phase1(ctx context.Context) error {
	var eg errgroup.Group

	// トップ画面で検索を行う

	// 通常視聴者
	// countは広告費用係数に合わせて増やす
	for i := 0; i < config.AdvertiseCost; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-benchscore.Achieve():
					return nil
				default:
					go runPostLivecommentScenario(ctx)
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
