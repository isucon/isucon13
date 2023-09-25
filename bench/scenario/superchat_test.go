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

func TestSuperchat(t *testing.T) {
	client, err := isupipe.NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	assert.NoError(t, err)

	ctx := context.Background()
	benchscore.InitScore(ctx)

	assert.NotPanics(t, func() { Superchat(ctx, client) })
	fmt.Fprintf(os.Stderr, "superchat: score ==> %d\n", benchscore.GetCurrentScore())
	fmt.Fprintf(os.Stderr, "superchat: profit ==> %d\n", benchscore.GetCurrentProfit())
}
