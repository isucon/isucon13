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

func TestReaction(t *testing.T) {
	client, err := isupipe.NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	assert.NoError(t, err)
	benchscore.SetAchivementGoal(0)

	ctx, cancel := context.WithTimeout(context.Background(), config.ScenarioTestTimeoutSeconds*time.Second)
	defer cancel()
	benchscore.InitScore(ctx)
	bencherror.InitPenalty(ctx)

	assert.NotPanics(t, func() { Reaction(ctx, client) })
	fmt.Fprintf(os.Stderr, "final score ==> %d\n", benchscore.GetCurrentScore())
	fmt.Fprintf(os.Stderr, "final profit ==> %d\n", benchscore.GetCurrentProfit())
}
