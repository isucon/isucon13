package scenario

import (
	"context"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/stretchr/testify/assert"
)

func TestPretest(t *testing.T, dnsResolver *resolver.DNSResolver) {
	ctx := context.Background()
	benchscore.InitCounter(ctx)
	benchscore.InitProfit(ctx)
	bencherror.InitErrors(ctx)

	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithTimeout(1*time.Minute),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Initialize(ctx)
	assert.NoError(t, err)

	err = Pretest(ctx, dnsResolver)
	assert.NoError(t, err)
}
