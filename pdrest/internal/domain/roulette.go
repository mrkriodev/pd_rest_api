package domain

// RouletteType represents the type of roulette
type RouletteType string

const (
	RouletteTypeOnStart     RouletteType = "on_start"
	RouletteTypeDuringEvent RouletteType = "during_event"
)

// RouletteConfig represents the configuration for a roulette
type RouletteConfig struct {
	ID        int          `json:"id"`
	Type      RouletteType `json:"type"`
	EventID   *string      `json:"event_id,omitempty"` // NULL for on_start, required for during_event
	MaxSpins  int          `json:"max_spins"`
	IsActive  bool         `json:"is_active"`
	CreatedAt int64        `json:"created_at"`
	UpdatedAt int64        `json:"updated_at"`
}

// RoulettePreauthToken represents a preauth token for roulette
type RoulettePreauthToken struct {
	ID               int     `json:"id"`
	Token            string  `json:"token"`
	UserUUID         *string `json:"user_uuid,omitempty"` // Optional, NULL for unauthenticated users
	RouletteConfigID int     `json:"roulette_config_id"`
	IsUsed           bool    `json:"is_used"`
	ExpiresAt        int64   `json:"expires_at"`
	CreatedAt        int64   `json:"created_at"`
}

// Roulette represents a roulette session (linked to preauth token, not user directly)
type Roulette struct {
	ID               int                    `json:"id"`
	RouletteConfigID int                    `json:"roulette_config_id"`
	PreauthTokenID   int                    `json:"preauth_token_id"`
	SpinNumber       int                    `json:"spin_number"`
	Prize            *string                `json:"prize,omitempty"`
	PrizeTaken       bool                   `json:"prize_taken"`
	SpinResult       map[string]interface{} `json:"spin_result,omitempty"`
	CreatedAt        int64                  `json:"created_at"`
	UpdatedAt        int64                  `json:"updated_at"`
	PrizeTakenAt     *int64                 `json:"prize_taken_at,omitempty"`
}

// SpinRequest represents a request to make a spin
type SpinRequest struct {
	PreauthToken string `json:"preauth_token"`
}

// SpinResponse represents the response after a spin
type SpinResponse struct {
	Roulette       *Roulette `json:"roulette"`
	RemainingSpins int       `json:"remaining_spins"`
	CanSpin        bool      `json:"can_spin"`
}

// TakePrizeRequest represents a request to take the prize
type TakePrizeRequest struct {
	PreauthToken string `json:"preauth_token,omitempty"` // Optional, will be generated from session+IP if not provided
}

// TakePrizeResponse represents the response after taking prize
type TakePrizeResponse struct {
	Success      bool   `json:"success"`
	Prize        string `json:"prize"`
	Message      string `json:"message"`
	PreauthToken string `json:"preauth_token,omitempty"` // Returned if user was unregistered
}

// GetRouletteStatusResponse represents the current status of user's roulette
type GetRouletteStatusResponse struct {
	Roulette       *Roulette       `json:"roulette,omitempty"`
	Config         *RouletteConfig `json:"config,omitempty"`
	RemainingSpins int             `json:"remaining_spins"`
	CanSpin        bool            `json:"can_spin"`
	PrizeTaken     bool            `json:"prize_taken"`
}
