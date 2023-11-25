package config

import "time"

// ベンチマーク走行時間タイムアウト
const DefaultBenchmarkTimeout = 60 * time.Second

// スパム離脱割合
const TooManySpamThresholdPercentage = 30.0

// 基本となる並列性
// セマフォの重みに使われます
const BaseParallelism = 1

// 動的に並列度を上げる並列性
// スコアに直結する視聴者シナリオなどのセマフォの重みに使われます
const ChangableParallelism = 100

// この数値だけは、最初のRegister, Login処理が保証されます
// NOTE: このような保証がないと、登録が一切できず、ベンチ走行までシナリオが全く実行されないケースが出てしまいます
const NumMustTryLogins = 10

// HTTPクライアント(isucandar/agent) のタイムアウト
const DefaultAgentTimeout = 20 * time.Second

// POST /api/initialize 時のタイムアウト
const InitializeAgentTimeout = 42 * time.Second

// SearchLivestreamsのLIMITのデフォルト
const NumSearchLivestreams = 50

// NOTE: --enable-ssl オプションによって変更されます
var (
	HTTPScheme         = "http"
	InsecureSkipVerify = true
)

const BaseDomain = "u.isucon.dev"

// 暇になってる接続のタイムアウト
// NOTE: これを設定しないと、keepaliveで繋ぎっぱなしの接続が増え、Nginxでworker_connectionが不十分だというエラーログが出るようになる
const ClientIdleConnTimeout = 5 * time.Second

const AttackHTTPClientContextKey = "dns-attack-http-realip"
