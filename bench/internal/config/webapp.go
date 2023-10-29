package config

import "fmt"

const (
	TargetPort = 8080
)

var (
	TargetBaseURL    string = fmt.Sprintf("http://pipe.u.isucon.dev:%d", TargetPort)
	TargetNameserver string = "127.0.0.1"
	DNSPort          int    = 1053
)
