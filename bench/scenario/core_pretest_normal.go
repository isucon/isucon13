package scenario

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/najeira/randstr"
	"go.uber.org/zap"
)

//go:embed testdata/NoImage.jpg
var fallbackImage []byte

// icon_hashが反映されるまでに許される猶予
const IconHashAppliedDelay = 2 * time.Second

// 基本機能のロジックpretest

func NormalUserPretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
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
		Password:    "test",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: user.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	if u, err := client.GetMe(ctx); err != nil {
		return err
	} else {
		if u.ID != user.ID {
			return fmt.Errorf("ログインしたユーザのIDが正しくありません (expected:%d actual:%d)", user.ID, u.ID)
		}
		if u.Name != user.Name {
			return fmt.Errorf("ログインしたユーザのNameが正しくありません (expected:%s actual:%s)", user.Name, u.Name)
		}
		if u.DisplayName != user.DisplayName {
			return fmt.Errorf("ログインしたユーザのDisplayNameが正しくありません (expected:%s actual:%s)", user.DisplayName, u.DisplayName)
		}
		if u.Theme.DarkMode != user.Theme.DarkMode {
			return fmt.Errorf("ログインしたユーザのThemeが正しくありません (expected:%v actual:%v)", user.Theme, u.Theme)
		}
	}

	if _, err := client.GetUser(ctx, user.Name); err != nil {
		return err
	}

	if t, err := client.GetStreamerTheme(ctx, user); err != nil {
		return err
	} else {
		if t.DarkMode != user.Theme.DarkMode {
			return fmt.Errorf("ユーザのThemeが正しくありません (expected:%v actual:%v)", user.Theme, t)
		}
	}

	return nil
}

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
	if tags != nil {
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
	}
	if strings.Index(livestream.PlaylistUrl, "https://media.xiii.isucon.dev") != 0 {
		return fmt.Errorf("%s livestreamのPlaylistUrlが正しくありません (actual:%s)", subject, livestream.PlaylistUrl)
	}
	if strings.Index(livestream.ThumbnailUrl, "https://media.xiii.isucon.dev") != 0 {
		return fmt.Errorf("%s livestreamのThumbnailUrlが正しくありません (actual:%s)", subject, livestream.ThumbnailUrl)
	}
	if livestream.StartAt != startAt.Unix() {
		return fmt.Errorf("%s livestreamのStartAtが異なります (expected:%d actual:%d)", subject, startAt.Unix(), livestream.StartAt)
	}
	if livestream.EndAt != endAt.Unix() {
		return fmt.Errorf("%s livestreamのEndAtが異なります (expected:%d actual:%d)", subject, endAt.Unix(), livestream.EndAt)
	}
	return nil
}

func NormalLivestreamPretest(ctx context.Context, contestantLogger *zap.Logger, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	// 機能的なテスト
	// 予約したライブ配信が一覧に見えるか、取得できるか、検索によって見つけられるか
	// enter/exitできるか (other)

	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	clientNoSession, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: testUser.Name,
		Password: defaultPasswordOrPretest(testUser.Name),
	}); err != nil {
		return err
	}

	tagResponse, err := client.GetTags(ctx)
	if err != nil {
		return err
	}

	if len(tagResponse.Tags) != scheduler.GetTagPoolLength() {
		return fmt.Errorf("取得した tag 一覧の数が正しくありません (expected:%d actual:%d)", scheduler.GetTagPoolLength(), len(tagResponse.Tags))
	}

	tagNames := map[int64]string{}
	pretestTags := map[int64]int{}
	for _, tag := range tagResponse.Tags {
		tagNames[tag.ID] = tag.Name
		pretestTags[tag.ID] = 0
	}

	// 全部のタグを検査
	poolTags := scheduler.GetTagsMap()
	if !reflect.DeepEqual(tagNames, poolTags) {
		return fmt.Errorf("取得した tag 一覧が正しくありません。過不足があります")
	}

	// userドメインでのgetTag
	{
		tagr, err := client.GetTagsWithUser(ctx, testUser.Name)
		if err != nil {
			return err
		}
		tagn := map[int64]string{}
		for _, tag := range tagr.Tags {
			tagn[tag.ID] = tag.Name
		}
		if !reflect.DeepEqual(tagn, poolTags) {
			return fmt.Errorf("取得した tag 一覧が正しくありません。過不足があります")
		}
	}

	{
		tagr, err := clientNoSession.GetTags(ctx)
		if err != nil {
			return err
		}
		tagn := map[int64]string{}
		for _, tag := range tagr.Tags {
			tagn[tag.ID] = tag.Name
		}
		if !reflect.DeepEqual(tagn, poolTags) {
			return fmt.Errorf("取得した tag 一覧が正しくありません。過不足があります")
		}
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

	reserveStreams := []int64{}

	var (
		startAt = time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local)
		endAt   = time.Date(2024, 4, 1, 1, 0, 0, 0, time.Local)
	)
	title := randstr.String(19)
	description := randstr.String(29)
	livestream, err := client.ReserveLivestream(ctx, testUser.Name, &isupipe.ReserveLivestreamRequest{
		Tags:         tags,
		Title:        title,
		Description:  description,
		PlaylistUrl:  "https://media.xiii.isucon.dev/api/4/playlist.m3u8",
		ThumbnailUrl: "https://media.xiii.isucon.dev/isucon12_final.webp",
		StartAt:      startAt.Unix(),
		EndAt:        endAt.Unix(),
	})
	if err != nil {
		return err
	}
	if err := checkPretestLivestream("予約した", livestream, title, description, tags, tagNames, startAt, endAt); err != nil {
		return err
	}
	reserveStreams = append([]int64{livestream.ID}, reserveStreams...)

	//配信主
	if livestream.Owner.ID != testUser.ID {
		return fmt.Errorf("予約したlivestreamのuser.IDが異なります (expected:%d actual:%d)", testUser.ID, livestream.Owner.ID)
	}
	if livestream.Owner.DisplayName != PreTestDisplayName {
		return fmt.Errorf("予約したlivestreamのuser.DisplayNameが異なります (expected:%s actual:%s)", PreTestDisplayName, livestream.Owner.DisplayName)
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

		// 1個目は今登録したやつ
		if err := checkPretestLivestream("「椅子」検索結果1個目の", searchedStream[0], title, description, tags, tagNames, startAt, endAt); err != nil {
			return err
		}
		if livestream.Owner.ID != testUser.ID {
			return fmt.Errorf("予約したlivestreamのuser.IDが異なります (expected:%d actual:%d)", testUser.ID, livestream.Owner.ID)
		}
		if livestream.Owner.DisplayName != testUser.DisplayName {
			return fmt.Errorf("予約したlivestreamのuser.DisplayNameが異なります (expected:%s actual:%s)", testUser.DisplayName, livestream.Owner.DisplayName)
		}

		// ランダムn個目
		for i := 0; i < 5; i++ {
			tagPool := scheduler.GetStreamIDsByTagID(103)
			randNumber := rand.Intn(50) + pretestTags[103]
			livestreamID := tagPool[len(tagPool)-randNumber-1]
			if searchedStream[randNumber+pretestTags[103]].ID != livestreamID {
				return fmt.Errorf("「椅子」検索結果の%d番目のlivestream.idが一致しません (expected:%d actual:%d)", randNumber, livestreamID, searchedStream[randNumber+pretestTags[103]].ID)
			}
			livestream := scheduler.GetLivestreamByID(livestreamID)
			if err := checkPretestLivestream(fmt.Sprintf("「椅子」検索結果の%d番目の", randNumber), searchedStream[randNumber+pretestTags[103]], livestream.Title, livestream.Description, scheduler.GetTagIDsByStreamID(livestreamID), tagNames, time.Unix(livestream.StartAt, 0), time.Unix(livestream.EndAt, 0)); err != nil {
				return err
			}
			livestreamOwner := scheduler.GetInitialUserByID(livestream.OwnerID)
			if searchedStream[randNumber+pretestTags[103]].Owner.DisplayName != livestreamOwner.DisplayName {
				return fmt.Errorf("「椅子」検索結果の%d番目のlivestreamのuser.DisplayNameが異なります (expected:%s actual:%s)", randNumber, searchedStream[randNumber+pretestTags[103]].Owner.DisplayName, livestreamOwner.DisplayName)
			}
		}
	}

	// もう一つ登録
	var (
		startAt2nd = time.Date(2024, 4, 1, 1, 0, 0, 0, time.Local)
		endAt2nd   = time.Date(2024, 4, 1, 2, 0, 0, 0, time.Local)
	)
	title2nd := randstr.String(12)
	description2nd := randstr.String(36)
	tags2nd := []int64{1, 2, 103}
	pretestTags[1]++
	pretestTags[2]++
	pretestTags[103]++

	livestream2nd, err := client.ReserveLivestream(ctx, testUser.Name, &isupipe.ReserveLivestreamRequest{
		Tags:         tags2nd,
		Title:        title2nd,
		Description:  description2nd,
		PlaylistUrl:  "https://media.xiii.isucon.dev/api/4/playlist.m3u8",
		ThumbnailUrl: "https://media.xiii.isucon.dev/isucon12_final.webp",
		StartAt:      startAt2nd.Unix(),
		EndAt:        endAt2nd.Unix(),
	})
	if err != nil {
		return err
	}
	if err := checkPretestLivestream("予約した", livestream2nd, title2nd, description2nd, tags2nd, tagNames, startAt2nd, endAt2nd); err != nil {
		return err
	}
	reserveStreams = append([]int64{livestream2nd.ID}, reserveStreams...)

	{
		//検索2回目
		searchedStream, err := clientNoSession.SearchLivestreams(ctx, isupipe.WithSearchTagQueryParam("ライブ配信")) // ID:1
		if err != nil {
			return err
		}
		if len(searchedStream) != len(scheduler.GetStreamIDsByTagID(1))+pretestTags[1] {
			return fmt.Errorf("「ライブ配信」streamの検索結果の数が一致しません (expected:%d actual:%d)", len(scheduler.GetStreamIDsByTagID(1))+pretestTags[1], len(searchedStream))
		}

		if err := checkPretestLivestream("「ライブ配信」検索結果2番目", searchedStream[1], title, description, tags, tagNames, startAt, endAt); err != nil {
			return err
		}
		if err := checkPretestLivestream("「ライブ配信」検索結果最初の", searchedStream[0], title2nd, description2nd, tags2nd, tagNames, startAt2nd, endAt2nd); err != nil {
			return err
		}
	}

	{
		// いくつか登録する
		for i := 0; i < 19; i++ {
			startAtExt := time.Date(2024, 4, 1, i+2, 0, 0, 0, time.Local)
			endAtExt := time.Date(2024, 4, 1, i+3, 0, 0, 0, time.Local)
			titleExt := randstr.String(17 + rand.Intn(19))
			descriptionExt := randstr.String(51 + rand.Intn(19))
			tagId := int64(rand.Intn(99)) + 1
			tagsExt := []int64{tagId, tagId + 1}
			pretestTags[tagId]++
			pretestTags[tagId+1]++
			livestreamExt, err := client.ReserveLivestream(ctx, testUser.Name, &isupipe.ReserveLivestreamRequest{
				Tags:         tagsExt,
				Title:        titleExt,
				Description:  descriptionExt,
				PlaylistUrl:  "https://media.xiii.isucon.dev/api/4/playlist.m3u8",
				ThumbnailUrl: "https://media.xiii.isucon.dev/isucon12_final.webp",
				StartAt:      startAtExt.Unix(),
				EndAt:        endAtExt.Unix(),
			})
			if err != nil {
				return err
			}
			if err := checkPretestLivestream("予約した", livestreamExt, titleExt, descriptionExt, tagsExt, tagNames, startAtExt, endAtExt); err != nil {
				return err
			}
			reserveStreams = append([]int64{livestreamExt.ID}, reserveStreams...)
		}
	}

	for i := 0; i < 7; i++ {
		tagID := int64(rand.Intn(len(tagResponse.Tags))) + 1
		searchedStream, err := client.SearchLivestreams(ctx, isupipe.WithSearchTagQueryParam(tagNames[tagID]))
		if err != nil {
			return err
		}
		if len(searchedStream) != len(scheduler.GetStreamIDsByTagID(tagID))+pretestTags[tagID] {
			return fmt.Errorf("「%s」streamの検索結果の数が想定外です (expected:%d actual:%d)", tagNames[tagID], len(scheduler.GetStreamIDsByTagID(tagID))+pretestTags[tagID], len(searchedStream))
		}

		tagPool := scheduler.GetStreamIDsByTagID(tagID)
		randNumber := rand.Intn(50) + pretestTags[tagID]
		livestreamID := tagPool[len(tagPool)-randNumber-1]
		if searchedStream[randNumber+pretestTags[tagID]].ID != livestreamID {
			return fmt.Errorf("「%s」検索結果の%d番目のlivestream.idが一致しません (expected:%d actual:%d)", tagNames[tagID], randNumber+1, livestreamID, searchedStream[randNumber+pretestTags[tagID]].ID)
		}
		livestream := scheduler.GetLivestreamByID(livestreamID)
		if err := checkPretestLivestream(fmt.Sprintf("「%s」検索結果の%d番目の", tagNames[tagID], randNumber+1), searchedStream[randNumber+pretestTags[tagID]], livestream.Title, livestream.Description, scheduler.GetTagIDsByStreamID(livestreamID), tagNames, time.Unix(livestream.StartAt, 0), time.Unix(livestream.EndAt, 0)); err != nil {
			return err
		}
		livestreamOwner := scheduler.GetInitialUserByID(livestream.OwnerID)
		if searchedStream[randNumber+pretestTags[tagID]].Owner.DisplayName != livestreamOwner.DisplayName {
			return fmt.Errorf("「%s」検索結果の%d番目のlivestreamのuser.DisplayNameが異なります (expected:%s actual:%s)", tagNames[tagID], randNumber+1, searchedStream[randNumber+pretestTags[tagID]].Owner.DisplayName, livestreamOwner.DisplayName)
		}

	}

	{
		// タグ指定なし検索
		searchedStream, err := client.SearchLivestreams(ctx, isupipe.WithLimitQueryParam(config.NumSearchLivestreams))
		if err != nil {
			return err
		}
		if len(searchedStream) != config.NumSearchLivestreams {
			return fmt.Errorf("タグ指定なし検索結果の数が想定外です (expected:%d actual:%d)", config.NumSearchLivestreams, len(searchedStream))
		}
		for i := 0; i < 5; i++ {
			randNumber := rand.Intn(20)
			if searchedStream[randNumber].ID != reserveStreams[randNumber] {
				return fmt.Errorf("タグ指定なし検索結果の%d番目のlivestream.idが一致しません (expected:%d actual:%d)", randNumber+1, reserveStreams[randNumber], searchedStream[randNumber].ID)
			}
		}
	}
	{
		// タグ指定なし検索
		searchedStream, err := clientNoSession.SearchLivestreams(ctx, isupipe.WithLimitQueryParam(config.NumSearchLivestreams))
		if err != nil {
			return err
		}
		if len(searchedStream) != config.NumSearchLivestreams {
			return fmt.Errorf("タグ指定なし検索結果の数が想定外です (expected:%d actual:%d)", config.NumSearchLivestreams, len(searchedStream))
		}
		for i := 0; i < 5; i++ {
			randNumber := rand.Intn(20)
			if searchedStream[randNumber].ID != reserveStreams[randNumber] {
				return fmt.Errorf("タグ指定なし検索結果の%d番目のlivestream.idが一致しません (expected:%d actual:%d)", randNumber+1, reserveStreams[randNumber], searchedStream[randNumber].ID)
			}
		}
		for i := 0; i < 5; i++ {
			randNumber := rand.Intn(20) + 25
			livestreamID := int64(scheduler.GetLivestreamLength()+len(reserveStreams)-randNumber) + 1
			if searchedStream[randNumber].ID != livestreamID {
				return fmt.Errorf("タグ指定なし検索結果の%d番目のlivestream.idが一致しません (expected:%d actual:%d)", randNumber+1, livestreamID, searchedStream[randNumber].ID)
			}
			livestream := scheduler.GetLivestreamByID(livestreamID)
			if err := checkPretestLivestream(fmt.Sprintf("タグ指定なし検索結果の%d番目の", randNumber+1), searchedStream[randNumber], livestream.Title, livestream.Description, scheduler.GetTagIDsByStreamID(livestreamID), tagNames, time.Unix(livestream.StartAt, 0), time.Unix(livestream.EndAt, 0)); err != nil {
				return err
			}
			livestreamOwner := scheduler.GetInitialUserByID(livestream.OwnerID)
			if searchedStream[randNumber].Owner.DisplayName != livestreamOwner.DisplayName {
				return fmt.Errorf("タグ指定なし検索結果の%d番目のlivestreamのuser.DisplayNameが異なります (expected:%s actual:%s)", randNumber+1, searchedStream[randNumber].Owner.DisplayName, livestreamOwner.DisplayName)
			}
		}
	}

	return nil
}

func NormalIconPretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
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

	// アイコンを投稿する前
	if ls, err := client.GetMyLivestreams(ctx); err != nil {
		return err
	} else {
		for _, l := range ls {
			if l.Owner.IconHash != fmt.Sprintf("%x", sha256.Sum256(fallbackImage)) {
				return fmt.Errorf("アイコン未設定の場合は、livestreamのicon_hashはNoImage.jpgのハッシュ値を返さなければなりません")
			}
		}
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

	icon2, err := client.GetIcon(ctx, "test001") // etagなし
	if err != nil {
		return err
	}
	icon2Hash := sha256.Sum256(icon2)
	if !bytes.Equal(icon2Hash[:], randomIcon.Hash[:]) {
		return fmt.Errorf("新たに設定したアイコンが反映されていません")
	}

	// マッチするetag付きでリクエストする(レスポンスは200でも304でもどっちでもOK)
	_, err = client.GetIcon(ctx, "test001", isupipe.WithETag(me2.IconHash))
	if err != nil {
		return err
	}

	// マッチしないetag付きでリクエストする(bodyが一致しないといけない)
	icon3, err := client.GetIcon(ctx, "test001", isupipe.WithETag("abcdef0123456890"))
	if err != nil {
		return err
	}
	icon3Hash := sha256.Sum256(icon3)
	if !bytes.Equal(icon3Hash[:], randomIcon.Hash[:]) {
		return fmt.Errorf("設定したアイコンが反映されていません")
	}

	// アイコンを投稿後、期待するアイコンが設定されているか
	if ls, err := client.GetMyLivestreams(ctx); err != nil {
		return err
	} else {
		for _, l := range ls {
			if l.Owner.IconHash != fmt.Sprintf("%x", randomIcon.Hash) {
				return fmt.Errorf("設定したアイコンがlivestreamに反映されていません")
			}
		}
	}

	return nil
}

func NormalPostLivecommentPretest(ctx context.Context, contestantLogger *zap.Logger, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: testUser.Name,
		Password: defaultPasswordOrPretest(testUser.Name),
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
		return fmt.Errorf("自分がownerではないlivestreamが返されました expected:%s actual:%s", testUser.Name, livestream.Owner.Name)
	}

	if reports, err := client.GetLivecommentReports(ctx, livestream.ID, livestream.Owner.Name); err != nil {
		return err
	} else {
		for _, r := range reports {
			if r.Livecomment.Livestream.Owner.ID != testUser.ID {
				return fmt.Errorf("自分がownerではないlivestreamのスパム報告が返されました expected:%s actual:%s", testUser.Name, r.Livecomment.Livestream.Owner.Name)
			}
			client.GetIcon(ctx, r.Livecomment.User.Name, isupipe.WithETag(r.Livecomment.User.IconHash))
			// icon取得のエラーは無視
		}
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

	if err := client.ReportLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecomments[0].ID, isupipe.WithValidateReportLivecomment()); err != nil {
		return err
	}

	return nil
}

func NormalReactionPretest(ctx context.Context, contestantLogger *zap.Logger, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	// 投稿したリアクションがGETできるか
	// limitをつけられるか
	// 初期データが期待する件数あるか
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: testUser.Name,
		Password: defaultPasswordOrPretest(testUser.Name),
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
	for _, livestream := range livestreams {
		reactions1, err := client.GetReactions(ctx, livestream.ID, livestream.Owner.Name)
		if err != nil {
			return err
		}

		if r, err := client.PostReaction(ctx, livestream.ID, livestream.Owner.Name, &isupipe.PostReactionRequest{
			EmojiName: "chair",
		}); err != nil {
			return err
		} else {
			if r.Livestream.ID != livestream.ID {
				return fmt.Errorf("投稿されたリアクションのlivestream.IDが正しくありません expected:%d actual:%d", livestream.ID, r.Livestream.ID)
			}
			if r.Livestream.Owner.Name != livestream.Owner.Name {
				return fmt.Errorf("投稿されたリアクションのOwnerが正しくありません expected:%s actual:%s", livestream.Owner.Name, r.Livestream.Owner.Name)
			}
			if r.EmojiName != "chair" {
				return fmt.Errorf("投稿されたリアクションのEmojiNameが正しくありません expected:%s actual:%s", "chair", r.EmojiName)
			}
			if r.User.Name != testUser.Name {
				return fmt.Errorf("投稿されたリアクションのUserが正しくありません expected:%s actual:%s", testUser.Name, r.User.Name)
			}
		}

		reactions2, err := client.GetReactions(ctx, livestream.ID, livestream.Owner.Name)
		if err != nil {
			return err
		}
		if len(reactions2)-len(reactions1) != 1 {
			return fmt.Errorf("リアクション件数が不正です")
		}
	}
	return nil
}

func NormalReportLivecommentPretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	// ライブコメントを1件取得(limit=1)
	// ライブコメントを報告できるか (other)
	// 報告したものが確認できるか (owner)

	// 初期で報告が0件

	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
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

	reports, err := client.GetLivecommentReports(ctx, livestream.ID, livestream.Owner.Name)
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
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%srpt", randstr.String(11))
	passwd := randstr.String(13)
	reporter, err := reporterClient.Register(ctx, &isupipe.RegisterRequest{
		Name:        name,
		DisplayName: randDisplayName(),
		Description: "report",
		Password:    passwd,
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}
	if err := reporterClient.Login(ctx, &isupipe.LoginRequest{
		Username: reporter.Name,
		Password: passwd,
	}); err != nil {
		return err
	}

	if err := reporterClient.ReportLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecomment.ID, isupipe.WithValidateReportLivecomment()); err != nil {
		return err
	}

	reports2, err := client.GetLivecommentReports(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	if len(reports2) != 1 {
		return fmt.Errorf("報告後のtest001ユーザのライブ配信におけるスパム報告は1件でなければなりません")
	}

	return nil
}

func NormalModerateLivecommentPretest(ctx context.Context, contestantLogger *zap.Logger, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	// moderateしたngwordが、GET ngwordsに含まれるか
	// 投稿済みのスパムライブコメントが、moderateによって粛清されているか
	// ライブコメントを投稿してきちんとエラーを返せているか
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: testUser.Name,
		Password: defaultPasswordOrPretest(testUser.Name),
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
	for _, livestream := range livestreams {
		if livestream.Owner.ID != testUser.ID {
			return fmt.Errorf("自分がownerではないlivestreamが返されました expected:%s actual:%s", testUser.Name, livestream.Owner.Name)
		}
	}
	livestream := livestreams[rand.Intn(len(livestreams))] // ランダムに選ぶ

	ngwords, err := client.GetNgwords(ctx, livestream.ID, livestream.Owner.Name)
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
		contestantLogger,
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%sspm", randstr.String(11))
	passwd := randstr.String(18)
	_, err = spammerClient.Register(ctx, &isupipe.RegisterRequest{
		Name:        name,
		DisplayName: randDisplayName(),
		Description: `普段アナウンサーをしています。
よろしくおねがいします！

連絡は以下からお願いします。

ウェブサイト: http://eishikawa.example.com/
メールアドレス: eishikawa@example.com
`,
		Password: passwd,
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}
	if err := spammerClient.Login(ctx, &isupipe.LoginRequest{
		Username: name,
		Password: passwd,
	}); err != nil {
		return err
	}

	added := 0
	for i := 0; i <= 5; i++ {
		// spamではない普通のコメントをする
		livecomment := scheduler.LivecommentScheduler.GetLongPositiveComment()
		r, _, err := spammerClient.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, livecomment.Comment, &scheduler.Tip{Tip: 100})
		if err != nil {
			return err
		}
		if r.Livestream.ID != livestream.ID {
			return fmt.Errorf("投稿されたライブコメントのlivestream.IDが正しくありません expected:%d actual:%d", livestream.ID, r.Livestream.ID)
		}
		if r.Livestream.Owner.Name != livestream.Owner.Name {
			return fmt.Errorf("投稿されたライブコメントのOwnerが正しくありません expected:%s actual:%s", livestream.Owner.Name, r.Livestream.Owner.Name)
		}
		if r.Tip != int64(100) {
			return fmt.Errorf("投稿されたライブコメントのTipが正しくありません expected:%d actual:%d", 100, r.Tip)
		}
		added++
	}

	spamComment, _ := scheduler.LivecommentScheduler.GetNegativeComment()
	notip := &scheduler.Tip{}
	_, _, err = spammerClient.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, spamComment.Comment, notip)
	if err != nil {
		return err
	}
	added++

	livecomments2, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	if len(livecomments2)-len(livecomments1) != added {
		return fmt.Errorf("%d件ライブコメントが追加されたはずですが、件数が不正です", added)
	}

	// 粛清
	if err := client.Moderate(ctx, livestream.ID, livestream.Owner.Name, spamComment.NgWord); err != nil {
		return err
	}
	scheduler.LivecommentScheduler.Moderate(spamComment.Comment)

	livecomments3, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}
	if len(livecomments3)-len(livecomments2) != -1 {
		return fmt.Errorf("１件ライブコメントが削除されたはずですが、件数が不正です")
	}
	for _, comment := range livecomments3 {
		if strings.Contains(comment.Comment, spamComment.NgWord) {
			return fmt.Errorf("削除されたはずのライブコメントが残っています")
		}
	}

	// ngwordに登録されたので投稿できないはず
	_, _, err = spammerClient.PostLivecomment(ctx, livestream.ID, livestream.Owner.Name, spamComment.Comment, notip)
	if err == nil {
		return fmt.Errorf("ngwordに登録されたはずのライブコメントが投稿できています")
	}

	return nil
}
