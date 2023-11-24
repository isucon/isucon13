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

var statsCalcRandSource = rand.New(rand.NewSource(257482710848044431))

// ユーザ統計の計算処理がきちんとできているか
func normalStatsCalcPretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	streamerID := int64(22)
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
		Tags:         []int64{},
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

	// ユーザ(操作前)
	beforeUserStats, err := streamerClient.GetUserStatistics(ctx, streamer.Name)
	if err != nil {
		return err
	}
	beforeWantUserStats, err := scheduler.StatsSched.GetUserStats(streamer.Name)
	if err != nil {
		return err
	}
	beforeWantUserRank, err := scheduler.StatsSched.GetUserRank(streamer.Name)
	if err != nil {
		return err
	}
	if beforeUserStats.Rank != beforeWantUserRank {
		return fmt.Errorf("ユーザ %s のランクが不正です: expected=%d, actual=%d", streamer.Name, beforeWantUserRank, beforeUserStats.Rank)
	}
	beforeWantFavoriteEmoji, ok := beforeWantUserStats.FavoriteEmoji()
	if len(beforeWantFavoriteEmoji) > 0 {
		if ok {
			if beforeUserStats.FavoriteEmoji != beforeWantFavoriteEmoji {
				return fmt.Errorf("ユーザ %s のお気に入り絵文字が不正です: expected=%s, actual=%s", streamer.Name, beforeWantFavoriteEmoji, beforeUserStats.FavoriteEmoji)
			}
		}
	}
	if beforeUserStats.TotalReactions != beforeWantUserStats.TotalReactions() {
		return fmt.Errorf("ユーザ %s の総リアクション数が不正です: expected=%d, actual=%d", streamer.Name, beforeWantUserStats.TotalReactions(), beforeUserStats.TotalReactions)
	}
	if beforeUserStats.ViewersCount != 0 {
		return fmt.Errorf("ユーザ %s の総視聴者数が不正です: expected=%d, actual=%d", streamer.Name, 0, beforeUserStats.TotalLivecomments)
	}
	// 配信(操作前)
	beforeLiveStats, err := streamerClient.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	beforeWantLiveStats, err := scheduler.StatsSched.GetLivestreamStats(livestream.ID)
	if err != nil {
		return err
	}
	beforeWantLiveRank, err := scheduler.StatsSched.GetLivestreamRank(livestream.ID)
	if err != nil {
		return err
	}
	if beforeLiveStats.Rank != beforeWantLiveRank {
		return fmt.Errorf("配信 %d のランクが不正です: expected=%d, actual=%d", livestream.ID, beforeWantLiveRank, beforeLiveStats.Rank)
	}
	if beforeLiveStats.MaxTip != beforeWantLiveStats.MaxTip {
		return fmt.Errorf("配信 %d の最大チップが不正です: expected=%d, actual=%d", livestream.ID, beforeWantLiveStats.MaxTip, beforeLiveStats.MaxTip)
	}
	if beforeLiveStats.TotalReactions != beforeWantLiveStats.TotalReactions {
		return fmt.Errorf("配信 %d の総リアクション数が不正です: expected=%d, actual=%d", livestream.ID, beforeWantLiveStats.TotalReactions, beforeLiveStats.TotalReactions)
	}
	if beforeLiveStats.TotalReports != 0 {
		return fmt.Errorf("配信 %d の総スパム報告数が不正です: expected=%d, actual=%d", livestream.ID, 0, beforeLiveStats.TotalReports)
	}
	if beforeLiveStats.ViewersCount != 0 {
		return fmt.Errorf("配信 %d の総視聴者数が不正です: expected=%d, actual=%d", livestream.ID, 0, beforeLiveStats.ViewersCount)
	}

	// 操作
	viewerClient, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}
	viewerID := int64(22)
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

	if err := viewerClient.EnterLivestream(ctx, livestream.ID, livestream.Owner.Name); err != nil {
		return err
	}
	if err := scheduler.StatsSched.EnterLivestream(livestream.Owner.Name, livestream.ID); err != nil {
		return fmt.Errorf("EnterLivestreamに失敗。内部的なエラーであるため、運営に連絡してください。")
	}
	reactionCount := 1 + statsCalcRandSource.Intn(3)
	for r := 0; r < reactionCount; r++ {
		reaction := scheduler.GetReaction()
		if _, err := viewerClient.PostReaction(ctx, livestream.ID, livestream.Owner.Name, &isupipe.PostReactionRequest{
			EmojiName: reaction,
		}); err != nil {
			return err
		}
		if err := scheduler.StatsSched.AddReaction(livestream.Owner.Name, livestream.ID, reaction); err != nil {
			return fmt.Errorf("AddReactionに失敗。内部的なエラーであるため、運営に連絡してください。")
		}
	}
	livecommentCount := 1 + statsCalcRandSource.Intn(3)
	var livecomments []*isupipe.PostLivecommentResponse
	for l := 0; l < livecommentCount; l++ {
		livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
		tip := &scheduler.Tip{Tip: rand.Intn(10)}
		resp, _, err := viewerClient.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecomment.Comment, tip)
		if err != nil {
			return err
		}
		if err := scheduler.StatsSched.AddLivecomment(livestream.Owner.Name, livestream.ID, tip); err != nil {
			return fmt.Errorf("AddLivecommentに失敗。内部的なエラーであるため、運営に連絡してください。")
		}
		livecomments = append(livecomments, resp)
	}
	for _, livecomment := range livecomments {
		if err := viewerClient.ReportLivecomment(ctx, livecomment.Livestream.ID, livecomment.User.Name, livecomment.ID); err != nil {
			return err
		}
		if err := scheduler.StatsSched.AddReport(livecomment.User.Name, livecomment.Livestream.ID); err != nil {
			return fmt.Errorf("AddReportに失敗。内部的なエラーであるため、運営に連絡してください。")
		}
	}

	// ユーザ(操作後)
	afterUserStats, err := streamerClient.GetUserStatistics(ctx, streamer.Name)
	if err != nil {
		return err
	}
	afterWantUserStats, err := scheduler.StatsSched.GetUserStats(streamer.Name)
	if err != nil {
		return err
	}
	afterWantUserRank, err := scheduler.StatsSched.GetUserRank(streamer.Name)
	if err != nil {
		return err
	}
	if afterUserStats.Rank != afterWantUserRank {
		return fmt.Errorf("ユーザ %s のランクが不正です: expected=%d, actual=%d", streamer.Name, afterWantUserRank, afterUserStats.Rank)
	}
	afterWantFavoriteEmoji, ok := afterWantUserStats.FavoriteEmoji()
	if len(afterWantFavoriteEmoji) > 0 {
		if ok {
			if afterUserStats.FavoriteEmoji != afterWantFavoriteEmoji {
				return fmt.Errorf("ユーザ %s のお気に入り絵文字が不正です: expected=%s, actual=%s", streamer.Name, afterWantFavoriteEmoji, afterUserStats.FavoriteEmoji)
			}
		}
	}
	if afterUserStats.TotalReactions != afterWantUserStats.TotalReactions() {
		return fmt.Errorf("ユーザ %s の総リアクション数が不正です: expected=%d, actual=%d", streamer.Name, afterWantUserStats.TotalReactions(), afterUserStats.TotalReactions)
	}
	if afterUserStats.ViewersCount != afterWantUserStats.TotalViewers {
		return fmt.Errorf("ユーザ %s の総視聴者数が不正です: expected=%d, actual=%d", streamer.Name, afterWantUserStats.TotalLivecomments, afterUserStats.TotalLivecomments)
	}

	// 配信(操作後)
	afterLiveStats, err := streamerClient.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	afterWantLiveStats, err := scheduler.StatsSched.GetLivestreamStats(livestream.ID)
	if err != nil {
		return err
	}
	afterWantLiveRank, err := scheduler.StatsSched.GetLivestreamRank(livestream.ID)
	if err != nil {
		return err
	}
	if afterLiveStats.Rank != afterWantLiveRank {
		return fmt.Errorf("配信 %d のランクが不正です: expected=%d, actual=%d", livestream.ID, afterWantLiveRank, afterLiveStats.Rank)
	}
	if afterLiveStats.MaxTip != afterWantLiveStats.MaxTip {
		return fmt.Errorf("配信 %d の最大チップが不正です: expected=%d, actual=%d", livestream.ID, afterWantLiveStats.MaxTip, afterLiveStats.MaxTip)
	}
	if afterLiveStats.TotalReactions != afterWantLiveStats.TotalReactions {
		return fmt.Errorf("配信 %d の総リアクション数が不正です: expected=%d, actual=%d", livestream.ID, afterWantLiveStats.TotalReactions, afterLiveStats.TotalReactions)
	}
	if afterLiveStats.TotalReports != afterWantLiveStats.TotalReports {
		return fmt.Errorf("ユーザ %s の総スパム報告数が不正です: expected=%d, actual=%d", streamer.Name, afterWantLiveStats.TotalReports, afterLiveStats.TotalReports)
	}
	if afterLiveStats.ViewersCount != afterWantLiveStats.TotalViewers {
		return fmt.Errorf("ユーザ %s の総視聴者数が不正です: expected=%d, actual=%d", streamer.Name, afterWantLiveStats.TotalReports, afterLiveStats.ViewersCount)
	}

	return nil
}
