package attacker

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/miekg/dns"
	"github.com/valyala/bytebufferpool"
)

// refers https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go

const (
	letters       = "abcdefghijklmnopqrstuvwxyz0123456789"
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
	zone          = "u.isucon.dev."
	maxLength     = 22
)

var src = rand.NewSource(time.Now().UnixNano())
var mutex sync.Mutex

type DnsWaterTortureAttacker struct {
	connected               bool
	dnsClient               *dns.Client
	dnsConn                 *dns.Conn
	maxRequestPerConnection int
	numRequestPerConnection int
	request                 uint64
	resolvedRequests        uint64
}

func int63() int64 {
	mutex.Lock()
	n := src.Int63()
	mutex.Unlock()
	return n
}

func randString(bb io.ByteWriter, n int) {
	for i, cache, remain := n-1, int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letters) {
			bb.WriteByte(letters[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
}

func NewDnsWaterTortureAttacker() *DnsWaterTortureAttacker {
	c := &dns.Client{Net: "udp", Timeout: 1 * time.Second}
	return &DnsWaterTortureAttacker{
		maxRequestPerConnection: 10,
		numRequestPerConnection: 0,
		dnsClient:               c,
	}
}

func (a *DnsWaterTortureAttacker) Attack(ctx context.Context, httpClient *http.Client) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	numOfLabel := 1
	if atomic.AddUint64(&a.request, 1)%30 == 0 {
		numOfLabel += rand.Intn(3)
	}
	for i := 0; i < numOfLabel; i++ {
		length := 10 + rand.Intn(maxLength)
		randString(buf, length)
		buf.WriteByte('0')
		buf.WriteByte('.')
	}
	buf.WriteString(zone)
	b := buf.Bytes()
	name := unsafe.String(&b[0], len(b))
	ip := a.lookup(ctx, name)
	if ip != nil && atomic.AddUint64(&a.resolvedRequests, 1)%5 == 0 {
		host := fmt.Sprintf("%s:%d", strings.TrimRight(name, "."), config.TargetPort)
		// root, favicon.ico, api/user/me, api/tag, api/livestream/search?limit=50 をランダムに
		endpoints := []string{"", "", "", "favicon.ico", "api/user/me", "api/tag", "api/livestream/search?limit=50"}
		url := fmt.Sprintf("%s://%s/%s",
			config.HTTPScheme,
			host,
			endpoints[rand.Intn(len(endpoints))],
		)
		valueCtx := context.WithValue(ctx, config.AttackHTTPClientContextKey,
			fmt.Sprintf("%s:%d", ip.String(), config.TargetPort))
		req, err := http.NewRequestWithContext(valueCtx, "GET", url, nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", "isucandar")
		res, err := httpClient.Do(req)
		if err != nil {
			return
		}
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}
}

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

func (a *DnsWaterTortureAttacker) lookup(ctx context.Context, name string) net.IP {
	if !a.connected {
		nameserver := net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort))
		dnsConn, err := a.dnsClient.Dial(nameserver)
		if err != nil {
			return nil
		}
		a.connected = true
		a.dnsConn = dnsConn
	}

	msg := msgPool.Get().(*dns.Msg)
	defer msgPool.Put(msg)
	msg.Id = uint16(atomic.AddUint64(&atomicId, 1))
	msg.Question[0].Name = name
	msg.RecursionDesired = false

	a.numRequestPerConnection++
	in, _, err := a.dnsClient.ExchangeWithConn(msg, a.dnsConn)
	if err != nil {
		a.dnsConn.Close()
		a.connected = false
		benchscore.IncDNSFailed()
		return nil
	}
	if a.numRequestPerConnection >= a.maxRequestPerConnection {
		a.dnsConn.Close()
		a.connected = false
		a.numRequestPerConnection = 0
	}
	// プロトコル上成功をカウントする
	benchscore.IncResolves()

	for _, ans := range in.Answer {
		if record, ok := ans.(*dns.A); ok {
			if config.IsWebappIP(record.A) {
				return record.A
			}
		}
	}
	return nil

}
