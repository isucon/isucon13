package scenario

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"math/rand"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
)

//go:embed testdata/NoImage.jpg
var fallbackImage []byte

// 基本機能のロジックpretest

func NormalUserPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	user, err := client.Register(ctx, &isupipe.RegisterRequest{
		Name:        "test",
		DisplayName: "test",
		Description: "blah blah blah",
		Password:    testUserRawPassword,
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: user.Name,
		Password: testUserRawPassword,
	}); err != nil {
		return err
	}

	if _, err := client.GetMe(ctx); err != nil {
		return err
	}

	if err := client.GetUser(ctx, user.Name); err != nil {
		return err
	}

	if _, err := client.GetStreamerTheme(ctx, user); err != nil {
		return err
	}

	return nil
}

func NormalLivestreamPretest(ctx context.Context, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	// 機能的なテスト
	// 予約したライブ配信が一覧に見えるか、取得できるか、検索によって見つけられるか
	// enter/exitできるか (other)

	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: testUser.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	tagResponse, err := client.GetTags(ctx)
	if err != nil {
		return err
	}

	var (
		tagCount    = rand.Intn(5)
		tagStartIdx = rand.Intn(len(tagResponse.Tags))
		tagEndIdx   = min(tagStartIdx+tagCount, len(tagResponse.Tags))
	)
	tags := []int64{}
	for _, tag := range tagResponse.Tags[tagStartIdx:tagEndIdx] {
		tags = append(tags, int64(tag.ID))
	}

	var (
		startAt = time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local)
		endAt   = time.Date(2024, 4, 1, 1, 0, 0, 0, time.Local)
	)
	livestream, err := client.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		Tags:        tags,
		Title:       "pretest",
		Description: "pretest",
		// FIXME: フロントで困らないようにちゃんとしたのを設定
		PlaylistUrl:  "",
		ThumbnailUrl: "",
		StartAt:      startAt.Unix(),
		EndAt:        endAt.Unix(),
	})
	if err != nil {
		return err
	}

	if err = client.GetLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	if err := client.EnterLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	if err := client.ExitLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	// FIXME: 結果をちゃんと確認
	if err := client.SearchLivestreamsByTag(ctx, "椅子" /* tag name */); err != nil {
		return err
	}

	return nil
}

func NormalIconPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: "test001",
		Password: "test",
	}); err != nil {
		return err
	}

	// アイコンを投稿する前、No Imageの画像が返されているか
	icon1, err := client.GetIcon(ctx, "test001")
	if err != nil {
		return err
	}
	if !bytes.Equal(icon1[:], fallbackImage[:]) {
		return fmt.Errorf("アイコン未設定の場合は、NoImage.jpgを返さなければなりません")
	}

	// アイコンを投稿後、期待するアイコンが設定されているか
	randomIcon := scheduler.IconSched.GetRandomIcon()
	if _, err := client.PostIcon(ctx, &isupipe.PostIconRequest{
		Image: randomIcon.Image,
	}); err != nil {
		return err
	}

	icon2, err := client.GetIcon(ctx, "test001")
	if err != nil {
		return err
	}
	icon2Hash := sha256.Sum256(icon2)
	if !bytes.Equal(icon2Hash[:], randomIcon.Hash[:]) {
		return fmt.Errorf("新たに設定したアイコンが反映されていません")
	}

	return nil
}

func NormalPostLivecommentPretest(ctx context.Context, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: testUser.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	livestream := livestreams[0]

	//

	if _, err = client.GetLivecommentReports(ctx, livestream.ID); err != nil {
		return err
	}

	notip := &scheduler.Tip{}
	_, _, err = client.PostLivecomment(ctx, livestream.ID, "test", notip)
	if err != nil {
		return err
	}

	livecomments, err := client.GetLivecomments(ctx, livestream.ID, isupipe.WithLimitQueryParam(&isupipe.LimitParam{
		Limit: 1,
	}))
	if err != nil {
		return err
	}

	livecomment := livecomments[0]

	if err := client.ReportLivecomment(ctx, livestream.ID, livecomment.ID); err != nil {
		return err
	}

	return nil
}

func NormalReactionPretest(ctx context.Context, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	// 投稿したリアクションがGETできるか
	// limitをつけられるか
	// 初期データが期待する件数あるか
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: testUser.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	livestream := livestreams[0]

	if _, err := client.PostReaction(ctx, livestream.ID /* livestream id*/, &isupipe.PostReactionRequest{
		EmojiName: "chair",
	}); err != nil {
		return err
	}

	reactions, err := client.GetReactions(ctx, livestream.ID /* livestream id*/)
	if err != nil {
		return err
	}
	if len(reactions) != 1 {
		return fmt.Errorf("リアクション件数が不正です")
	}

	return nil
}

func NormalReportLivecommentPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	// ライブコメントを1件取得(limit=1)
	// ライブコメントを報告できるか (other)
	// 報告したものが確認できるか (owner)

	// 初期で報告が0件

	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: "test001",
		Password: "test",
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	livestream := livestreams[0]

	reports, err := client.GetLivecommentReports(ctx, livestream.ID)
	if err != nil {
		return err
	}
	if len(reports) != 0 {
		return fmt.Errorf("初期のtest001ユーザのライブ配信におけるスパム報告は0件でなければなりません")
	}

	livecomments, err := client.GetLivecomments(ctx, livestream.ID, isupipe.WithLimitQueryParam(&isupipe.LimitParam{
		Limit: 10,
	}))
	if err != nil {
		return err
	}
	if len(livecomments) != 10 {
		return fmt.Errorf("limitを使用してライブコメント取得しましたが、指定件数が返ってきませんでした")
	}

	rand.Shuffle(len(livecomments), func(i, j int) {
		livecomments[i], livecomments[j] = livecomments[j], livecomments[i]
	})
	livecomment := livecomments[0]

	reporterClient, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}
	reporter, err := reporterClient.Register(ctx, &isupipe.RegisterRequest{
		Name:        "report",
		DisplayName: "report",
		Description: "report",
		Password:    "test",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}
	if err := reporterClient.Login(ctx, &isupipe.LoginRequest{
		UserName: reporter.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	if err := reporterClient.ReportLivecomment(ctx, livestream.ID, livecomment.ID); err != nil {
		return err
	}

	reports2, err := client.GetLivecommentReports(ctx, livestream.ID)
	if err != nil {
		return err
	}
	if len(reports2) != 1 {
		return fmt.Errorf("報告後のtest001ユーザのライブ配信におけるスパム報告は1件でなければなりません")
	}

	return nil
}

func NormalModerateLivecommentPretest(ctx context.Context, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	// moderateしたngwordが、GET ngwordsに含まれるか
	// 投稿済みのスパムライブコメントが、moderateによって粛清されているか
	// ライブコメントを投稿してきちんとエラーを返せているか
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: testUser.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	livestream := livestreams[0]

	ngwords, err := client.GetNgwords(ctx, livestream.ID)
	if err != nil {
		return err
	}
	if len(ngwords) != 0 {
		return fmt.Errorf("初期状態ではngwordはないはずです")
	}

	livecomments1, err := client.GetLivecomments(ctx, livestream.ID)
	if err != nil {
		return err
	}

	// スパム投稿
	spammerClient, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	_, err = spammerClient.Register(ctx, &isupipe.RegisterRequest{
		Name:        "spam",
		DisplayName: "spam",
		Description: "spam",
		Password:    "test",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}
	if err := spammerClient.Login(ctx, &isupipe.LoginRequest{
		UserName: "spam",
		Password: "test",
	}); err != nil {
		return err
	}

	spamComment := scheduler.LivecommentScheduler.GetNegativeComment()
	notip := &scheduler.Tip{}
	_, _, err = spammerClient.PostLivecomment(ctx, livestream.ID, spamComment.Comment, notip)
	if err != nil {
		return err
	}

	livecomments2, err := client.GetLivecomments(ctx, livestream.ID)
	if err != nil {
		return err
	}
	if len(livecomments2)-len(livecomments1) != 1 {
		return fmt.Errorf("１件ライブコメント(spam)が追加されたはずですが、件数が不正です")
	}

	// 粛清
	if err := client.Moderate(ctx, livestream.ID, spamComment.NgWord); err != nil {
		return err
	}

	livecomments3, err := client.GetLivecomments(ctx, livestream.ID)
	if err != nil {
		return err
	}
	if len(livecomments3)-len(livecomments1) != 0 {
		return fmt.Errorf("１件ライブコメントが粛清されたはずですが、件数が不正です")
	}

	return nil
}
