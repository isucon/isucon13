package config

import "time"

const DefaultBenchmarkTimeout = 60 * time.Second
const TooManySpamThresholdPercentage = 20.0

const DefaultAgentTimeout = 5 * time.Second

// NOTE: --enable-ssl オプションによって変更されます
var (
	HTTPScheme         = "http"
	InsecureSkipVerify = true
)

const BaseDomain = "u.isucon.dev"

const ClientIdleConnTimeout = 5 * time.Second

const AttackHTTPClientContextKey string = "dns-attack-http-realip"
