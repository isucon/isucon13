package config

import (
	"fmt"
	"net"
)

var (
	TargetBaseURL    string   = fmt.Sprintf("%s://pipe.%s:%d", HTTPScheme, BaseDomain, TargetPort)
	TargetNameserver string   = "127.0.0.1"
	TargetWebapps    []string = []string{}
	DNSPort          int      = 1053
	TargetPort       int      = 8080
)

func IsWebappIP(ip net.IP) bool {
	for _, s := range TargetWebapps {
		if ip.String() == s {
			return true
		}
	}
	return false
}
