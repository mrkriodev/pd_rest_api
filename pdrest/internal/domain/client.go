package domain

type UserLastLogin struct {
	UUID        string `json:"uuid"`
	LastLoginAt *int64 `json:"last_login_at,omitempty"`
}

type UserProfile struct {
	UUID     string  `json:"uuid"`
	Username *string `json:"username,omitempty"`
}

type User struct {
	UUID              string  `json:"uuid"`
	GoogleID          *string `json:"google_id,omitempty"`
	GoogleEmail       *string `json:"google_email,omitempty"`
	GoogleName        *string `json:"google_name,omitempty"`
	TelegramID        *int64  `json:"telegram_id,omitempty"`
	TelegramUsername  *string `json:"telegram_username,omitempty"`
	TelegramFirstName *string `json:"telegram_first_name,omitempty"`
	TelegramLastName  *string `json:"telegram_last_name,omitempty"`
}
