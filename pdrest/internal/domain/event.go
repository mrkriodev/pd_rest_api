package domain

import "time"

type Event struct {
	ID       string    `json:"id"`
	Badge    string    `json:"badge"`
	Title    string    `json:"title"`
	Desc     string    `json:"desc"`
	StartTime time.Time `json:"startTime"`
	Deadline time.Time `json:"deadline"`
	Tags     string    `json:"tags"`
	Reward   []Reward  `json:"reward"`
	Info     string    `json:"info"`
}

type Reward struct {
	Place string `json:"place"`
	Value string `json:"value"`
}

type EventsResponse struct {
	Events []Event `json:"events"`
}

type UserEventEntry struct {
	ID               string     `json:"id"`
	Badge            string     `json:"badge"`
	Title            string     `json:"title"`
	Desc             string     `json:"desc"`
	StartTime        time.Time  `json:"startTime"`
	Deadline         time.Time  `json:"deadline"`
	Tags             string     `json:"tags"`
	Reward           []Reward   `json:"reward"`
	Info             string     `json:"info"`
	Status           string     `json:"status"`
	JoinedAt         *time.Time `json:"joinedAt,omitempty"`
	HasPrise         *bool      `json:"hasPriseStatus,omitempty"`
	PrizeDesc        *string    `json:"prizeDesc,omitempty"`
	PrizeTakenStatus bool       `json:"prizeTakenStatus"`
}

type UserEventsResponse struct {
	Events []UserEventEntry `json:"events"`
}

type BetPrizeLeaderboardEntry struct {
	UserUUID  string `json:"userUUID"`
	WinCount  int    `json:"winCount"`
	LossCount int    `json:"lossCount"`
	NetPoints int64  `json:"netPoints"`
}

type EventProgressResponse struct {
	EventID         string `json:"eventId"`
	Participating   bool   `json:"participating"`
	CollectedPoints int64  `json:"collectedPoints"`
}

type EventLeaderResponse struct {
	LeaderImage string `json:"leader_image,omitempty"`
	Points      int64  `json:"points"`
}