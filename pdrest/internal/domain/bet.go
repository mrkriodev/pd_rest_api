package domain

import "time"

type Bet struct {
	ID         int        `json:"id"`
	UserID     string     `json:"userID"`
	Side       string     `json:"side"` // "pump" or "dump"
	Sum        float64    `json:"sum"`
	Pair       string     `json:"pair"`      // e.g., "ETH/USDT"
	Timeframe  int        `json:"timeframe"` // in seconds
	OpenPrice  float64    `json:"openPrice"`
	ClosePrice *float64   `json:"closePrice,omitempty"`
	OpenTime   time.Time  `json:"openTime"`
	CloseTime  *time.Time `json:"closeTime,omitempty"`
	Claimed    bool       `json:"claimedStatus"`
	PrizeStatus string    `json:"prizeStatus,omitempty"`
	CreatedAt  int64      `json:"created_at,omitempty"`
	UpdatedAt  int64      `json:"updated_at,omitempty"`
}

type OpenBetRequest struct {
	Side      string    `json:"side"` // "pump" or "dump"
	Sum       float64   `json:"sum"`
	Pair      string    `json:"pair"`
	Timeframe int       `json:"timeframe"`
	OpenPrice float64   `json:"openPrice"`
	OpenTime  time.Time `json:"openTime"`
}

type OpenBetResponse struct {
	ID int `json:"id"`
}

type BetStatusResponse struct {
	Side       string   `json:"side"`
	Sum        float64  `json:"sum"`
	Pair       string   `json:"pair"`
	Timeframe  int      `json:"timeframe"`
	OpenPrice  float64  `json:"openPrice"`
	ClosePrice *float64 `json:"closePrice,omitempty"`
	OpenTime   time.Time `json:"openTime"`
	Claimed    bool      `json:"claimedStatus"`
}
