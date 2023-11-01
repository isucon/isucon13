package isupipe

import (
	"context"
	"testing"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestPayment(t *testing.T) {
	ctx := context.Background()

	client, err := NewClient(
		agent.WithBaseURL(config.TargetBaseURL),
	)
	assert.NoError(t, err)

	result1, err := client.GetPaymentResult(ctx)
	assert.NoError(t, err)

	// 投げ銭投稿

	// 変動チェック
}
