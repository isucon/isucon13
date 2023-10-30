package config

import (
	"fmt"
	"time"
)

const (
	TargetPort = 8080
)

// FIXME: httpsになりそうなので、スキームも置き換えられるようにする
var (
	TargetBaseURL    string = fmt.Sprintf("http://pipe.u.isucon.dev:%d", TargetPort)
	TargetNameserver string = "127.0.0.1"
	DNSPort          int    = 1053
)

var RequestTooSlowThreshold = 500 * time.Millisecond
