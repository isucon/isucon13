package config

import "time"

const DefaultBenchmarkTimeout = 60 * time.Second
const TooManySpamThresholdPercentage = 50.0

const InsecureSkipVerify = true
const DefaultAgentTimeout = 300 * time.Millisecond

const HTTPScheme = "http"
const BaseDomain = "u.isucon.dev"
