package scenario

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/benchtest"
	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/stretchr/testify/assert"
)

func TestNormal(t *testing.T) {
	testResource, err := benchtest.Setup()
	if err != nil {
		log.Fatalln(err)
	}
	assert.NoError(t, err)
	defer benchtest.Teardown(testResource)

	client, err := isupipe.NewClient(
		agent.WithBaseURL(testResource.WebappIPAddress()),
	)
	assert.NoError(t, err)

	ctx := context.Background()
	benchscore.InitScore(ctx)

	assert.NotPanics(t, func() { Normal(ctx, client) })
	fmt.Fprintf(os.Stderr, "final score ==> %d\n", benchscore.GetFinalScore())
}
