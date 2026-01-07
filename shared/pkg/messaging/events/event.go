package events

import "time"

type Event struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	OccurredAt time.Time `json:"occurred_at"`
	Data       any       `json:"data"`
}
