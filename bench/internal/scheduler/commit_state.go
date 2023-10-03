package scheduler

// 割当のコミット状態を管理
type CommitState int

const (
	CommitState_None CommitState = iota
	CommitState_Inflight
	CommitState_Committed
)
