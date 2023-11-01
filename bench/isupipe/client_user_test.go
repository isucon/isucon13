package isupipe

import (
	"context"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestClientUser_Login(t *testing.T) {
	ctx := context.Background()

	client, err := NewClient(
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
		UserName: streamer.Name,
		Password: streamer.RawPassword,
	})
	assert.NoError(t, err)

	user, err := client.GetMe(ctx)
	assert.NoError(t, err)
	assert.Equal(t, streamer.Name, user.Name)
}
