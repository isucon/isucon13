package isupipe

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/logger"
)

func TestMain(m *testing.M) {
	testLogger, err := logger.InitTestLogger()
	if err != nil {
		log.Fatalln(err)
	}

	client, err := NewClient(
		testLogger,
		agent.WithTimeout(1*time.Minute),
	)
	if err != nil {
		log.Fatalln(err)
	}
	config.TargetWebapps = []string{"127.0.0.1"}
	if _, err := client.Initialize(context.Background()); err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	benchscore.InitCounter(ctx)
	bencherror.InitErrors(ctx)

	m.Run()
}
