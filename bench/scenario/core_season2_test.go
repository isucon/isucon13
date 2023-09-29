package scenario

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSeason2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultBenchmarkWorkerTimeoutSeconds*time.Second)
	defer cancel()
	benchscore.InitScore(ctx)
	bencherror.InitPenalty(ctx)
	benchscore.SetAchivementGoal(0)

	config.AdvertiseCost = 10
	assert.NotPanics(t, func() { Season2(ctx, webappIPAddress) })
	fmt.Fprintf(os.Stderr, "season2: score ==> %d\n", benchscore.GetCurrentScore())
	fmt.Fprintf(os.Stderr, "season2: profit ==> %d\n", benchscore.GetCurrentProfit())
	fmt.Fprintf(os.Stderr, "season2: penalty ==> %d\n", bencherror.GetCurrentPenalty())

}
