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
)

func TestMain(m *testing.M) {
	client, err := NewClient(
		agent.WithTimeout(1 * time.Minute),
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
