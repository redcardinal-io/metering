package models

import (
	"encoding/json"
	"time"
)

type Event struct {
	// The event ID.
	ID string `json:"id"`
	// The event type.
	Type string `json:"type"`
	// The event source.
	Source string `json:"source"`
	// The event source metadata.
	SourceMetadata map[string]string `json:"source_metadata"`
	// ID of the organization that user belongs to.
	Organization string `json:"organization"`
	// The ID of the user that owns the event.
	User string `json:"user"`
	// The event time.
	Time time.Time `json:"time"`
	// The event data as a JSON string.
	Properties map[string]any `json:"properties"`
	// The time the event was ingested.
	IngestedAt time.Time
	// The time the event was stored.
	ValidationErrors []error
}

// ToJSON serializes the event to JSON for Kafka
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes JSON into an event
func (e *Event) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// EventBatch represents a batch of events
type EventBatch struct {
	Events []*Event `json:"events"`
}
