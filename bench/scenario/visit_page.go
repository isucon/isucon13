package scenario

import (
	"context"

	"github.com/isucon/isucon13/bench/isupipe"
)

// 訪問時に行うGET操作をまとめた関数郡

func VisitTop(ctx context.Context, client *isupipe.Client) error {
	// FIXME: プロフィールアイコン取得

	// FIXME: 10件程度ライブストリーム取得
	_, err := client.SearchLivestreams(ctx)
	if err != nil {
		return err
	}

	// FIXME: 検索

	return nil
}

// ライブ配信画面訪問
func VisitLivestream(ctx context.Context, client *isupipe.Client, livestream *isupipe.Livestream) error {

	if err := client.EnterLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	// FIXME: 統計情報取得
	_, err := client.GetLivestreamStatistics(ctx, livestream.ID)
	if err != nil {
		return err
	}

	// FIXME: 処理中定期的にGET /livestream/:livestreamid/livecomment を叩く

	return nil
}

func GoAwayFromLivestream(ctx context.Context, client *isupipe.Client, livestream *isupipe.Livestream) error {

	if err := client.ExitLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	return nil
}

// ライブ配信管理画面訪問
func VisitLivestreamAdmin(ctx context.Context, client *isupipe.Client) error {

	// ライブコメント一覧取得
	// FIXME: 自分のライブストリーム一覧を取ってくる必要がある
	_, err := client.SearchLivestreams(ctx)
	if err != nil {
		return err
	}

	// for _, livestream := range livestreams {
	// 	livestreamID := livestream.ID
	// 	if _, err := client.GetLivecommentReports(ctx, livestreamID); err != nil {
	// 		return err
	// 	}
	// }

	// NGワード一覧取得
	// FIXME: webapp側にこのエンドポイントがないので実装から

	return nil
}
