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

func TestLivecomment(t *testing.T) {
	client, err := isupipe.NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), config.ScenarioTestTimeoutSeconds*time.Second)
	defer cancel()
	benchscore.InitScore(ctx)
	bencherror.InitPenalty(ctx)
	benchscore.SetAchivementGoal(0)

	assert.NotPanics(t, func() { Livecomment(ctx, client) })
	fmt.Fprintf(os.Stderr, "livecomment: score ==> %d\n", benchscore.GetCurrentScore())
	fmt.Fprintf(os.Stderr, "livecomment: profit ==> %d\n", benchscore.GetCurrentProfit())
}
