package domain

// PrizeType represents the type of prize
type PrizeType string

const (
	PrizeTypeRouletteOnStart     PrizeType = "roulette_on_start"
	PrizeTypeRouletteDuringEvent PrizeType = "roulette_during_event"
	PrizeTypeEventReward         PrizeType = "event_reward"
)

// Prize represents a prize awarded to a user
type Prize struct {
	ID             int       `json:"id"`
	EventID        *string   `json:"event_id,omitempty"`
	UserID         *string   `json:"userID"`                     // Now mandatory
	PrizeValueID   *int      `json:"prize_value_id,omitempty"`   // Reference to prize_values table
	PreauthTokenID *int      `json:"preauth_token_id,omitempty"` // Optional, for tracking
	RouletteID     *int      `json:"roulette_id,omitempty"`
	PrizeValue     string    `json:"prize_value"` // Prize value text (e.g., "100 USDT")
	PrizeType      PrizeType `json:"prize_type"`
	AwardedAt      int64     `json:"awarded_at"`
	CreatedAt      int64     `json:"created_at"`
}

// CreatePrizeRequest represents a request to create a prize
type CreatePrizeRequest struct {
	EventID        *string   `json:"event_id,omitempty"`
	UserID         *string   `json:"userID"`
	PrizeValueID   *int      `json:"prize_value_id,omitempty"`
	PreauthTokenID *int      `json:"preauth_token_id,omitempty"`
	RouletteID     *int      `json:"roulette_id,omitempty"`
	PrizeValue     string    `json:"prize_value"`
	PrizeType      PrizeType `json:"prize_type"`
}
