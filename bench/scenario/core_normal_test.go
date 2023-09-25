package scenario

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/stretchr/testify/assert"
)

func TestNormal(t *testing.T) {
	client, err := isupipe.NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	assert.NoError(t, err)

	ctx := context.Background()
	benchscore.InitScore(ctx)

	assert.NotPanics(t, func() { Normal(ctx, client) })
	fmt.Fprintf(os.Stderr, "final score ==> %d\n", benchscore.GetFinalScore())
	fmt.Fprintf(os.Stderr, "final profit ==> %d\n", benchscore.GetProfit())
}
