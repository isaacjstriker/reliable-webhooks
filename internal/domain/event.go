package domain

import "time"

type EventStatus string

const (
	StatusReceived  EventStatus = "received"
	StatusProcessing EventStatus = "processing"
	StatusProcessed  EventStatus = "processed"
	StatusFailed     EventStatus = "failed"
)

type Event struct {
	ID          int64
	Provider    string
	EventID     string
	Payload     []byte
	Status      EventStatus
	Attempts    int
	LastError   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ProcessedAt *time.Time
}