package config

import (
	"fmt"
)

var (
	TargetBaseURL    string = fmt.Sprintf("%s://pipe.%s:%d", HTTPScheme, BaseDomain, TargetPort)
	TargetNameserver string = "127.0.0.1"
	DNSPort          int    = 1053
	TargetPort       int    = 8080
)
