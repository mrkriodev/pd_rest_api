package domain

// RatingSource enumerates supported rating sources.
type RatingSource string

const (
	RatingSourceFromEvent    RatingSource = "from_event"
	RatingSourceBetBonus     RatingSource = "bet_bonus"
	RatingSourcePromoBonus   RatingSource = "promo_bonus"
	RatingSourceServiceBonus RatingSource = "servivce_bonus"
)

// RatingTotals aggregates points per source for a user.
type RatingTotals struct {
	FromEvent    int64 `json:"from_event"`
	BetBonus     int64 `json:"bet_bonus"`
	PromoBonus   int64 `json:"promo_bonus"`
	ServiceBonus int64 `json:"servivce_bonus"`
}

// TotalPoints returns the sum of all point sources.
func (t RatingTotals) TotalPoints() int64 {
	return t.FromEvent + t.BetBonus + t.PromoBonus + t.ServiceBonus
}

// UserAssets represents a user's points portfolio.
type UserAssets struct {
	UserID      string       `json:"userID"`
	Points      RatingTotals `json:"points"`
	TotalPoints int64        `json:"total_points"`
}

// GlobalRatingEntry represents a single entry in the global rating.
type GlobalRatingEntry struct {
	UserID string `json:"userID"`
	Value  int64  `json:"value"`
}

// FriendRatingEntry represents aggregated points for a referred friend.
type FriendRatingEntry struct {
	UserID string `json:"userId"`
	Value  int64  `json:"value"`
}
