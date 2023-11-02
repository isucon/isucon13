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

func TestLivestreamSearch(t *testing.T) {
	ctx := context.Background()

	client, err := NewClient(
		agent.WithBaseURL(config.TargetBaseURL),
	)
	assert.NoError(t, err)

	user := scheduler.UserScheduler.GetRandomStreamer()
	err = client.Login(ctx, &LoginRequest{
		UserName: user.Name,
		Password: user.RawPassword,
	})
	assert.NoError(t, err)

	livestreams, err := client.SearchLivestreams(ctx)
	assert.NoError(t, err)
	assert.NotZero(t, len(livestreams))
}
