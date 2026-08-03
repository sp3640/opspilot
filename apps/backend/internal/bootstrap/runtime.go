package bootstrap

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultDiscoveryInterval = 5 * time.Minute
	defaultMetricsInterval   = 1 * time.Minute
)

type ClusterCatalog interface {
	LoadEnabledClusters(ctx context.Context) ([]RuntimeCluster, error)
}

type DiscoveryRunner interface {
	Start(ctx context.Context)
	Stop()
	RegisterCluster(clusterID uuid.UUID)
	UnregisterCluster(clusterID uuid.UUID)
}

type DiscoveryExecutor interface {
	CancelAll()
	Stop()
}

type Runtime struct {
	mu sync.Mutex

	catalog ClusterCatalog

	discoveryRunner   DiscoveryRunner
	discoveryExecutor DiscoveryExecutor
	metricsBootstrap  *MetricsBootstrap

	started bool
}

type RuntimeDependencies struct {
	Catalog            ClusterCatalog
	DiscoveryRunner    DiscoveryRunner
	DiscoveryExecutor  DiscoveryExecutor
	MetricsBootstrap   *MetricsBootstrap
	DiscoveryInterval  time.Duration
	MetricsInterval    time.Duration
	DiscoveryBootstrap *DiscoveryBootstrap
}

func NewRuntime(deps RuntimeDependencies) *Runtime {
	runtime := &Runtime{
		catalog:           deps.Catalog,
		discoveryRunner:   deps.DiscoveryRunner,
		discoveryExecutor: deps.DiscoveryExecutor,
		metricsBootstrap:  deps.MetricsBootstrap,
	}

	if deps.DiscoveryBootstrap != nil {
		runtime.discoveryRunner = deps.DiscoveryBootstrap.Runner
		runtime.discoveryExecutor = deps.DiscoveryBootstrap.Executor
	}

	return runtime
}

func (r *Runtime) Startup(ctx context.Context) error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	clusters := make([]RuntimeCluster, 0)
	if r.catalog != nil {
		loaded, err := r.catalog.LoadEnabledClusters(ctx)
		if err != nil {
			return err
		}
		clusters = loaded
	}

	for _, cluster := range clusters {
		if r.discoveryRunner != nil {
			r.discoveryRunner.RegisterCluster(cluster.ID)
		}

		if r.metricsBootstrap != nil {
			r.metricsBootstrap.RegisterKubernetesCollector(cluster)
		}
	}

	if r.discoveryRunner != nil {
		r.discoveryRunner.Start(ctx)
	}

	if r.metricsBootstrap != nil {
		if err := r.metricsBootstrap.Start(ctx); err != nil {
			if r.discoveryRunner != nil {
				r.discoveryRunner.Stop()
			}
			return err
		}
	}

	r.mu.Lock()
	r.started = true
	r.mu.Unlock()

	return nil
}

func (r *Runtime) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	if !r.started {
		r.mu.Unlock()
		return nil
	}
	r.started = false
	r.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	if r.discoveryExecutor != nil {
		r.discoveryExecutor.CancelAll()
	}

	if r.discoveryRunner != nil {
		r.discoveryRunner.Stop()
	}

	if r.discoveryExecutor != nil {
		r.discoveryExecutor.Stop()
	}

	if r.metricsBootstrap != nil {
		if err := r.metricsBootstrap.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
