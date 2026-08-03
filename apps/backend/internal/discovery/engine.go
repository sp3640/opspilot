package discovery

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/monitoring"
)

var ErrJobNotFound = errors.New("discovery job not found")

type DiscoveryEngine struct {
	mu         sync.RWMutex
	activeJobs map[string]DiscoveryJob
}

func NewDiscoveryEngine() *DiscoveryEngine {
	return &DiscoveryEngine{
		activeJobs: make(map[string]DiscoveryJob),
	}
}

func (e *DiscoveryEngine) StartJob(provider monitoring.ProviderType, projectID, clusterID uuid.UUID) DiscoveryJob {
	now := time.Now().UTC()
	job := DiscoveryJob{
		ID:        uuid.NewString(),
		Provider:  provider,
		ProjectID: projectID,
		ClusterID: clusterID,
		Status:    DiscoveryJobStatusRunning,
		StartedAt: &now,
	}

	e.mu.Lock()
	e.activeJobs[job.ID] = job
	e.mu.Unlock()

	return job
}

func (e *DiscoveryEngine) FinishJob(jobID string, result *DiscoveryResult, jobErr error) (DiscoveryJob, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	job, ok := e.activeJobs[jobID]
	if !ok {
		return DiscoveryJob{}, ErrJobNotFound
	}

	completedAt := time.Now().UTC()
	job.CompletedAt = &completedAt
	if job.StartedAt != nil {
		job.Duration = completedAt.Sub(*job.StartedAt)
	}

	if result != nil {
		job.DiscoveredCount = len(result.Resources)
		if result.Duration == 0 {
			result.Duration = job.Duration
		}
	}

	if jobErr != nil {
		job.Status = DiscoveryJobStatusFailed
		job.Error = jobErr.Error()
	} else {
		job.Status = DiscoveryJobStatusCompleted
		job.Error = ""
	}

	delete(e.activeJobs, jobID)

	return job, nil
}

func (e *DiscoveryEngine) CancelJob(jobID string) (DiscoveryJob, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	job, ok := e.activeJobs[jobID]
	if !ok {
		return DiscoveryJob{}, ErrJobNotFound
	}

	completedAt := time.Now().UTC()
	job.CompletedAt = &completedAt
	if job.StartedAt != nil {
		job.Duration = completedAt.Sub(*job.StartedAt)
	}
	job.Status = DiscoveryJobStatusCanceled
	job.Error = "discovery job canceled"

	delete(e.activeJobs, jobID)

	return job, nil
}

func (e *DiscoveryEngine) RetryFailedJob(job DiscoveryJob) DiscoveryJob {
	return e.StartJob(job.Provider, job.ProjectID, job.ClusterID)
}

func (e *DiscoveryEngine) ActiveJobs() []DiscoveryJob {
	e.mu.RLock()
	defer e.mu.RUnlock()

	jobs := make([]DiscoveryJob, 0, len(e.activeJobs))
	for _, job := range e.activeJobs {
		jobs = append(jobs, job)
	}

	return jobs
}

func (e *DiscoveryEngine) GetActiveJob(jobID string) (DiscoveryJob, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	job, ok := e.activeJobs[jobID]
	return job, ok
}
