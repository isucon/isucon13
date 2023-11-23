package isupipe

import (
	"context"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/logger"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestGetUserStats(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitTestLogger()
	assert.NoError(t, err)

	client, err := NewClient(testLogger, agent.WithTimeout(30*time.Second))
	assert.NoError(t, err)

	user, err := client.Register(ctx, &RegisterRequest{
		Name:        "get-user-stats",
		DisplayName: "get-user-stats",
		Description: "blah",
		Password:    "test",
		Theme: Theme{
			DarkMode: true,
		},
	})
	assert.NoError(t, err)

	err = client.Login(ctx, &LoginRequest{
		Username: user.Name,
		Password: "test",
	})
	assert.NoError(t, err)

	// (２つくらい配信作る)
	streamer1Client, err := NewClient(nil)
	assert.NoError(t, err)
	streamer1, err := streamer1Client.Register(ctx, &RegisterRequest{
		Name:        "get-user-stats-streamer1",
		DisplayName: "get-user-stats-streamer1",
		Description: "blah",
		Password:    "test",
		Theme: Theme{
			DarkMode: true,
		},
	})
	assert.NoError(t, err)
	err = streamer1Client.Login(ctx, &LoginRequest{
		Username: streamer1.Name,
		Password: "test",
	})
	assert.NoError(t, err)

	stats, err := client.GetUserStatistics(ctx, streamer1.Name)
	assert.NoError(t, err)

	livestream1, err := streamer1Client.ReserveLivestream(ctx, streamer1.Name, &ReserveLivestreamRequest{
		Title:        "get-user-stats-stream1",
		Description:  "get-user-stats-stream1",
		PlaylistUrl:  "https://example.com",
		ThumbnailUrl: "https://example.com",
		StartAt:      time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC).Unix(),
		EndAt:        time.Date(2024, 11, 10, 4, 0, 0, 0, time.UTC).Unix(),
		Tags:         []int64{},
	})
	assert.NoError(t, err)
	livestream2, err := streamer1Client.ReserveLivestream(ctx, streamer1.Name, &ReserveLivestreamRequest{
		Title:        "get-user-stats-stream2",
		Description:  "get-user-stats-stream2",
		PlaylistUrl:  "https://example.com",
		ThumbnailUrl: "https://example.com",
		StartAt:      time.Date(2024, 11, 24, 0, 0, 0, 0, time.UTC).Unix(),
		EndAt:        time.Date(2024, 11, 24, 9, 0, 0, 0, time.UTC).Unix(),
		Tags:         []int64{},
	})
	assert.NoError(t, err)

	// 視聴者を増やしてみる
	err = client.EnterLivestream(ctx, livestream1.ID, streamer1.Name)
	assert.NoError(t, err)
	err = client.EnterLivestream(ctx, livestream2.ID, streamer1.Name)
	assert.NoError(t, err)

	stats2, err := client.GetUserStatistics(ctx, streamer1.Name)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), stats2.ViewersCount-stats.ViewersCount)

	// 配信にリアクションを投稿してみる -> リアクション数, お気に入り絵文字変動
	_, err = client.PostReaction(ctx, livestream1.ID, streamer1.Name, &PostReactionRequest{
		EmojiName: "helicopter",
	})
	assert.NoError(t, err)

	stats3, err := client.GetUserStatistics(ctx, streamer1.Name)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), stats3.TotalReactions-stats.TotalReactions)
	assert.Equal(t, "helicopter", stats3.FavoriteEmoji)

	// 配信にライブコメントを投稿してみる
	_, tipAmount, err := client.PostLivecomment(ctx, livestream1.ID, streamer1.Name, "isu~", &scheduler.Tip{
		Level: 1,
		Tip:   10,
	})
	assert.NoError(t, err)

	stats4, err := client.GetUserStatistics(ctx, streamer1.Name)
	assert.NoError(t, err)
	assert.Equal(t, int64(tipAmount), stats4.TotalTip-stats.TotalTip)
	assert.Equal(t, int64(1), stats4.TotalLivecomments-stats.TotalLivecomments)
}

func TestGetLivestreamStats(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitTestLogger()
	assert.NoError(t, err)

	client, err := NewClient(testLogger, agent.WithTimeout(20*time.Second))
	assert.NoError(t, err)

	user, err := client.Register(ctx, &RegisterRequest{
		Name:        "get-livestream-stats",
		DisplayName: "get-livestraem-stats",
		Description: "blah",
		Password:    "test",
		Theme: Theme{
			DarkMode: true,
		},
	})
	assert.NoError(t, err)

	err = client.Login(ctx, &LoginRequest{
		Username: user.Name,
		Password: "test",
	})
	assert.NoError(t, err)

	stats, err := client.GetLivestreamStatistics(ctx, 1, "test001")
	assert.NoError(t, err)

	// 視聴者を増やす
	err = client.EnterLivestream(ctx, 1, "test001")
	assert.NoError(t, err)
	stats2, err := client.GetLivestreamStatistics(ctx, 1, "test001")
	assert.Equal(t, int64(1), stats2.ViewersCount-stats.ViewersCount)

	// リアクション投稿
	_, err = client.PostReaction(ctx, 1, "test001", &PostReactionRequest{
		EmojiName: "isu",
	})
	assert.NoError(t, err)
	stats3, err := client.GetLivestreamStatistics(ctx, 1, "test001")
	assert.Equal(t, int64(1), stats3.TotalReactions-stats2.TotalReactions)

	// コメント (チップ)
	commenterClient, err := NewClient(nil)
	assert.NoError(t, err)
	commenter, _ := commenterClient.Register(ctx, &RegisterRequest{
		Name:        "get-livestream-stats-commenter",
		DisplayName: "get-livestraem-stats-commenter",
		Description: "blah",
		Password:    "test",
		Theme: Theme{
			DarkMode: true,
		},
	})
	err = commenterClient.Login(ctx, &LoginRequest{
		Username: commenter.Name,
		Password: "test",
	})
	assert.NoError(t, err)
	livecomment, tipAmount, err := commenterClient.PostLivecomment(ctx, 1, "test001", "isuisu", &scheduler.Tip{
		Level: 5,
		Tip:   10000,
	})
	assert.NoError(t, err)
	stats4, err := client.GetLivestreamStatistics(ctx, 1, "test001")
	assert.NoError(t, err)
	assert.Equal(t, int64(tipAmount), stats4.MaxTip)

	// スパム報告
	err = client.ReportLivecomment(ctx, 1, "test001", livecomment.ID)
	assert.NoError(t, err)
	stats5, err := client.GetLivestreamStatistics(ctx, 1, "test001")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), stats5.TotalReports-stats.TotalReports)
}

func TestStatsRank(t *testing.T) {

}
