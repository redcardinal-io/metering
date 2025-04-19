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
	// ID of the organization that user belongs to.
	Organization string `json:"organization"`
	// The ID of the user that owns the event.
	User string `json:"user"`
	// The event time.
	Timestamp string `json:"timestamp"`
	// The event data as a JSON string.
	Properties string `json:"properties"`
	// The time the event was ingested.
	IngestedAt *time.Time `json:"ingested_at,omitempty"`
}

type eventInput struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Source       string `json:"source"`
	Organization string `json:"organization"`
	User         string `json:"user"`
	Timestamp    string `json:"timestamp"`
	Properties   string `json:"properties"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (e *Event) UnmarshalJSON(data []byte) error {
	var input eventInput
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	// Copy the fields to the actual Event struct
	e.ID = input.ID
	e.Type = input.Type
	e.Source = input.Source
	e.Organization = input.Organization
	e.User = input.User
	e.Timestamp = input.Timestamp
	e.Properties = input.Properties

	return nil
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

// PublishEventsResult contains information about the batch processing result
type PublishEventsResult struct {
	SuccessCount int            // Number of events successfully published
	FailedEvents []*FailedEvent // Details about failed events
	Error        error          // Overall error, if any
}

// FailedEvent contains information about a single event that failed processing
type FailedEvent struct {
	Event *Event
	Error error
}
