package isupipe

import (
	"context"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/stretchr/testify/assert"
)

// FIXME: 変動をテスト
func TestPayment(t *testing.T) {
	ctx := context.Background()

	client, err := NewClient(
		resolver.NewDNSResolver(),
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithTimeout(3*time.Second),
	)
	assert.NoError(t, err)

	result1, err := client.GetPaymentResult(ctx)
	assert.NoError(t, err)

	// 投げ銭投稿

	// 変動チェック
	_ = result1
}
