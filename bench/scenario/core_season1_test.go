package scenario

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/stretchr/testify/assert"
)

func TestSeason1(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	benchscore.InitScore(ctx)

	assert.NotPanics(t, func() { Season1(ctx, webappIPAddress) })
	fmt.Fprintf(os.Stderr, "season1: score ==> %d\n", benchscore.GetCurrentScore())
	fmt.Fprintf(os.Stderr, "season1: profit ==> %d\n", benchscore.GetCurrentProfit())
	fmt.Fprintf(os.Stderr, "season1: penalty ==> %d\n", benchscore.GetCurrentPenalty())

}
