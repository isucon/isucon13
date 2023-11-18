package scenario

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
)

func dnsrecordPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	for i := 0; i < 10; i++ {
		r := config.DefaultDNSRecord[rand.Intn(len(config.DefaultDNSRecord))]
		_, err := dnsResolver.Lookup(ctx, "udp", fmt.Sprintf("%s.%s", r, config.BaseDomain))
		if err != nil {
			return err
		}
	}
	return nil
}
