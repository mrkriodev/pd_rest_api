package domain

// Achievement represents an achievement that users can earn.
type Achievement struct {
	ID       string `json:"id"`
	Badge    string `json:"badge"`
	Title    string `json:"title"`
	ImageURL string `json:"imageUrl"`
	Desc     string `json:"desc"`
	Tags     string `json:"tags"`
	Steps    int    `json:"steps"`
	StepDesc string `json:"stepDesc"`
	PrizeID  *int   `json:"prizeId,omitempty"`
}

// UserAchievement represents an achievement earned by a user.
type UserAchievement struct {
	UserID        string `json:"userID"`
	AchievementID string `json:"achievementID"`
}

type UserAchievementStatus struct {
	UserID        string `json:"userID"`
	AchievementID string `json:"achievementID"`
	StepsGot      int    `json:"stepsGot"`
	NeedSteps     int    `json:"needSteps"`
	ClaimedStatus bool   `json:"claimedStatus"`
}

type UserAchievementEntry struct {
	ID            string  `json:"id"`
	Badge         string  `json:"badge"`
	Title         string  `json:"title"`
	ImageURL      string  `json:"imageUrl"`
	Desc          string  `json:"desc"`
	Tags          string  `json:"tags"`
	PrizeDesc     *string `json:"prizeDesc,omitempty"`
	Steps         int     `json:"steps"`
	StepDesc      string  `json:"stepDesc"`
	StepsGot      *int    `json:"stepsGot,omitempty"`
	NeedSteps     *int    `json:"needSteps,omitempty"`
	ClaimedStatus bool    `json:"claimedStatus"`
}

// AchievementsResponse represents the response for available achievements.
type AchievementsResponse struct {
	Achievements []Achievement `json:"achievements"`
}

// UserAchievementsResponse represents the response for user achievements.
type UserAchievementsResponse struct {
	Achievements []UserAchievementEntry `json:"achievements"`
}
