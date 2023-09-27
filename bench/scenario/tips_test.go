package scenario

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/stretchr/testify/assert"
)

func TestTips(t *testing.T) {
	client, err := isupipe.NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultBenchmarkWorkerTimeoutSeconds*time.Second)
	defer cancel()
	benchscore.InitScore(ctx)

	assert.NotPanics(t, func() { Tips(ctx, client) })
	fmt.Fprintf(os.Stderr, "tips: score ==> %d\n", benchscore.GetCurrentScore())
	fmt.Fprintf(os.Stderr, "tips: profit ==> %d\n", benchscore.GetCurrentProfit())
}
