package monitoring

import "time"

type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusStopped   JobStatus = "STOPPED"
)

type Job struct {
	ID          string
	ConnectorID string
	Status      JobStatus
	StartedAt   *time.Time
	CompletedAt *time.Time
	Duration    time.Duration
	Error       string
	Metadata    map[string]any
}
