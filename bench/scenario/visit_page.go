package scenario

import (
	"context"

	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

// 訪問時に行うGET操作をまとめた関数郡

func VisitTop(ctx context.Context, contestantLogger *zap.Logger, client *isupipe.Client) error {
	if _, err := client.GetMyIcon(ctx); err != nil {
		return err
	}

	livestreams, err := client.SearchLivestreams(ctx, isupipe.WithLimitQueryParam(config.NumSearchLivestreams))
	if err != nil {
		return err
	}
	for _, livestream := range livestreams {
		client.GetIcon(ctx, livestream.Owner.Name, isupipe.WithETag(livestream.Owner.IconHash))
		// iconの取得失敗は無視
	}

	tags, err := client.GetRandomSearchTags(ctx, 1)
	if err != nil {
		return err
	}

	if livestreams, err := client.SearchLivestreams(ctx, isupipe.WithSearchTagQueryParam(tags[0])); err != nil {
		return err
	} else {
		for _, livestream := range livestreams {
			client.GetIcon(ctx, livestream.Owner.Name, isupipe.WithETag(livestream.Owner.IconHash))
			// iconの取得失敗は無視
		}
	}

	return nil
}

// ライブ配信画面訪問
func VisitLivestream(ctx context.Context, contestantLogger *zap.Logger, client *isupipe.Client, livestream *isupipe.Livestream) error {

	if err := client.EnterLivestream(ctx, livestream.ID, livestream.Owner.Name); err != nil {
		return err
	}

	/*
		// NOTE: viwersのほうで呼ぶのでこっちは一旦やめます
		_, err := client.GetLivestreamStatistics(ctx, livestream.ID, livestream.Owner.Name)
		if err != nil {
			return err
		}
	*/

	livecomments, err := client.GetLivecomments(ctx, livestream.ID, livestream.Owner.Name, isupipe.WithLimitQueryParam(10))
	if err != nil {
		return err
	}
	for _, livecomment := range livecomments {
		client.GetIcon(ctx, livecomment.User.Name, isupipe.WithETag(livecomment.User.IconHash))
		// iconの取得失敗は無視
	}

	_, err = client.GetReactions(ctx, livestream.ID, livestream.Owner.Name, isupipe.WithLimitQueryParam(10))
	if err != nil {
		return err
	}

	return nil
}

func LeaveFromLivestream(ctx context.Context, contestantLogger *zap.Logger, client *isupipe.Client, livestream *isupipe.Livestream) error {

	if err := client.ExitLivestream(ctx, livestream.ID, livestream.Owner.Name); err != nil {
		return err
	}

	return nil
}

// ライブ配信管理画面訪問
func VisitLivestreamAdmin(ctx context.Context, contestantLogger *zap.Logger, client *isupipe.Client, livestream *isupipe.Livestream) error {

	// ライブコメント一覧取得
	// FIXME: 自分のライブストリーム一覧を取ってくる必要がある
	_, err := client.SearchLivestreams(ctx, isupipe.WithLimitQueryParam(config.NumSearchLivestreams))
	if err != nil {
		return err
	}

	_, err = client.GetNgwords(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}

	_, err = client.GetLivecommentReports(ctx, livestream.ID, livestream.Owner.Name)
	if err != nil {
		return err
	}

	return nil
}

func VisitUserProfile(ctx context.Context, contestantLogger *zap.Logger, client *isupipe.Client, user *isupipe.User) error {
	if _, err := client.GetStreamerTheme(ctx, user); err != nil {
		return err
	}

	if _, err := client.GetIcon(ctx, user.Name, isupipe.WithETag(user.IconHash)); err != nil {
		return err
	}

	if _, err := client.GetUserStatistics(ctx, user.Name); err != nil {
		return err
	}

	if _, err := client.GetUserLivestreams(ctx, user.Name); err != nil {
		return err
	}

	return nil
}
