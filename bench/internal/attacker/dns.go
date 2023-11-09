package attacker

import (
	"bytes"
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/isucon/isucon13/bench/internal/resolver"
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
	resolver *resolver.DNSResolver
}

func int63() int64 {
	mutex.Lock()
	n := src.Int63()
	mutex.Unlock()
	return n
}

func randString(bb *bytes.Buffer, n int) {
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
	dnsResolver := resolver.NewDNSResolver()
	return &DnsWaterTortureAttacker{
		resolver: dnsResolver,
	}
}

func (a *DnsWaterTortureAttacker) Attack(ctx context.Context) {
	buf := new(bytes.Buffer)
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
	a.resolver.Lookup(ctx, "udp", unsafe.String(&b[0], len(b)))
}
