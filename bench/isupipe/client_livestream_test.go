package isupipe

import (
	"context"
	"testing"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/stretchr/testify/assert"
)

// FIXME: 予約期間、枠数などテスト

func TestLivestream(t *testing.T) {
	ctx := context.Background()

	client, err := NewClient(
		agent.WithBaseURL(config.TargetBaseURL),
	)
	assert.NoError(t, err)

	user := scheduler.UserScheduler.GetRandomStreamer()
	_, err = client.Register(ctx, &RegisterRequest{
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Description: user.Description,
		Password:    user.RawPassword,
		Theme: Theme{
			DarkMode: user.DarkMode,
		},
	})

	err = client.Login(ctx, &LoginRequest{
		UserName: user.Name,
		Password: user.RawPassword,
	})
	assert.NoError(t, err)

	livestreams, err := client.SearchLivestreams(ctx)
	assert.NoError(t, err)
	assert.NotZero(t, len(livestreams))

	// FIXME: 予約
	// * 期間チェック
	// * コラボレーター
	// * 予約枠

	// 期間外の予約がきちんとエラーで弾かれるかチェック

	// 同じユーザーで同じ時間帯の予約を実施して弾かれるかチェック

	// 同じ時間枠に枠数以上予約
	// ループでクライアントログインして同じ時間に予約
	// 同一ユーザは同一時間で１つしか予約を取れない（枠関係なく)ので、注意
}
