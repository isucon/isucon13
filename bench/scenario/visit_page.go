package scenario

import (
	"context"

	"github.com/isucon/isucon13/bench/isupipe"
)

// 訪問時に行うGET操作をまとめた関数郡

func VisitTop(ctx context.Context, client *isupipe.Client) error {
	return nil
}

// ライブ配信画面訪問
func VisitLivestream(ctx context.Context, client *isupipe.Client, livestream *isupipe.Livestream) error {

	// FIXME: 統計情報取得

	// FIXME: 処理中定期的にGET /livestream/:livestreamid/livecomment を叩く

	// FIXME: ライブコメント投稿

	return nil
}

// ライブ配信管理画面訪問
func VisitLivestreamAdmin(ctx context.Context, client *isupipe.Client) error {

	// ライブコメント一覧取得

	// NGワード一覧取得

	return nil
}
