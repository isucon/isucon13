package isupipe

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

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
		agent.WithTimeout(1*time.Minute),
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
	_, err = client.ReserveLivestream(ctx, &ReserveLivestreamRequest{
		Title:        "livestream-test",
		Description:  "livestream-test",
		PlaylistUrl:  "https://example.com",
		ThumbnailUrl: "https://example.com",
		StartAt:      time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC).Unix(),
		EndAt:        time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Tags:         []int64{},
	}, WithStatusCode(http.StatusBadRequest))
	assert.NoError(t, err)
	_, err = client.ReserveLivestream(ctx, &ReserveLivestreamRequest{
		Title:        "livestream-test",
		Description:  "livestream-test",
		PlaylistUrl:  "https://example.com",
		ThumbnailUrl: "https://example.com",
		StartAt:      time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC).Unix(),
		EndAt:        time.Date(2025, 4, 1, 1, 0, 0, 0, time.UTC).Unix(),
		Tags:         []int64{},
	}, WithStatusCode(http.StatusBadRequest))
	assert.NoError(t, err)

	// 同じ時間枠に枠数以上予約
	// ループでクライアントログインして同じ時間に予約
	// 同一ユーザは同一時間で１つしか予約を取れない（枠関係なく)ので、注意
	var (
		startAt = time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC).Unix()
		endAt   = time.Date(2024, 9, 1, 9, 0, 0, 0, time.UTC).Unix()
	)
	for i := 1; i <= config.NumSlots*2; i++ {
		loopClient, err := NewClient(
			agent.WithBaseURL(config.TargetBaseURL),
		)
		assert.NoError(t, err)

		_, err = loopClient.Register(ctx, &RegisterRequest{
			Name:        fmt.Sprintf("user%d", i),
			DisplayName: fmt.Sprintf("user%d", i),
			Description: "livestream-test-loop",
			Password:    "test",
			Theme: Theme{
				DarkMode: user.DarkMode,
			},
		})
		err = loopClient.Login(ctx, &LoginRequest{
			UserName: fmt.Sprintf("user%d", i),
			Password: "test",
		})
		assert.NoError(t, err)

		if i > config.NumSlots {
			_, err = loopClient.ReserveLivestream(ctx, &ReserveLivestreamRequest{
				Title:        "livestream-test1",
				Description:  "livestream-test1",
				PlaylistUrl:  "https://example.com",
				ThumbnailUrl: "https://example.com",
				StartAt:      startAt,
				EndAt:        endAt,
				Tags:         []int64{},
			}, WithStatusCode(http.StatusBadRequest))
		} else {
			_, err = loopClient.ReserveLivestream(ctx, &ReserveLivestreamRequest{
				Title:        "livestream-test1",
				Description:  "livestream-test1",
				PlaylistUrl:  "https://example.com",
				ThumbnailUrl: "https://example.com",
				StartAt:      startAt,
				EndAt:        endAt,
				Tags:         []int64{},
			})
		}
		assert.NoError(t, err)
	}
}
