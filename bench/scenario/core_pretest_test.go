package scenario

import (
	"context"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/stretchr/testify/assert"
)

func TestPretest(t *testing.T) {
	ctx := context.Background()
	benchscore.InitScore(ctx)
	// bencherror.InitPenalty(ctx)

	client, err := isupipe.NewClient(
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithTimeout(1*time.Minute),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Initialize(ctx)
	assert.NoError(t, err)

	err = Pretest(ctx, client)
	assert.NoError(t, err)
}
