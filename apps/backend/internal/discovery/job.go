package discovery

import (
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/monitoring"
)

type DiscoveryJobStatus string

const (
	DiscoveryJobStatusPending   DiscoveryJobStatus = "PENDING"
	DiscoveryJobStatusRunning   DiscoveryJobStatus = "RUNNING"
	DiscoveryJobStatusCompleted DiscoveryJobStatus = "COMPLETED"
	DiscoveryJobStatusFailed    DiscoveryJobStatus = "FAILED"
	DiscoveryJobStatusCanceled  DiscoveryJobStatus = "CANCELED"
)

type DiscoveryJob struct {
	ID              string
	Provider        monitoring.ProviderType
	ProjectID       uuid.UUID
	ClusterID       uuid.UUID
	Status          DiscoveryJobStatus
	StartedAt       *time.Time
	CompletedAt     *time.Time
	Duration        time.Duration
	Error           string
	DiscoveredCount int
}
