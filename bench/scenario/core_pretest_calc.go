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

	randNumber := rand.Intn(100)
	loginUser := scheduler.GetInitialUserByID(int64(1 + randNumber))
	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: loginUser.Name,
		Password: loginUser.RawPassword,
	}); err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: loginUser.Name,
		Password: loginUser.RawPassword,
	}); err != nil {
		return err
	}

	userID := int64(1 + (rand.Int() % 10))
	user, err := scheduler.UserScheduler.GetInitialUserForPretest(userID)
	if err != nil {
		return err
	}
	username := user.Name

	stats1, err := client.GetUserStatistics(ctx, username)
	if err != nil {
		return err
	}

	wantStats1, err := scheduler.StatsSched.GetUserStats(username)
	if err != nil {
		return err
	}

	log.Printf("stats1 = %+v\n", stats1)
	log.Printf("wantStats1 = %+v\n", wantStats1)

	userRank, err := scheduler.StatsSched.GetUserRank(username)
	if err != nil {
		return err
	}
	if stats1.Rank != userRank {
		return fmt.Errorf("ユーザ %s のランクが不正です: expected=%d, actual=%d", username, userRank, stats1.Rank)
	}
	favoriteEmoji, ok := wantStats1.FavoriteEmoji()
	if ok {
		if stats1.FavoriteEmoji != favoriteEmoji {
			return fmt.Errorf("ユーザ %s のお気に入り絵文字が不正です: expected=%s, actual=%s", username, favoriteEmoji, stats1.FavoriteEmoji)
		}
	}
	if stats1.TotalReactions != wantStats1.TotalReactions() {
		return fmt.Errorf("ユーザ %s の総リアクション数が不正です: expected=%d, actual=%d", username, stats1.TotalReactions, wantStats1.TotalReactions())
	}
	if stats1.TotalLivecomments != wantStats1.TotalLivecomments {
		return fmt.Errorf("ユーザ %s の総ライブコメント数が不正です: expected=%d, actual=%d", username, stats1.TotalLivecomments, wantStats1.TotalLivecomments)
	}

	// // LivestreamStatsのイテレーション数 * 配信数(2)とかにして、LivestreamStatsのユーザより上に位置するようにする
	// count := 5 + rand.Intn(10)
	// for i := 0; i < count; i++ {
	// 	viewerClient, err := isupipe.NewCustomResolverClient(
	// 		contestantLogger,
	// 		dnsResolver,
	// 		agent.WithTimeout(config.PretestTimeout),
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	name := fmt.Sprintf("user-stats-calc-viewer%d", i)
	// 	viewer, err := viewerClient.Register(ctx, &isupipe.RegisterRequest{
	// 		Name:        name,
	// 		DisplayName: name,
	// 		Description: name,
	// 		Password:    "test",
	// 		Theme: isupipe.Theme{
	// 			DarkMode: true,
	// 		},
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if err := viewerClient.Login(ctx, &isupipe.LoginRequest{
	// 		Username: viewer.Name,
	// 		Password: "test",
	// 	}); err != nil {
	// 		return err
	// 	}
	// }

	// stats2, err := client.GetUserStatistics(ctx, user.Name)
	// if err != nil {
	// 	return err
	// }

	// _ = stats1
	// _ = stats2

	return nil
}

func normalLivestreamStatsCalcPretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	// ライブストリーム統計の計算処理がきちんとできているか

	// FIXME: 処理前、統計情報がすべて0になっていることをチェック
	// FIXME: いくつかの処理後、統計情報がピタリ一致することをチェック
	//        (処理数、処理データにランダム性をもたせる)
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	randNumber := rand.Intn(100)
	user := scheduler.GetInitialUserByID(int64(1 + randNumber))
	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: user.Name,
		Password: user.RawPassword,
	}); err != nil {
		return err
	}

	livestreamID := int64(scheduler.GetLivestreamLength() - randNumber)
	livestream := scheduler.GetLivestreamByID(livestreamID)
	streamer, err := scheduler.UserScheduler.GetInitialUserForPretest(livestream.OwnerID)
	if err != nil {
		return err
	}

	stats1, err := client.GetLivestreamStatistics(ctx, livestreamID, streamer.Name)
	if err != nil {
		return err
	}

	wantStats1, err := scheduler.StatsSched.GetLivestreamStats(livestreamID)
	if err != nil {
		return err
	}
	livestreamRank, err := scheduler.StatsSched.GetLivestreamRank(livestreamID)
	if err != nil {
		return err
	}

	if stats1.Rank != livestreamRank {
		return fmt.Errorf("配信 %d のランクが不正です: expected=%d, actual=%d", livestreamID, livestreamRank, stats1.Rank)
	}
	if stats1.MaxTip != wantStats1.MaxTip {
		return fmt.Errorf("配信 %d の最大チップが不正です: expected=%d, actual=%d", livestreamID, livestreamRank, stats1.Rank)
	}
	if stats1.TotalReactions != wantStats1.TotalReactions {
		return fmt.Errorf("配信 %d の総リアクション数が不正です: expected=%d, actual=%d", livestreamID, livestreamRank, stats1.Rank)
	}

	// count := 5 + rand.Intn(10)
	// for i := 0; i < count; i++ {
	// 	viewer, err := isupipe.NewCustomResolverClient(
	// 		contestantLogger,
	// 		dnsResolver,
	// 		agent.WithTimeout(config.PretestTimeout),
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	_, err = viewer.Register(ctx, &isupipe.RegisterRequest{
	// 		// FIXME: ユーザ
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if err := viewer.EnterLivestream(ctx, livestream.ID, livestream.Owner.Name); err != nil {
	// 		return err
	// 	}

	// 	_, err = viewer.PostReaction(ctx, livestream.ID, livestream.Owner.Name, &isupipe.PostReactionRequest{
	// 		EmojiName: "innocent",
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	livecommentResp, _, err := viewer.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, "isuisu~", &scheduler.Tip{
	// 		Tip: i,
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = viewer.ReportLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecommentResp.ID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// stats2, err := client.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name)
	// if err != nil {
	// 	return err
	// }

	// _ = stats1
	// _ = stats2

	return nil
}
