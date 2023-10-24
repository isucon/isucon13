package pubsub

import (
	"context"
	"sync"
)

// FIXME: mainで初期化してRunしたのち、ベンチマーカに渡す

// PubSub は、Publisherから供給されたアイテムを先着順かつ公平(飢餓しない)にSubscriberに分配します
// NOTE: 内部的にチャネルを用いているため、ブロックする場合があります
type PubSub struct {
	itemCh    chan interface{}
	processCh chan chan interface{}
	closeOnce sync.Once
}

func NewPubSub(itemCapacity int) *PubSub {
	return &PubSub{
		itemCh:    make(chan interface{}, itemCapacity),
		processCh: make(chan chan interface{}),
	}
}

// NOTE: 誰もSubscriberがいない状態で、itemだけたくさん書き出し続けているとブロックする場合があります(通常想定されない)
func (p *PubSub) Publish(ctx context.Context, v interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.itemCh <- v:
		return nil
	}
}

// NOTE: アイテム供給ができていないと、Subscribeがブロックする場合があります (通常想定されない)
func (p *PubSub) Subscribe(ctx context.Context) (interface{}, error) {
	ch := make(chan interface{})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p.processCh <- ch:
		// NOTE: processChに書き込みを行うのがSubscribeのみであり、かつ書き込み後即座にchから読み込むのが重要
		// processCh書き込み後に読み込みを行わずに寄り道すると、公平に分配するPubSub.Run(context.Context)が停止する
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case v := <-ch:
			return v, nil
		}
	}
}

// Run は、公平にアイテムをSubscriberへ分配します。PublisherやSubScriber動作前に実行しておく必要があります
func (p *PubSub) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
			case subscriberCh := <-p.processCh:
				select {
				case <-ctx.Done():
				case subscriberCh <- (<-p.itemCh):
					close(subscriberCh)
				}
			}
		}
	}()
}

func (p *PubSub) Close() {
	p.closeOnce.Do(func() {
		close(p.itemCh)
		close(p.processCh)
	})
}
