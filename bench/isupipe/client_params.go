package isupipe

type InitializeResponse struct {
	AdvertiseLevel int    `json:"advertise_level"`
	Language       string `json:"language"`
}

type (
	PostUserRequest struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		// Password is non-hashed password.
		Password string `json:"password"`
		Theme    Theme  `json:"theme"`
	}
	LoginRequest struct {
		UserName string `json:"username"`
		// Password is non-hashed password.
		Password string `json:"password"`
	}
)

type (
	ReserveLivestreamRequest struct {
		Tags        []int  `json:"tags"`
		Title       string `json:"title"`
		Description string `json:"description"`
		StartAt     int64  `json:"start_at"`
		EndAt       int64  `json:"end_at"`
	}
)

type (
	PostLivecommentRequest struct {
		Comment string `json:"comment"`
		Tip     int    `json:"tip"`
	}
	PostLivecommentResponse struct {
		Id           int    `json:"id"`
		UserId       int    `json:"user_id"`
		LivestreamId int    `json:"livestream_id"`
		Comment      string `json:"comment"`
		Tip          int    `json:"tip"`
		CreatedAt    int    `json:"created_at"`
		UpdatedAt    int    `json:"updated_at"`
	}
)

type ModerateRequest struct {
	NGWord string `json:"ng_word"`
}

type PostReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

type Theme struct {
	DarkMode bool `json:"dark_mode"`
}

type Livecomment struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id"`
	LivestreamId int    `json:"livestream_id"`
	Comment      string `json:"comment"`
	Tip          int    `json:"tip"`
	ReportCount  int    `json:"report_count"`
	CreatedAt    int    `json:"created_at"`
	UpdatedAt    int    `json:"updated_at"`
}

type Reaction struct {
	Id           int    `json:"id"`
	EmojiName    string `json:"emoji_name"`
	UserId       int    `json:"user_id"`
	LivestreamId int    `json:"livestream_id"`
	CreatedAt    int    `json:"created_at"`
}

type User struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	// HashedPassword is hashed password.
	HashedPassword string `json:"password"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt int `json:"created_at"`
	UpdatedAt int `json:"updated_at"`

	IsFamous bool `json:"is_famous"`
}

type Livestream struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PlaylistUrl  string `json:"playlist_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
	ViewersCount int    `json:"viewers_count"`
	StartAt      int    `json:"start_at"`
	EndAt        int    `json:"end_at"`
	CreatedAt    int    `json:"created_at"`
	UpdatedAt    int    `json:"updated_at"`
}

type LivecommentReport struct {
	Id            int `json:"id"`
	UserId        int `json:"user_id"`
	LivestreamId  int `json:"livestream_id"`
	LivecommentId int `json:"livecomment_id"`
	CreatedAt     int `json:"created_at"`
	UpdatedAt     int `json:"updated_at"`
}

type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt int `json:"created_at"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}

type Payment struct {
	ReservationId int
	Tip           int
}

type PaymentResult struct {
	Total    int
	Payments []*Payment
}

type LivestreamStatistics struct {
	MostTipRanking            []TipRank      `json:"most_tip_ranking"`
	MostPostedReactionRanking []ReactionRank `json:"most_posted_reaction_ranking"`
}

type UserStatistics struct {
	TipRankPerLivestreams map[int]TipRank `json:"tip_rank_by_livestream"`
}

type TipRank struct {
	Rank     int `json:"tip_rank"`
	TotalTip int `json:"total_tip"`
}

type ReactionRank struct {
	Rank      int    `json:"reaction_rank"`
	EmojiName string `json:"emoji_name"`
}
