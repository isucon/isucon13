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
		Title         string `json:"title"`
		Description   string `json:"description"`
		PrivacyStatus string `json:"privacy_status"`
		StartAt       int64  `json:"start_at"`
		EndAt         int64  `json:"end_at"`
	}
)

type (
	PostSuperchatRequest struct {
		Comment string `json:"comment"`
		Tip     int    `json:"tip"`
	}
	PostSuperchatResponse struct {
		ID           int       `json:"superchat_id"`
		UserID       int       `json:"user_id"`
		LivestreamID int       `json:"livestream_id"`
		Comment      string    `json:"comment"`
		Tip          int       `json:"tip"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}
)

type PostReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

type Theme struct {
	DarkMode bool `json:"dark_mode"`
}

type Superchat struct {
	ID           int       `json:"id"`
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
	UserID       string    `json:"user_id"`
	LivestreamID string    `json:"livestream_id"`
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
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	PrivacyStatus string    `json:"privacy_status"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
