package domain

// Achievement represents an achievement that users can earn.
type Achievement struct {
	ID       string  `json:"id"`
	Badge    string  `json:"badge"`
	Title    string  `json:"title"`
	ImageURL string  `json:"imageUrl"`
	Desc     string  `json:"desc"`
	Tags     string  `json:"tags"`
	Summ     float64 `json:"summ"`
	Steps    int     `json:"steps"`
	StepDesc string  `json:"stepDesc"`
}

// UserAchievement represents an achievement earned by a user.
type UserAchievement struct {
	UserID        string `json:"userID"`
	AchievementID string `json:"achievementID"`
}

// AchievementsResponse represents the response for available achievements.
type AchievementsResponse struct {
	Achievements []Achievement `json:"achievements"`
}

// UserAchievementsResponse represents the response for user achievements.
type UserAchievementsResponse struct {
	Achievements []Achievement `json:"achievements"`
}
