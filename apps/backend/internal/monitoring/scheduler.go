package monitoring

import (
	"sync"
	"time"
)

type SchedulerState string

const (
	SchedulerStateStopped SchedulerState = "STOPPED"
	SchedulerStateRunning SchedulerState = "RUNNING"
)

type SchedulerStatus struct {
	State          SchedulerState
	StartedAt      *time.Time
	ConnectorCount int
	JobCount       int
}

type Scheduler interface {
	Start() error
	Stop() error
	Register(connector Connector)
	Remove(connectorID string)
	Jobs() []Job
	Status() SchedulerStatus
}

type InMemoryScheduler struct {
	mu        sync.RWMutex
	registry  *ConnectorRegistry
	jobs      []Job
	state     SchedulerState
	startedAt *time.Time
}

func NewInMemoryScheduler(registry *ConnectorRegistry) *InMemoryScheduler {
	if registry == nil {
		registry = NewConnectorRegistry()
	}

	return &InMemoryScheduler{
		registry: registry,
		jobs:     make([]Job, 0),
		state:    SchedulerStateStopped,
	}
}

func (s *InMemoryScheduler) Start() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == SchedulerStateRunning {
		return nil
	}

	now := time.Now()
	s.startedAt = &now
	s.state = SchedulerStateRunning

	return nil
}

func (s *InMemoryScheduler) Stop() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = SchedulerStateStopped

	return nil
}

func (s *InMemoryScheduler) Register(connector Connector) {
	if s == nil || s.registry == nil {
		return
	}

	s.registry.Register(connector)
}

func (s *InMemoryScheduler) Remove(connectorID string) {
	if s == nil || s.registry == nil {
		return
	}

	s.registry.Unregister(connectorID)
}

func (s *InMemoryScheduler) Jobs() []Job {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]Job, 0, len(s.jobs))
	jobs = append(jobs, s.jobs...)

	return jobs
}

func (s *InMemoryScheduler) Status() SchedulerStatus {
	if s == nil {
		return SchedulerStatus{State: SchedulerStateStopped}
	}

	s.mu.RLock()
	state := s.state
	startedAt := s.startedAt
	jobCount := len(s.jobs)
	s.mu.RUnlock()

	connectorCount := 0
	if s.registry != nil {
		connectorCount = len(s.registry.GetAll())
	}

	return SchedulerStatus{
		State:          state,
		StartedAt:      startedAt,
		ConnectorCount: connectorCount,
		JobCount:       jobCount,
	}
}
