package scenario

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSeason1(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultBenchmarkWorkerTimeoutSeconds*time.Second)
	defer cancel()
	benchscore.InitScore(ctx)

	config.AdvertiseCost = 10
	assert.NotPanics(t, func() { Season1(ctx, webappIPAddress) })
	fmt.Fprintf(os.Stderr, "season1: score ==> %d\n", benchscore.GetCurrentScore())
	fmt.Fprintf(os.Stderr, "season1: profit ==> %d\n", benchscore.GetCurrentProfit())
	fmt.Fprintf(os.Stderr, "season1: penalty ==> %d\n", benchscore.GetCurrentPenalty())

}
