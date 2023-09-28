package config

const (
	// NOTE: あまり並列度高く長い時間ベンチさせると、ポートが枯渇する
	DefaultBenchmarkerParallelism        = 2
	DefaultBenchmarkWorkerTimeoutSeconds = 10
)

var (
	// 広告費用
	// 1~10で設定
	AdvertiseCost = 1
)
