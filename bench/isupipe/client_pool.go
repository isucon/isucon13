package isupipe

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/bench/internal/pubsub"
)

// var (
// 	PopularStreamerClientPool = newClientPool()
// 	StreamerClientPool        = newClientPool()
// 	ViewerClientPool          = newClientPool()
// )

// var (
// 	LivestreamPool = newLivestreamPool()
// )

type ClientPool struct {
	pool *pubsub.PubSub
}

func NewClientPool(ctx context.Context) *ClientPool {
	pool := pubsub.NewPubSub(2000)
	pool.Run(ctx)
	return &ClientPool{
		pool: pool,
	}
}

func (p *ClientPool) Get(ctx context.Context) (*Client, error) {
	v, err := p.pool.Subscribe(ctx)
	if err != nil {
		return nil, err
	}

	client, ok := v.(*Client)
	if !ok {
		return nil, fmt.Errorf("got invalid client from pool")
	}

	return client, nil
}

func (p *ClientPool) Put(ctx context.Context, c *Client) {
	p.pool.Publish(ctx, c)
}
