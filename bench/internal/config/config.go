package config

const (
// NOTE: あまり並列度高く長い時間ベンチさせると、ポートが枯渇する
// FIXME: 2以上にすると、シナリオの実装上、ランダムに選ばれたlivestream_idが重複したworker fnが同時に走り、
//
//	POST /livestream/:livestream_id/enter が重複した配信に対して行われ、
//	テーブルのUNIQUE制約(user_id-livestream_id)を侵す可能性がある
//	対策としては、アルゴリズム側で必ず重複しないように調整するか、
//	worker parallelismを1にしつつ、視聴者のシミュレータをgoroutineで吐き出して並行性を担保する
//	とりあえず後者で対応
//
// DefaultBenchmarkerParallelism = 5
// シナリオテストのタイムアウト[秒]
// ScenarioTestTimeoutSeconds = 3
)

var Language string = "unknown"
