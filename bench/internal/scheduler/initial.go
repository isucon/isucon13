package scheduler

type InitialNgWord struct {
	UserID       int64
	LivestreamID int64
	Word         string
}

type InitialLivecomment struct {
	UserID       int64
	LivestreamID int64
	Comment      string
}

type InitialReaction struct {
	UserID       int64
	LivestreamID int64
	EmojiName    string
}
