package resolver

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/isucon/isucon13/bench/internal/config"
)

type NativeDNSResolver struct {
	Nameserver string
	Timeout    time.Duration
}

func NewNativeDNSResolver() *NativeDNSResolver {
	return &NativeDNSResolver{
		Nameserver: net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort)),
		Timeout:    10000 * time.Millisecond,
	}
}

func (r *NativeDNSResolver) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	dialer := net.Dialer{
		Timeout: r.Timeout,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialer := net.Dialer{Timeout: r.Timeout}
				return dialer.DialContext(ctx, "udp", r.Nameserver)
			},
		},
	}
	return dialer.DialContext(ctx, network, address)
}
