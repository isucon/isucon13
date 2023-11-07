package attacker

import (
	"context"
	"math/rand"

	"github.com/isucon/isucon13/bench/internal/resolver"
)

const (
	asciiLowercase = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialChars   = "-."
)

type DnsWaterTortureAttacker struct {
	resolver *resolver.DNSResolver
}

func NewDnsWaterTortureAttacker() *DnsWaterTortureAttacker {
	dnsResolver := resolver.NewDNSResolver()
	return &DnsWaterTortureAttacker{
		resolver: dnsResolver,
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

func (a *DnsWaterTortureAttacker) Attack(ctx context.Context) {
	length := 10 + rand.Intn(40)
	payload := a.makePayload(length)
	a.resolver.Lookup(ctx, "udp", payload)
}
