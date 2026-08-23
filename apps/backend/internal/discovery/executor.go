package discovery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/logger"
)

var (
	ErrDiscoveryAlreadyRunning = errors.New("discovery already running for cluster")
	ErrDiscoveryExecutorClosed = errors.New("discovery executor is stopped")
)

// jobTimeout bounds a single cluster's discovery run. Without this, a
// cluster whose kubeconfig points somewhere that TCP-hangs (firewalled, no
// RST) rather than fails fast would hold its semaphore slot forever -
// wedging discovery for every other cluster behind it once maxConcurrent
// slots are all stuck this way (with the default maxConcurrent=1, a single
// hung cluster is enough to wedge all of them).
const jobTimeout = 2 * time.Minute

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "PENDING"
	ExecutionStatusRunning   ExecutionStatus = "RUNNING"
	ExecutionStatusCompleted ExecutionStatus = "COMPLETED"
	ExecutionStatusFailed    ExecutionStatus = "FAILED"
	ExecutionStatusCanceled  ExecutionStatus = "CANCELED"
)

type ExecutionRecord struct {
	ClusterID    uuid.UUID
	StartedAt    *time.Time
	CompletedAt  *time.Time
	Duration     time.Duration
	Status       ExecutionStatus
	Error        string
	cancel       context.CancelFunc
	runCompleted chan struct{}
}

type ExecutionFunc func(ctx context.Context, clusterID uuid.UUID) error

type DiscoveryExecutor struct {
	mu            sync.RWMutex
	maxConcurrent int
	semaphore     chan struct{}
	activeJobs    map[uuid.UUID]*ExecutionRecord
	stopped       bool
	wg            sync.WaitGroup
}

func NewDiscoveryExecutor(maxConcurrent int) *DiscoveryExecutor {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}

	return &DiscoveryExecutor{
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
		activeJobs:    make(map[uuid.UUID]*ExecutionRecord),
	}
}

func (e *DiscoveryExecutor) Execute(ctx context.Context, clusterID uuid.UUID, execute ExecutionFunc) error {
	if execute == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	e.mu.Lock()
	if e.stopped {
		e.mu.Unlock()
		return ErrDiscoveryExecutorClosed
	}
	if _, exists := e.activeJobs[clusterID]; exists {
		e.mu.Unlock()
		return ErrDiscoveryAlreadyRunning
	}

	jobCtx, cancel := context.WithCancel(ctx)
	record := &ExecutionRecord{
		ClusterID:    clusterID,
		Status:       ExecutionStatusPending,
		cancel:       cancel,
		runCompleted: make(chan struct{}),
	}
	e.activeJobs[clusterID] = record
	e.wg.Add(1)
	e.mu.Unlock()

	go e.run(jobCtx, record, execute)

	return nil
}

func (e *DiscoveryExecutor) run(ctx context.Context, record *ExecutionRecord, execute ExecutionFunc) {
	defer e.wg.Done()

	select {
	case e.semaphore <- struct{}{}:
	case <-ctx.Done():
		e.finish(record.ClusterID, ExecutionStatusCanceled, ctx.Err())
		return
	}
	defer func() {
		<-e.semaphore
	}()

	startedAt := time.Now().UTC()
	e.mu.Lock()
	record.StartedAt = &startedAt
	record.Status = ExecutionStatusRunning
	e.mu.Unlock()

	// A panic here (e.g. a malformed API response deep in a Kubernetes
	// client call) must not take down the whole process - only every other
	// in-flight HTTP request would be affected too, since nothing else
	// recovers panics outside the Gin middleware chain.
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, "discovery job panicked",
				slog.String("cluster_id", record.ClusterID.String()),
				slog.Any("panic", r),
			)
			e.finish(record.ClusterID, ExecutionStatusFailed, fmt.Errorf("discovery job panicked: %v", r))
		}
	}()

	runCtx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	err := execute(runCtx, record.ClusterID)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			e.finish(record.ClusterID, ExecutionStatusCanceled, err)
			return
		}

		e.finish(record.ClusterID, ExecutionStatusFailed, err)
		return
	}

	e.finish(record.ClusterID, ExecutionStatusCompleted, nil)
}

func (e *DiscoveryExecutor) Cancel(clusterID uuid.UUID) error {
	e.mu.RLock()
	record, ok := e.activeJobs[clusterID]
	e.mu.RUnlock()
	if !ok {
		return nil
	}

	record.cancel()
	return nil
}

func (e *DiscoveryExecutor) CancelAll() {
	e.mu.RLock()
	clusterIDs := make([]uuid.UUID, 0, len(e.activeJobs))
	for clusterID := range e.activeJobs {
		clusterIDs = append(clusterIDs, clusterID)
	}
	e.mu.RUnlock()

	for _, clusterID := range clusterIDs {
		_ = e.Cancel(clusterID)
	}
}

func (e *DiscoveryExecutor) Running(clusterID uuid.UUID) bool {
	e.mu.RLock()
	_, exists := e.activeJobs[clusterID]
	e.mu.RUnlock()

	return exists
}

func (e *DiscoveryExecutor) ActiveJobs() []ExecutionRecord {
	e.mu.RLock()
	defer e.mu.RUnlock()

	jobs := make([]ExecutionRecord, 0, len(e.activeJobs))
	for _, record := range e.activeJobs {
		jobs = append(jobs, snapshotExecutionRecord(record))
	}

	return jobs
}

func (e *DiscoveryExecutor) Stop() {
	e.mu.Lock()
	e.stopped = true
	e.mu.Unlock()

	e.CancelAll()
	e.wg.Wait()

	e.mu.Lock()
	e.stopped = false
	e.mu.Unlock()
}

func (e *DiscoveryExecutor) finish(clusterID uuid.UUID, status ExecutionStatus, runErr error) {
	e.mu.Lock()
	record, ok := e.activeJobs[clusterID]
	if !ok {
		e.mu.Unlock()
		return
	}

	completedAt := time.Now().UTC()
	record.CompletedAt = &completedAt
	if record.StartedAt != nil {
		record.Duration = completedAt.Sub(*record.StartedAt)
	}
	record.Status = status
	if runErr != nil {
		record.Error = runErr.Error()
	} else {
		record.Error = ""
	}
	close(record.runCompleted)
	delete(e.activeJobs, clusterID)
	e.mu.Unlock()
}

func snapshotExecutionRecord(record *ExecutionRecord) ExecutionRecord {
	if record == nil {
		return ExecutionRecord{}
	}

	return ExecutionRecord{
		ClusterID:   record.ClusterID,
		StartedAt:   record.StartedAt,
		CompletedAt: record.CompletedAt,
		Duration:    record.Duration,
		Status:      record.Status,
		Error:       record.Error,
	}
}
