package metrics

import "context"

type MetricsCollector interface {
	Collect(ctx context.Context) (*CollectedMetrics, error)
	Supports(ctx context.Context) bool
	Name() string
}

func Collect(ctx context.Context, collector MetricsCollector) (*CollectedMetrics, error) {
	if collector == nil {
		return nil, nil
	}
	if !collector.Supports(ctx) {
		return nil, nil
	}

	return collector.Collect(ctx)
}
