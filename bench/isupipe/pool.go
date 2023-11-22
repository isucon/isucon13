package isupipe

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/bench/internal/pubsub"
	"golang.org/x/time/rate"
)

// ClientPool は、ログイン後のクライアントプールです
type ClientPool struct {
	pool    *pubsub.PubSub
	limiter *rate.Limiter
}

type ClientPoolOptions struct {
	WithoutLimitter bool
}

type ClientPoolOption func(o *ClientPoolOptions)

func WithoutPoolLimiter() ClientPoolOption {
	return func(o *ClientPoolOptions) {
		o.WithoutLimitter = true
	}
}

func NewClientPool(ctx context.Context, opts ...ClientPoolOption) *ClientPool {
	o := &ClientPoolOptions{}
	for _, setOpt := range opts {
		setOpt(o)
	}
	pool := pubsub.NewPubSub(2000)
	pool.Run(ctx)
	cp := &ClientPool{
		pool: pool,
	}
	if !o.WithoutLimitter {
		cp.limiter = rate.NewLimiter(1, 1) // 1秒間に1回
	}
	return cp
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
