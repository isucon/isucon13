package isupipe

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/bench/internal/pubsub"
)

// ClientPool は、ログイン後のクライアントプールです
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

// LivestreamPool は、予約後のライブ配信プールです
type LivestreamPool struct {
	pool *pubsub.PubSub
}

func NewLivestreamPool(ctx context.Context) *LivestreamPool {
	pool := pubsub.NewPubSub(10000)
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

type LivecommentPool struct {
	pool *pubsub.PubSub
}

func NewLivecommentPool(ctx context.Context) *LivecommentPool {
	pool := pubsub.NewPubSub(100000)
	pool.Run(ctx)
	return &LivecommentPool{
		pool: pool,
	}
}

func (p *LivecommentPool) Get(ctx context.Context) (*Livecomment, error) {
	v, err := p.pool.Subscribe(ctx)
	if err != nil {
		return nil, err
	}

	livecomment, ok := v.(*Livecomment)
	if !ok {
		return nil, fmt.Errorf("got invalid livestream from pool")
	}

	return livecomment, nil
}

func (p *LivecommentPool) Put(ctx context.Context, livecomment *Livecomment) {
	p.pool.Publish(ctx, livecomment)
}
