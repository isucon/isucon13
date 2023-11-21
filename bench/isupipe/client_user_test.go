package isupipe

import (
	"bytes"
	"context"
	"crypto/sha256"
	"image"
	_ "image/jpeg"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/logger"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestClientUser_Login(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitTestLogger()

	client, err := NewClient(
		testLogger,
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithTimeout(1*time.Minute),
	)
	assert.NoError(t, err)

	streamer := scheduler.UserScheduler.GetRandomStreamer()
	assert.NoError(t, err)

	_, err = client.Register(ctx, &RegisterRequest{
		Name:        streamer.Name,
		DisplayName: streamer.DisplayName,
		Description: streamer.Description,
		Password:    streamer.RawPassword,
		Theme: Theme{
			DarkMode: streamer.DarkMode,
		},
	})
	assert.NoError(t, err)

	err = client.Login(ctx, &LoginRequest{
		Username: streamer.Name,
		Password: streamer.RawPassword,
	})
	assert.NoError(t, err)

	// 自身の情報確認
	user, err := client.GetMe(ctx)
	assert.NoError(t, err)
	assert.Equal(t, streamer.Name, user.Name)

	// テーマ取得
	theme, err := client.GetStreamerTheme(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, streamer.DarkMode, theme.DarkMode)

	// アイコンアップロード・取得
	img := scheduler.IconSched.GetRandomIcon()
	postIconResp, err := client.PostIcon(ctx, &PostIconRequest{
		Image: img.Image,
	})
	assert.NoError(t, err)
	assert.NotZero(t, postIconResp.ID)
	beforeHash := sha256.Sum256(img.Image[:])

	imageBytes, err := client.GetIcon(ctx, user.Name)
	assert.NoError(t, err)
	afterHash := sha256.Sum256(imageBytes[:])
	assert.True(t, bytes.Equal(beforeHash[:], afterHash[:]))
	_, imageFormat, err := image.Decode(bytes.NewBuffer(imageBytes))
	assert.NoError(t, err)
	assert.Equal(t, "jpeg", imageFormat)
}
