package domain

import "time"

type Event struct {
	ID       string    `json:"id"`
	Badge    string    `json:"badge"`
	Title    string    `json:"title"`
	Desc     string    `json:"desc"`
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
	ID       string     `json:"id"`
	Badge    string     `json:"badge"`
	Title    string     `json:"title"`
	Desc     string     `json:"desc"`
	Deadline time.Time  `json:"deadline"`
	Tags     string     `json:"tags"`
	Reward   []Reward   `json:"reward"`
	Info     string     `json:"info"`
	Status   string     `json:"status"`
	JoinedAt *time.Time `json:"joinedAt,omitempty"`
	HasPrise *bool      `json:"hasPriseStatus,omitempty"`
	PrizeTakenStatus bool `json:"prizeTakenStatus"`
}

type UserEventsResponse struct {
	Events []UserEventEntry `json:"events"`
}