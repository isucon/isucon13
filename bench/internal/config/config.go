package config

const (
	// NOTE: あまり並列度高く長い時間ベンチさせると、ポートが枯渇する
	DefaultBenchmarkerParallelism        = 2
	DefaultBenchmarkWorkerTimeoutSeconds = 10
)
