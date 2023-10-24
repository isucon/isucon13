package isupipe

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/bench/internal/pubsub"
)

type LivestreamPool struct {
	pool *pubsub.PubSub
}

func NewLivestreamPool(ctx context.Context) *LivestreamPool {
	pool := pubsub.NewPubSub(1000)
	pool.Run(ctx)
	return &LivestreamPool{
		pool: pool,
	}
}

func (p *LivestreamPool) Get(ctx context.Context) (*Livestream, error) {
	v, err := p.pool.Subscribe(ctx)
	if err != nil {
		return nil, err
	}

	livestream, ok := v.(*Livestream)
	if !ok {
		return nil, fmt.Errorf("got invalid livestream from pool")
	}

	return livestream, nil
}

func (p *LivestreamPool) Put(ctx context.Context, livestream *Livestream) {
	p.pool.Publish(ctx, livestream)
}
