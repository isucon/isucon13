package isupipe

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
)

func TestMain(m *testing.M) {
	client, err := NewClient(
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithTimeout(1*time.Minute),
	)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err := client.Initialize(context.Background()); err != nil {
		log.Fatalln(err)
	}

	m.Run()
}
