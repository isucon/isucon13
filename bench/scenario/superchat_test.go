package scenario

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/stretchr/testify/assert"
)

func TestSuperchat(t *testing.T) {
	client, err := isupipe.NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultBenchmarkWorkerTimeoutSeconds*time.Second)
	defer cancel()
	benchscore.InitScore(ctx)
	bencherror.InitPenalty(ctx)

	assert.NotPanics(t, func() { Superchat(ctx, client) })
	fmt.Fprintf(os.Stderr, "superchat: score ==> %d\n", benchscore.GetCurrentScore())
	fmt.Fprintf(os.Stderr, "superchat: profit ==> %d\n", benchscore.GetCurrentProfit())
}
