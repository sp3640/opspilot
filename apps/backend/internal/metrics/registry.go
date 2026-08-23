package metrics

import (
	"context"
	"sync"
	"time"
)

// collectorTimeout bounds a single collector's Supports+Collect call. Without
// this, one unreachable/slow cluster would block CollectAll's sequential
// loop indefinitely, delaying metric persistence and alert reconciliation
// for every other cluster behind it in the same tick.
const collectorTimeout = 30 * time.Second

type Registry struct {
	mu         sync.RWMutex
	collectors map[string]MetricsCollector
}

func NewRegistry() *Registry {
	return &Registry{
		collectors: make(map[string]MetricsCollector),
	}
}

func (r *Registry) Register(collector MetricsCollector) {
	if r == nil || collector == nil {
		return
	}

	r.mu.Lock()
	r.collectors[collector.Name()] = collector
	r.mu.Unlock()
}

func (r *Registry) Unregister(name string) {
	if r == nil {
		return
	}

	r.mu.Lock()
	delete(r.collectors, name)
	r.mu.Unlock()
}

func (r *Registry) CollectAll(ctx context.Context) ([]CollectedMetrics, []error) {
	if r == nil {
		return nil, nil
	}

	r.mu.RLock()
	collectors := make([]MetricsCollector, 0, len(r.collectors))
	for _, collector := range r.collectors {
		collectors = append(collectors, collector)
	}
	r.mu.RUnlock()

	results := make([]CollectedMetrics, 0, len(collectors))
	errs := make([]error, 0)

	for _, collector := range collectors {
		if collector == nil {
			continue
		}

		metrics, err := collectOne(ctx, collector)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if metrics == nil {
			continue
		}

		results = append(results, *metrics)
	}

	return results, errs
}

// collectOne runs one collector's Supports+Collect pair under a bounded
// timeout so a single hung/unreachable cluster can never stall the rest of
// CollectAll's batch.
func collectOne(ctx context.Context, collector MetricsCollector) (*CollectedMetrics, error) {
	collectCtx, cancel := context.WithTimeout(ctx, collectorTimeout)
	defer cancel()

	if !collector.Supports(collectCtx) {
		return nil, nil
	}

	return collector.Collect(collectCtx)
}
