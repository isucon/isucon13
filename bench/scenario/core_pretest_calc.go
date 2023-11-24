package scenario

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

// 計算処理のpretest

func normalPaymentCalcPretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	// チップ投稿により正しく計算されるか
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	payment1, err := client.GetPaymentResult(ctx)
	if err != nil {
		return err
	}

	_ = payment1

	// FIXME: 処理前、paymentが0円になってることをチェック
	// FIXME: 処理後、paymentが指定金額になっていることをチェック

	payment2, err := client.GetPaymentResult(ctx)
	if err != nil {
		return err
	}

	_ = payment2

	return nil
}

// ユーザ統計の計算処理がきちんとできているか
func normalUserStatsCalcPretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	streamerID := int64(1 + (rand.Int() % 10))
	streamer, err := scheduler.UserScheduler.GetInitialUserForPretest(streamerID)
	if err != nil {
		return err
	}
	streamerClient, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}
	if err := streamerClient.Login(ctx, &isupipe.LoginRequest{
		Username: streamer.Name,
		Password: streamer.RawPassword,
	}); err != nil {
		return err
	}

	coldReservation, err := scheduler.ReservationSched.GetColdShortReservation()
	if err != nil {
		log.Println(err)
		return err
	}
	livestream, err := streamerClient.ReserveLivestream(ctx, streamer.Name, &isupipe.ReserveLivestreamRequest{
		Title:        coldReservation.Title,
		Description:  coldReservation.Description,
		PlaylistUrl:  "https://media.xiii.isucon.dev/api/4/playlist.m3u8",
		ThumbnailUrl: "https://media.xiii.isucon.dev/isucon12_final.webp",
		StartAt:      coldReservation.StartAt,
		EndAt:        coldReservation.EndAt,
	})
	if err != nil {
		scheduler.ReservationSched.AbortReservation(coldReservation)
		return err
	}
	scheduler.StatsSched.AddLivestream(livestream.ID)
	scheduler.ReservationSched.CommitReservation(coldReservation)

	stats1, err := client.GetUserStatistics(ctx, streamer.Name)
	if err != nil {
		return err
	}
	userStats1, err := scheduler.StatsSched.GetUserStats(streamer.Name)
	if err != nil {
		return err
	}
	userRank, err := scheduler.StatsSched.GetUserRank(streamer.Name)
	if err != nil {
		return err
	}
	if stats1.Rank != userRank {
		return fmt.Errorf("ユーザ %s のランクが不正です: expected=%d, actual=%d", streamer.Name, userRank, stats1.Rank)
	}
	favoriteEmoji, ok := userStats1.FavoriteEmoji()
	if ok {
		if stats1.FavoriteEmoji != favoriteEmoji {
			return fmt.Errorf("ユーザ %s のお気に入り絵文字が不正です: expected=%s, actual=%s", streamer.Name, favoriteEmoji, stats1.FavoriteEmoji)
		}
	}
	if stats1.TotalReactions != userStats1.TotalReactions() {
		return fmt.Errorf("ユーザ %s の総リアクション数が不正です: expected=%d, actual=%d", streamer.Name, stats1.TotalReactions, userStats1.TotalReactions())
	}
	if stats1.TotalLivecomments != userStats1.TotalLivecomments {
		return fmt.Errorf("ユーザ %s の総ライブコメント数が不正です: expected=%d, actual=%d", streamer.Name, stats1.TotalLivecomments, userStats1.TotalLivecomments)
	}
	if stats1.ViewersCount != 0 {
		return fmt.Errorf("ユーザ %s の総視聴者数が不正です: expected=%d, actual=%d", streamer.Name, stats1.TotalLivecomments, userStats1.TotalLivecomments)
	}
	liveStats1, err := client.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	wantStats1, err := scheduler.StatsSched.GetLivestreamStats(livestream.ID)
	if err != nil {
		return err
	}
	livestreamRank, err := scheduler.StatsSched.GetLivestreamRank(livestream.ID)
	if err != nil {
		return err
	}
	if liveStats1.Rank != livestreamRank {
		return fmt.Errorf("配信 %d のランクが不正です: expected=%d, actual=%d", livestream.ID, livestreamRank, liveStats1.Rank)
	}
	if liveStats1.MaxTip != wantStats1.MaxTip {
		return fmt.Errorf("配信 %d の最大チップが不正です: expected=%d, actual=%d", livestream.ID, wantStats1.MaxTip, liveStats1.MaxTip)
	}
	if liveStats1.TotalReactions != wantStats1.TotalReactions {
		return fmt.Errorf("配信 %d の総リアクション数が不正です: expected=%d, actual=%d", livestream.ID, wantStats1.TotalReactions, liveStats1.TotalReactions)
	}
	if liveStats1.TotalReports != 0 {
		return fmt.Errorf("ユーザ %s の総スパム報告数が不正です: expected=%d, actual=%d", streamer.Name, wantStats1.TotalReports, liveStats1.TotalReports)
	}
	if liveStats1.ViewersCount != 0 {
		return fmt.Errorf("ユーザ %s の総視聴者数が不正です: expected=%d, actual=%d", streamer.Name, wantStats1.TotalViewers, wantStats1.TotalViewers)
	}

	count := 1 + rand.Intn(5)
	for i := 0; i < count; i++ {
		viewerClient, err := isupipe.NewCustomResolverClient(
			contestantLogger,
			dnsResolver,
			agent.WithTimeout(config.PretestTimeout),
		)
		if err != nil {
			return err
		}

		viewerID := int64(1 + (rand.Int() % 10))
		viewer, err := scheduler.UserScheduler.GetInitialUserForPretest(viewerID)
		if err != nil {
			return err
		}
		if err := viewerClient.Login(ctx, &isupipe.LoginRequest{
			Username: viewer.Name,
			Password: viewer.RawPassword,
		}); err != nil {
			return err
		}

		viewerClient.EnterLivestream(ctx, livestream.ID, livestream.Owner.Name)
		scheduler.StatsSched.EnterLivestream(livestream.Owner.Name, livestream.ID)

		reactionCount := 1 + rand.Intn(10)
		for r := 0; r < reactionCount; r++ {
			reaction := scheduler.GetReaction()
			viewerClient.PostReaction(ctx, livestream.ID, livestream.Owner.Name, &isupipe.PostReactionRequest{
				EmojiName: reaction,
			})
			scheduler.StatsSched.AddReaction(livestream.Owner.Name, livestream.ID, reaction)
		}

		livecommentCount := 1 + rand.Intn(10)
		for l := 0; l < livecommentCount; l++ {
			livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
			tip := &scheduler.Tip{Tip: rand.Intn(10)}
			viewerClient.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecomment.Comment, tip)
			scheduler.StatsSched.AddLivecomment(livestream.Owner.Name, livestream.ID, tip)
		}
	}

	stats2, err := client.GetUserStatistics(ctx, streamer.Name)
	if err != nil {
		return err
	}
	userStats2, err := scheduler.StatsSched.GetUserStats(streamer.Name)
	if err != nil {
		return err
	}
	userRank2, err := scheduler.StatsSched.GetUserRank(streamer.Name)
	if err != nil {
		return err
	}
	if stats2.Rank != userRank2 {
		return fmt.Errorf("ユーザ %s のランクが不正です: expected=%d, actual=%d", streamer.Name, userRank, stats1.Rank)
	}
	favoriteEmoji2, ok := userStats1.FavoriteEmoji()
	if ok {
		if stats2.FavoriteEmoji != favoriteEmoji2 {
			return fmt.Errorf("ユーザ %s のお気に入り絵文字が不正です: expected=%s, actual=%s", streamer.Name, favoriteEmoji2, stats2.FavoriteEmoji)
		}
	}
	if stats2.TotalReactions != userStats2.TotalReactions() {
		return fmt.Errorf("ユーザ %s の総リアクション数が不正です: expected=%d, actual=%d", streamer.Name, stats2.TotalReactions, userStats2.TotalReactions())
	}
	if stats2.TotalLivecomments != userStats2.TotalLivecomments {
		return fmt.Errorf("ユーザ %s の総ライブコメント数が不正です: expected=%d, actual=%d", streamer.Name, stats2.TotalLivecomments, userStats2.TotalLivecomments)
	}
	if stats2.ViewersCount != userStats2.TotalViewers {
		return fmt.Errorf("ユーザ %s の総視聴者数が不正です: expected=%d, actual=%d", streamer.Name, stats2.TotalLivecomments, userStats2.TotalLivecomments)
	}
	liveStats2, err := client.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	wantStats2, err := scheduler.StatsSched.GetLivestreamStats(livestream.ID)
	if err != nil {
		return err
	}
	livestreamRank2, err := scheduler.StatsSched.GetLivestreamRank(livestream.ID)
	if err != nil {
		return err
	}
	if liveStats2.Rank != livestreamRank2 {
		return fmt.Errorf("配信 %d のランクが不正です: expected=%d, actual=%d", livestream.ID, livestreamRank, liveStats2.Rank)
	}
	if liveStats2.MaxTip != wantStats2.MaxTip {
		return fmt.Errorf("配信 %d の最大チップが不正です: expected=%d, actual=%d", livestream.ID, wantStats2.MaxTip, liveStats2.MaxTip)
	}
	if liveStats2.TotalReactions != wantStats2.TotalReactions {
		return fmt.Errorf("配信 %d の総リアクション数が不正です: expected=%d, actual=%d", livestream.ID, wantStats2.TotalReactions, liveStats2.TotalReactions)
	}
	if liveStats2.TotalReports != wantStats2.TotalReports {
		return fmt.Errorf("ユーザ %s の総スパム報告数が不正です: expected=%d, actual=%d", streamer.Name, wantStats2.TotalReports, liveStats2.TotalReports)
	}
	if liveStats2.ViewersCount != wantStats2.TotalViewers {
		return fmt.Errorf("ユーザ %s の総視聴者数が不正です: expected=%d, actual=%d", streamer.Name, wantStats2.TotalReports, liveStats2.ViewersCount)
	}

	return nil
}
