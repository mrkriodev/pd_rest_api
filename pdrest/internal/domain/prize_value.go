package domain

// PrizeValue represents an available prize value for an event
type PrizeValue struct {
	ID        int     `json:"id"`
	EventID   string  `json:"event_id"`
	Value     int64   `json:"value"`                // Prize value in points (exact points to add to user balance)
	Label     string  `json:"label"`                // Display label (e.g., "100 USDT")
	SegmentID *string `json:"segment_id,omitempty"` // Optional segment ID for roulette wheel
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}
