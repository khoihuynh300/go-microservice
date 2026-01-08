package events

import "time"

type Event struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	OccurredAt time.Time `json:"occurred_at"`
	TraceID    string    `json:"trace_id"`
	Data       any       `json:"data"`
}
