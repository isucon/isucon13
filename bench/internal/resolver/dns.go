package resolver

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/miekg/dns"
)

type DNSResolver struct {
	Nameserver string
	Timeout    time.Duration
}

func NewDNSResolver() *DNSResolver {
	return &DNSResolver{
		Nameserver: net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort)),
		Timeout:    10000 * time.Millisecond,
	}
}

func (r *DNSResolver) Lookup(ctx context.Context, network, addr string) (net.IP, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(addr), dns.TypeA)
	msg.RecursionDesired = false

	client := new(dns.Client)
	in, _, err := client.ExchangeContext(ctx, msg, r.Nameserver)
	if err != nil {
		return nil, err
	}

	if in.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("failed to resolve with rcode=%d", in.Rcode)
	}

	for _, ans := range in.Answer {
		if record, ok := ans.(*dns.A); ok {
			benchscore.IncResolves()
			return record.A, nil
		}
	}

	return nil, fmt.Errorf("failed to resolve %s", addr)
}

func (r *DNSResolver) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	ip, err := r.Lookup(ctx, network, host)
	if err != nil {
		return nil, err
	}

	d := new(net.Dialer)
	d.Timeout = r.Timeout
	return d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
}
