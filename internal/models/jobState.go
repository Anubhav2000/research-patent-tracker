package models

type JobState struct {
	Queries      map[string]int `json:"queries"`       // Tracks last processed offset per query
	TotalFetched int            `json:"total_fetched"` // Tracks total papers fetched
	IsPaused     bool           `json:"is_paused"`
}
