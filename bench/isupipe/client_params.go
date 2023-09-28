package isupipe

import "time"

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
	PostSuperchatRequest struct {
		Comment string `json:"comment"`
		Tip     int    `json:"tip"`
	}
	PostSuperchatResponse struct {
		Id           int       `json:"id"`
		UserID       int       `json:"user_id"`
		LivestreamID int       `json:"livestream_id"`
		Comment      string    `json:"comment"`
		Tip          int       `json:"tip"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
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

type Superchat struct {
	Id           int       `json:"id"`
	UserID       int       `json:"user_id"`
	LivestreamID int       `json:"livestream_id"`
	Comment      string    `json:"comment"`
	Tip          int       `json:"tip"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Reaction struct {
	ID           int       `json:"id"`
	EmojiName    string    `json:"emoji_name"`
	UserID       int       `json:"user_id"`
	LivestreamID int       `json:"livestream_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	// HashedPassword is hashed password.
	HashedPassword string `json:"password"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Livestream struct {
	Id          int       `json:"id"`
	UserId      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt time.Time `json:"created_at"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}
