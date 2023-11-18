package resolver

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/miekg/dns"
)

var atomicId = uint64(1)
var msgPool = sync.Pool{
	New: func() any {
		msg := new(dns.Msg)
		msg.Question = make([]dns.Question, 1)
		msg.Question[0] = dns.Question{
			Qtype:  dns.TypeA,
			Qclass: dns.ClassINET,
		}
		return msg
	},
}

type DNSResolver struct {
	Nameserver      string
	Timeout         time.Duration
	ResolveAttempts uint
}

func NewDNSResolver() *DNSResolver {
	return &DNSResolver{
		Nameserver:      net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort)),
		Timeout:         2 * time.Second,
		ResolveAttempts: 1,
	}
}

func (r *DNSResolver) Lookup(ctx context.Context, network, addr string) (net.IP, error) {
	msg := msgPool.Get().(*dns.Msg)
	defer msgPool.Put(msg)
	msg.Id = uint16(atomic.AddUint64(&atomicId, 1))
	msg.Question[0].Name = dns.Fqdn(addr)
	msg.RecursionDesired = false

	client := new(dns.Client)

	var in *dns.Msg
	var err error

	for i := uint(0); i < r.ResolveAttempts; i++ {
		in, _, err = client.ExchangeContext(ctx, msg, r.Nameserver)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		benchscore.IncDNSFailed()
		return nil, err
	}

	// プロトコル上成功をカウントする
	benchscore.IncResolves()

	if in.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("failed to resolve %s with rcode=%d", addr, in.Rcode)
	}

	for _, ans := range in.Answer {
		if record, ok := ans.(*dns.A); ok {
			// TODO: IPアドレスが競技者のものか確認
			return record.A, nil
		}
	}

	return nil, fmt.Errorf("failed to resolve %s: not A record in response", addr)
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
