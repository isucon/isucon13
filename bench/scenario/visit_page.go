package scenario

import "context"

// 訪問時に行うGET操作をまとめた関数郡

func VisitTop(ctx context.Context) error {

}

// ライブ配信画面訪問
func VisitLivestream(ctx context.Context) error {

	// FIXME: 処理中定期的にGET /livestream/:livestreamid/livecomment を叩く

	return nil
}

// ライブ配信管理画面訪問
func VisitLivestreamAdmin(ctx context.Context) error {

	// ライブコメント一覧取得

	// NGワード一覧取得

	return nil
}
