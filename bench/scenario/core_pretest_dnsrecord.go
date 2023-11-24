package scenario

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/najeira/randstr"
)

func dnsRecordPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	_, err := dnsResolver.Lookup(ctx, "udp", fmt.Sprintf("%s.%s", "pipe", config.BaseDomain))
	if err != nil {
		return fmt.Errorf("名前解決エラー: %v", err)
	}
	for i := 0; i < 10; i++ {
		r := config.DefaultDNSRecord[rand.Intn(len(config.DefaultDNSRecord))]
		_, err := dnsResolver.Lookup(ctx, "udp", fmt.Sprintf("%s.%s", r, config.BaseDomain))
		if err != nil {
			return fmt.Errorf("名前解決エラー: %v", err)
		}
	}
	// 存在しない名前で
	for i := 0; i < 3; i++ {
		r := strings.ToLower(randstr.String(16))
		_, err = dnsResolver.Lookup(ctx, "udp", fmt.Sprintf("%s.%s", r, config.BaseDomain))
		if err != nil && strings.Contains(err.Error(), "サーバーリストに含まれていません") {
			// is not in the server listの時だけerr。それ以外は無視できる
			return fmt.Errorf("名前解決エラー: %v", err)
		}
	}
	return nil
}
