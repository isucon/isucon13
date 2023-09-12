package isupipe

type (
	PostUserRequest struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		// Password is non-hashed password.
		Password string `json:"password"`
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
)
