package models

import "time"

type IngestionJobStatus string

const (
	JobPending  IngestionJobStatus = "PENDING"
	JobRunning  IngestionJobStatus = "RUNNING"
	JobFailed   IngestionJobStatus = "FAILED"
	JobComplete IngestionJobStatus = "COMPLETE"
)

type IngestionJob struct {
	ID        string
	Wallet    string
	Chain     string
	Status    IngestionJobStatus
	Cursor    string
	UpdatedAt time.Time
}

type Transfer struct {
	Hash      string
	From      string
	To        string
	Token     string
	Value     string
	Timestamp time.Time
	Raw       map[string]any
}

type Balance struct {
	Token  string
	Symbol string
	Amount string
	USD    float64
	Raw    map[string]any
}
