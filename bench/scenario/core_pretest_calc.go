package scenario

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/najeira/randstr"
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

	name := randstr.String(12) + "sc"
	passwd := randstr.String(12)
	user, err := client.Register(ctx, &isupipe.RegisterRequest{
		Name:        name,
		DisplayName: randDisplayName(),
		Description: `普段医療事務員をしています。
よろしくおねがいします！

連絡は以下からお願いします。

ウェブサイト: http://itomanabu.example.com/
メールアドレス: itomanabu@example.com
`,
		Password: passwd,
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: name,
		Password: passwd,
	}); err != nil {
		return err
	}

	stats1, err := client.GetUserStatistics(ctx, user.Name)
	if err != nil {
		return err
	}

	// LivestreamStatsのイテレーション数 * 配信数(2)とかにして、LivestreamStatsのユーザより上に位置するようにする
	count := 5 + rand.Intn(10)
	for i := 0; i < count; i++ {
		viewerClient, err := isupipe.NewCustomResolverClient(
			contestantLogger,
			dnsResolver,
			agent.WithTimeout(config.PretestTimeout),
		)
		if err != nil {
			return err
		}

		name := fmt.Sprintf("%suscv%d", randstr.String(11), i)
		passwd := randstr.String(11)
		viewer, err := viewerClient.Register(ctx, &isupipe.RegisterRequest{
			Name:        name,
			DisplayName: randDisplayName(),
			Description: `普段営業をしています。
よろしくおねがいします！

連絡は以下からお願いします。

ウェブサイト: http://vfujii.example.com/
メールアドレス: vfujii@example.com
`,
			Password: passwd,
			Theme: isupipe.Theme{
				DarkMode: true,
			},
		})
		if err != nil {
			return err
		}

		if err := viewerClient.Login(ctx, &isupipe.LoginRequest{
			Username: viewer.Name,
			Password: passwd,
		}); err != nil {
			return err
		}
	}

	stats2, err := client.GetUserStatistics(ctx, user.Name)
	if err != nil {
		return err
	}

	_ = stats1
	_ = stats2

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

	name := fmt.Sprintf("%slsc", randstr.String(11))
	passwd := randstr.String(17)
	_, err = client.Register(ctx, &isupipe.RegisterRequest{
		Name:        name,
		DisplayName: randDisplayName(),
		Description: `普段薬剤師をしています。
よろしくおねがいします！

連絡は以下からお願いします。

ウェブサイト: http://kobayashiminoru.example.com/
メールアドレス: kobayashiminoru@example.com
`,
		Password: passwd,
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: name,
		Password: passwd,
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}
	if len(livestreams) != 1 {
		return fmt.Errorf("test user has just one livestream")
	}

	// FIXME: ライブコメント投稿のスパム処理にて、正しいNGワードと件数のエラー文が返ってくるように検証

	livestream := livestreams[0]

	// NOTE: rankは変動をみる
	stats1, err := client.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	if stats1.MaxTip != 0 ||
		stats1.TotalReactions != 0 ||
		stats1.TotalReports != 0 ||
		stats1.ViewersCount != 0 {
		return fmt.Errorf("initial livestream stats must be zero")
	}

	count := 5 + rand.Intn(10)
	for i := 0; i < count; i++ {
		viewer, err := isupipe.NewCustomResolverClient(
			contestantLogger,
			dnsResolver,
			agent.WithTimeout(config.PretestTimeout),
		)
		if err != nil {
			return err
		}

		_, err = viewer.Register(ctx, &isupipe.RegisterRequest{
			// FIXME: ユーザ
		})
		if err != nil {
			return err
		}

		if err := viewer.EnterLivestream(ctx, livestream.ID, livestream.Owner.Name); err != nil {
			return err
		}

		_, err = viewer.PostReaction(ctx, livestream.ID, livestream.Owner.Name, &isupipe.PostReactionRequest{
			EmojiName: "innocent",
		})
		if err != nil {
			return err
		}

		livecommentResp, _, err := viewer.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, "isuisu~", &scheduler.Tip{
			Tip: i,
		})
		if err != nil {
			return err
		}

		err = viewer.ReportLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecommentResp.ID)
		if err != nil {
			return err
		}
	}

	stats2, err := client.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}

	_ = stats1
	_ = stats2

	return nil
}
