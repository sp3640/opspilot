package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const defaultDiscoveryInterval = 5 * time.Minute

type BackgroundRunner struct {
	mu       sync.RWMutex
	interval time.Duration
	worker   *DiscoveryWorker
	executor *DiscoveryExecutor
	clusters map[uuid.UUID]struct{}
	running  bool
	cancel   context.CancelFunc
	done     chan struct{}
}

func NewBackgroundRunner(interval time.Duration, worker *DiscoveryWorker, executor *DiscoveryExecutor) *BackgroundRunner {
	if interval <= 0 {
		interval = defaultDiscoveryInterval
	}
	if executor == nil {
		executor = NewDiscoveryExecutor(1)
	}

	return &BackgroundRunner{
		interval: interval,
		worker:   worker,
		executor: executor,
		clusters: make(map[uuid.UUID]struct{}),
	}
}

func (r *BackgroundRunner) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}

	runnerCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.done = make(chan struct{})
	r.running = true
	r.mu.Unlock()

	go r.loop(runnerCtx)
}

func (r *BackgroundRunner) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}

	cancel := r.cancel
	done := r.done
	r.running = false
	r.cancel = nil
	r.done = nil
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	if r.executor != nil {
		r.executor.Stop()
	}
}

func (r *BackgroundRunner) RegisterCluster(clusterID uuid.UUID) {
	r.mu.Lock()
	r.clusters[clusterID] = struct{}{}
	r.mu.Unlock()
}

func (r *BackgroundRunner) UnregisterCluster(clusterID uuid.UUID) {
	r.mu.Lock()
	delete(r.clusters, clusterID)
	r.mu.Unlock()

	if r.executor != nil {
		_ = r.executor.Cancel(clusterID)
	}
}

func (r *BackgroundRunner) Trigger(clusterID uuid.UUID) error {
	r.mu.RLock()
	running := r.running
	r.mu.RUnlock()
	if !running {
		return nil
	}

	return r.submit(context.Background(), clusterID)
}

func (r *BackgroundRunner) Running() bool {
	r.mu.RLock()
	running := r.running
	r.mu.RUnlock()

	return running
}

func (r *BackgroundRunner) loop(ctx context.Context) {
	defer close(r.done)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, clusterID := range r.registeredClusters() {
				if ctx.Err() != nil {
					return
				}
				_ = r.submit(ctx, clusterID)
			}
		}
	}
}

func (r *BackgroundRunner) submit(ctx context.Context, clusterID uuid.UUID) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if r.executor == nil || r.worker == nil {
		return nil
	}
	if r.executor.Running(clusterID) {
		return nil
	}

	return r.executor.Execute(ctx, clusterID, func(jobCtx context.Context, clusterID uuid.UUID) error {
		_, err := r.worker.Run(jobCtx, clusterID)
		return err
	})
}

func (r *BackgroundRunner) registeredClusters() []uuid.UUID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clusterIDs := make([]uuid.UUID, 0, len(r.clusters))
	for clusterID := range r.clusters {
		clusterIDs = append(clusterIDs, clusterID)
	}

	return clusterIDs
}
