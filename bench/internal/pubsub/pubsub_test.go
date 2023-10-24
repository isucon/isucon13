package pubsub

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Item struct{}

func TestPubSub(t *testing.T) {
	pool := NewPubSub(10)
	go pool.Run(context.TODO())

	pool.Publish(context.TODO(), &Item{})
	pool.Publish(context.TODO(), &Item{})
	pool.Publish(context.TODO(), &Item{})
	pool.Publish(context.TODO(), &Item{})
	pool.Publish(context.TODO(), &Item{})

	v, err := pool.Subscribe(context.TODO())
	assert.NoError(t, err)
	fmt.Println(v)
}
