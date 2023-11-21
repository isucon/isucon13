package scenario

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/najeira/randstr"
)

//go:embed testdata/NoImage.jpg
var fallbackImage []byte

// icon_hashが反映されるまでに許される猶予
const IconHashAppliedDelay = 2 * time.Second

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
		Username: user.Name,
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

// FIXME PlaylistUrlとThumbnailUrlの確認
func checkPretestLivestream(subject string, livestream *isupipe.Livestream, title, description string, tags []int64, tagNames map[int64]string, startAt, endAt time.Time) error {
	// Check livestream
	if livestream.ID == 0 {
		return fmt.Errorf("%s livestreamのIDが正しくありません (actual:%d)", subject, livestream.ID)
	}
	if livestream.Title != title {
		return fmt.Errorf("%s livestreamのTitleが一致しません (expected:%s actual:%s)", subject, title, livestream.Title)
	}
	if livestream.Description != description {
		return fmt.Errorf("%s livestreamのDescriptionが一致しません (expected:%s actual:%s)", subject, description, livestream.Description)
	}
	if len(livestream.Tags) != len(tags) {
		return fmt.Errorf("%s livestreamのTagの数が一致しません (expected:%d actual:%d)", subject, len(tags), len(livestream.Tags))
	}
	for i := 0; i < len(tags); i++ {
		found := false
		for j := 0; j < len(livestream.Tags); j++ {
			n := tagNames[tags[i]]
			if tags[i] == livestream.Tags[j].ID && n == livestream.Tags[j].Name {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("%s livestreamにTag.IDがみつかりません (expected:%d)", subject, tags[i])
		}
	}
	if livestream.StartAt != startAt.Unix() {
		return fmt.Errorf("%s livestreamのStartAtが異なります (expected:%d actual:%d)", subject, startAt.Unix(), livestream.StartAt)
	}
	if livestream.EndAt != endAt.Unix() {
		return fmt.Errorf("%s livestreamのEndAtが異なります (expected:%d actual:%d)", subject, endAt.Unix(), livestream.EndAt)
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
		Username: testUser.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	tagResponse, err := client.GetTags(ctx)
	if err != nil {
		return err
	}

	tagNames := map[int64]string{}
	pretestTags := map[int64]int{}
	for _, tag := range tagResponse.Tags {
		tagNames[tag.ID] = tag.Name
		pretestTags[tag.ID] = 0
	}

	tags := []int64{1, 103}
	for len(tags) <= 10 {
		t := rand.Intn(len(tagResponse.Tags))
		tags = append(tags, tagResponse.Tags[t].ID)
		slices.Sort(tags)
		tags = slices.Compact(tags)
	}
	for _, t := range tags {
		pretestTags[t]++
	}

	var (
		startAt = time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local)
		endAt   = time.Date(2024, 4, 1, 1, 0, 0, 0, time.Local)
	)
	title := "pretest" + randstr.String(10)
	description := "pretest" + randstr.String(30)
	livestream, err := client.ReserveLivestream(ctx, testUser.Name, &isupipe.ReserveLivestreamRequest{
		Tags:        tags,
		Title:       title,
		Description: description,
		// FIXME: フロントで困らないようにちゃんとしたのを設定
		PlaylistUrl:  "",
		ThumbnailUrl: "",
		StartAt:      startAt.Unix(),
		EndAt:        endAt.Unix(),
	})
	if err != nil {
		return err
	}
	if err := checkPretestLivestream("予約した", livestream, title, description, tags, tagNames, startAt, endAt); err != nil {
		return err
	}
	//配信主
	if livestream.Owner.ID != testUser.ID {
		return fmt.Errorf("予約したlivestreamのuser.IDが異なります (expected:%d actual:%d)", testUser.ID, livestream.Owner.ID)
	}
	if livestream.Owner.DisplayName != "pretest user" {
		return fmt.Errorf("予約したlivestreamのuser.DisplayNameが異なります (expected:%s actual:%s)", "pretestuser", livestream.Owner.DisplayName)
	}

	gotLivestream, err := client.GetLivestream(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	if err := checkPretestLivestream("取得した", gotLivestream, title, description, tags, tagNames, startAt, endAt); err != nil {
		return err
	}

	if err := client.EnterLivestream(ctx, livestream.ID, livestream.Owner.Name); err != nil {
		return err
	}

	if err := client.ExitLivestream(ctx, livestream.ID, livestream.Owner.Name); err != nil {
		return err
	}

	{
		searchedStream, err := client.SearchLivestreams(ctx, isupipe.WithSearchTagQueryParam("椅子"))
		if err != nil {
			return err
		}
		if len(searchedStream) != len(scheduler.GetStreamIDsByTagID(103))+pretestTags[103] {
			return fmt.Errorf("「椅子」検索結果の数が一致しません (expected:%d actual:%d)", len(scheduler.GetStreamIDsByTagID(103))+pretestTags[103], len(searchedStream))
		}
		// FIXME: 初期データ生成により落ちるため、一時的にコメントアウト
		// if err := checkPretestLivestream("「椅子」検索結果1個目の", searchedStream[0], "ファッションライブ！夏のおすすめコーディネート", "この夏のおすすめのファッションコーディネートを紹介します。", []int64{103, 16}, tagNames, time.Unix(1690959600, 0), time.Unix(1690966800, 0)); err != nil {
		// 	return err
		// }

		if searchedStream[0].Owner.ID != 143 {
			return fmt.Errorf("「椅子」検索結果1個目のlivestreamのuser.IDが異なります (expected:%d actual:%d)", 143, searchedStream[0].Owner.ID)
		}
		wantSearchedStreamOwner, err := scheduler.UserScheduler.GetInitialUserForPretest(searchedStream[0].Owner.ID)
		if err != nil {
			return err
		}
		if searchedStream[0].Owner.DisplayName != wantSearchedStreamOwner.DisplayName {
			return fmt.Errorf("「椅子」検索結果1個目のlivestreamのuser.Nameが異なります (expected:%s actual:%s)", wantSearchedStreamOwner.DisplayName, searchedStream[0].Owner.DisplayName)
		}

		if err := checkPretestLivestream("「椅子」検索結果最後の", searchedStream[len(searchedStream)-1], title, description, tags, tagNames, startAt, endAt); err != nil {
			return err
		}
	}

	// もう一つ登録
	var (
		startAt2nd = time.Date(2024, 4, 1, 1, 0, 0, 0, time.Local)
		endAt2nd   = time.Date(2024, 4, 1, 2, 0, 0, 0, time.Local)
	)
	title2nd := "isutest" + randstr.String(10)
	description2nd := "isutest" + randstr.String(30)
	tags2nd := []int64{1, 2, 103}
	pretestTags[1]++
	pretestTags[2]++
	pretestTags[103]++

	livestream2nd, err := client.ReserveLivestream(ctx, testUser.Name, &isupipe.ReserveLivestreamRequest{
		Tags:        tags2nd,
		Title:       title2nd,
		Description: description2nd,
		// FIXME: フロントで困らないようにちゃんとしたのを設定
		PlaylistUrl:  "",
		ThumbnailUrl: "",
		StartAt:      startAt2nd.Unix(),
		EndAt:        endAt2nd.Unix(),
	})
	if err != nil {
		return err
	}
	if err := checkPretestLivestream("予約した", livestream2nd, title2nd, description2nd, tags2nd, tagNames, startAt2nd, endAt2nd); err != nil {
		return err
	}

	{
		//検索2回目
		searchedStream, err := client.SearchLivestreams(ctx, isupipe.WithSearchTagQueryParam("ライブ配信")) // ID:1
		if err != nil {
			return err
		}
		if len(searchedStream) != len(scheduler.GetStreamIDsByTagID(1))+pretestTags[1] {
			return fmt.Errorf("「ライブ配信」streamの検索結果の数が一致しません (expected:%d actual:%d)", len(scheduler.GetStreamIDsByTagID(1))+pretestTags[1], len(searchedStream))
		}

		if err := checkPretestLivestream("「ライブ配信」検索結果最後から2つ目の", searchedStream[len(searchedStream)-2], title, description, tags, tagNames, startAt, endAt); err != nil {
			return err
		}
		if err := checkPretestLivestream("「ライブ配信」検索結果最後の", searchedStream[len(searchedStream)-1], title2nd, description2nd, tags2nd, tagNames, startAt2nd, endAt2nd); err != nil {
			return err
		}
	}
	for i := 0; i < 3; i++ {
		//検索3-5回目 主に数があっているか確認
		tagID := int64(rand.Intn(len(tagResponse.Tags)))
		searchedStream, err := client.SearchLivestreams(ctx, isupipe.WithSearchTagQueryParam(tagNames[tagID]))
		if err != nil {
			return err
		}
		if len(searchedStream) != len(scheduler.GetStreamIDsByTagID(tagID))+pretestTags[tagID] {
			return fmt.Errorf("「%s」streamの検索結果の数が想定外です (expected:%d actual:%d)", tagNames[tagID], len(scheduler.GetStreamIDsByTagID(tagID))+pretestTags[tagID], len(searchedStream))
		}
		// FIXME もう少しチェックしたい
	}

	{
		// いくつか登録する
		for i := 0; i < 19; i++ {
			startAtExt := time.Date(2024, 4, 1, i+2, 0, 0, 0, time.Local)
			endAtExt := time.Date(2024, 4, 1, i+3, 0, 0, 0, time.Local)
			titleExt := "isutest" + randstr.String(10)
			descriptionExt := "isutest" + randstr.String(30)
			tagId := int64(rand.Intn(99)) + 1
			tagsExt := []int64{tagId, tagId + 1}
			livestreamExt, err := client.ReserveLivestream(ctx, testUser.Name, &isupipe.ReserveLivestreamRequest{
				Tags:        tagsExt,
				Title:       titleExt,
				Description: descriptionExt,
				// FIXME: フロントで困らないようにちゃんとしたのを設定
				PlaylistUrl:  "",
				ThumbnailUrl: "",
				StartAt:      startAtExt.Unix(),
				EndAt:        endAtExt.Unix(),
			})
			if err != nil {
				return err
			}
			if err := checkPretestLivestream("予約した", livestreamExt, titleExt, descriptionExt, tagsExt, tagNames, startAtExt, endAtExt); err != nil {
				return err
			}
		}
	}
	{
		// タグ指定なし検索
		searchedStream, err := client.SearchLivestreams(ctx, isupipe.WithLimitQueryParam(20))
		if err != nil {
			return err
		}
		if len(searchedStream) != 20 {
			return fmt.Errorf("limitありstreamの検索結果の数が想定外です (expected:%d actual:%d)", 20, len(searchedStream))
		}
		// FIXME もう少しチェックしたい
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
		Username: "test001",
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

	// アイコンを投稿する前、No Imageの画像のハッシュが返されているか
	me, err := client.GetMe(ctx)
	if err != nil {
		return err
	}
	if me.IconHash != fmt.Sprintf("%x", sha256.Sum256(fallbackImage)) {
		return fmt.Errorf("アイコン未設定の場合は、icon_hashはNoImage.jpgのハッシュ値を返さなければなりません")
	}

	// アイコンを投稿後、期待するアイコンが設定されているか
	randomIcon := scheduler.IconSched.GetRandomIcon()
	if _, err := client.PostIcon(ctx, &isupipe.PostIconRequest{
		Image: randomIcon.Image,
	}); err != nil {
		return err
	}

	// 反映されるまでに許される猶予
	time.Sleep(IconHashAppliedDelay)
	me2, err := client.GetMe(ctx)
	if err != nil {
		return err
	}
	if me2.IconHash != fmt.Sprintf("%x", randomIcon.Hash) {
		return fmt.Errorf("新たに設定したアイコンのハッシュ値がicon_hashに反映されていません")
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
		Username: testUser.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	if len(livestreams) == 0 {
		return fmt.Errorf("自分のライブ配信が存在しません")
	}

	livestream := livestreams[rand.Intn(len(livestreams))] // ランダムに選ぶ
	if livestream.Owner.ID != testUser.ID {
		return fmt.Errorf("自分がownerではないlivestreamが返されました expected:%s got:%s", testUser.Name, livestream.Owner.Name)
	}

	if _, err = client.GetLivecommentReports(ctx, livestream.ID); err != nil {
		return err
	}

	notip := &scheduler.Tip{}
	postedLiveComment, _, err := client.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, "test", notip)
	if err != nil {
		return err
	}

	// アイコンを投稿してLivecommentの中のicon_hashが更新されているかをみる
	randomIcon := scheduler.IconSched.GetRandomIcon()
	if _, err := client.PostIcon(ctx, &isupipe.PostIconRequest{
		Image: randomIcon.Image,
	}); err != nil {
		return err
	}
	// icon反映されるまでに許される猶予
	time.Sleep(IconHashAppliedDelay)

	livecomments, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name, isupipe.WithLimitQueryParam(1))
	if err != nil {
		return err
	}
	var found bool
	for _, livecomment := range livecomments {
		if livecomment.ID != postedLiveComment.ID {
			continue
		}
		if livecomment.User.ID != testUser.ID {
			return fmt.Errorf("投稿したライブコメントのuser.IDが正しくありません")
		}
		if livecomment.User.IconHash != fmt.Sprintf("%x", randomIcon.Hash) {
			return fmt.Errorf("新たに設定したアイコンのハッシュ値がicon_hashに反映されていません")
		}
		found = true
		break
	}
	if !found {
		return fmt.Errorf("投稿したライブコメントが見つかりません")
	}

	if err := client.ReportLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecomments[0].ID); err != nil {
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
		Username: testUser.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}

	livestream := livestreams[0]

	if _, err := client.PostReaction(ctx, livestream.ID, livestream.Owner.Name, &isupipe.PostReactionRequest{
		EmojiName: "chair",
	}); err != nil {
		return err
	}

	reactions, err := client.GetReactions(ctx, livestream.ID, livestream.Owner.Name)
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
		Username: "test001",
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

	livecomments, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name, isupipe.WithLimitQueryParam(10))
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
		Username: reporter.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	if err := reporterClient.ReportLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecomment.ID); err != nil {
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
		Username: testUser.Name,
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

	livecomments1, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name)
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
		Username: "spam",
		Password: "test",
	}); err != nil {
		return err
	}

	spamComment, _ := scheduler.LivecommentScheduler.GetNegativeComment()
	notip := &scheduler.Tip{}
	_, _, err = spammerClient.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, spamComment.Comment, notip)
	if err != nil {
		return err
	}

	livecomments2, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name)
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
	scheduler.LivecommentScheduler.Moderate(spamComment.Comment)

	livecomments3, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	if len(livecomments3)-len(livecomments1) != 0 {
		return fmt.Errorf("１件ライブコメントが粛清されたはずですが、件数が不正です")
	}

	return nil
}
