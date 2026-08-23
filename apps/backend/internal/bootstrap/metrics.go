package bootstrap

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/alerting"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	kubeintegration "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/metrics"
)

const defaultSchedulerBufferSize = 256

type SnapshotStore interface {
	StoreSnapshot(ctx context.Context, projectID uuid.UUID, userID uint, requests []dto.CreateMetricRequest) ([]dto.MetricResponse, error)
}

// AlertReconciler turns a batch of currently-firing conditions into Alert
// rows (create/update/resolve/reopen as appropriate). Satisfied by
// *services.AlertService's ReconcileConditions - kept as an interface here
// so this package never needs to import internal/services.
type AlertReconciler interface {
	ReconcileConditions(ctx context.Context, projectID uuid.UUID, userID uint, source string, conditions []alerting.EvaluatedCondition) error
}

type MetricsBootstrap struct {
	mu sync.RWMutex

	registry *metrics.Registry
	store    SnapshotStore
	alerts   AlertReconciler

	interval time.Duration

	cancel context.CancelFunc
	wg     sync.WaitGroup

	queue chan metricSnapshot

	collectors map[string]RuntimeCluster
	running    bool
}

type metricSnapshot struct {
	projectID uuid.UUID
	userID    uint
	metrics   []dto.CreateMetricRequest
}

func NewMetricsBootstrap(registry *metrics.Registry, store SnapshotStore, interval time.Duration) *MetricsBootstrap {
	if registry == nil {
		registry = metrics.NewRegistry()
	}
	if interval <= 0 {
		interval = defaultMetricsInterval
	}

	return &MetricsBootstrap{
		registry:   registry,
		store:      store,
		interval:   interval,
		queue:      make(chan metricSnapshot, defaultSchedulerBufferSize),
		collectors: make(map[string]RuntimeCluster),
	}
}

func (m *MetricsBootstrap) WithAlertReconciler(reconciler AlertReconciler) *MetricsBootstrap {
	if m != nil {
		m.alerts = reconciler
	}

	return m
}

func (m *MetricsBootstrap) Registry() *metrics.Registry {
	if m == nil {
		return nil
	}

	return m.registry
}

func (m *MetricsBootstrap) RegisterKubernetesCollector(cluster RuntimeCluster) {
	if m == nil || m.registry == nil {
		return
	}
	if strings.TrimSpace(strings.ToUpper(cluster.Provider)) != constants.ClusterProviderKubernetes {
		return
	}
	if cluster.ID == uuid.Nil || cluster.ProjectID == uuid.Nil {
		return
	}

	collectorName := collectorNameForCluster(cluster.ID)
	collector := kubeintegration.NewMetricsCollector(collectorName, kubeintegration.NewClient(cluster.KubeconfigEncrypted))
	m.registry.Register(collector)

	m.mu.Lock()
	m.collectors[collectorName] = cluster
	m.mu.Unlock()
}

func (m *MetricsBootstrap) Start(ctx context.Context) error {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.running = true
	m.mu.Unlock()

	m.wg.Add(2)
	go m.collectLoop(runCtx)
	go m.persistLoop(runCtx)

	return nil
}

func (m *MetricsBootstrap) Stop(ctx context.Context) error {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}
	cancel := m.cancel
	m.cancel = nil
	m.running = false
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *MetricsBootstrap) collectLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collectAndQueueSafely(ctx)
		}
	}
}

// collectAndQueueSafely isolates one tick's collection from a panic deep in
// a Kubernetes client call - without this, a single malformed API response
// would crash the whole process, taking down every other in-flight HTTP
// request on this replica along with it, not just this background job.
func (m *MetricsBootstrap) collectAndQueueSafely(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, "metrics collection panicked", slog.Any("panic", r))
		}
	}()

	m.collectAndQueue(ctx)
}

func (m *MetricsBootstrap) persistLoop(ctx context.Context) {
	defer m.wg.Done()

	for {
		select {
		case snapshot := <-m.queue:
			m.persistSnapshotSafely(ctx, snapshot)
		case <-ctx.Done():
			for {
				select {
				case snapshot := <-m.queue:
					m.persistSnapshotSafely(context.Background(), snapshot)
				default:
					return
				}
			}
		}
	}
}

// persistSnapshotSafely isolates one snapshot's persistence from a panic,
// for the same reason as collectAndQueueSafely above.
func (m *MetricsBootstrap) persistSnapshotSafely(ctx context.Context, snapshot metricSnapshot) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, "metric snapshot persistence panicked", slog.Any("panic", r))
		}
	}()

	m.persistSnapshot(ctx, snapshot)
}

func (m *MetricsBootstrap) collectAndQueue(ctx context.Context) {
	if m.registry == nil || m.store == nil {
		return
	}

	collected, _ := m.registry.CollectAll(ctx)
	for _, snapshot := range collected {
		cluster, ok := m.clusterForCollector(snapshot.CollectorName)
		if !ok {
			continue
		}

		if m.alerts != nil {
			conditions := alerting.Evaluate(alerting.EvaluateInput{
				ClusterID:   cluster.ID.String(),
				ClusterName: cluster.Name,
				Snapshot:    snapshot,
			})
			if err := m.alerts.ReconcileConditions(ctx, cluster.ProjectID, cluster.CreatedBy, constants.AlertSourceKubernetes, conditions); err != nil {
				logger.Error(ctx, "alert reconciliation failed", slog.String("cluster_id", cluster.ID.String()), slog.Any("error", err))
			}
		}

		requests := mapClusterSnapshotToRequests(cluster, snapshot)
		if len(requests) == 0 {
			continue
		}

		select {
		case m.queue <- metricSnapshot{projectID: cluster.ProjectID, userID: cluster.CreatedBy, metrics: requests}:
		default:
			m.persistSnapshot(ctx, metricSnapshot{projectID: cluster.ProjectID, userID: cluster.CreatedBy, metrics: requests})
		}
	}
}

func (m *MetricsBootstrap) persistSnapshot(ctx context.Context, snapshot metricSnapshot) {
	if m.store == nil || len(snapshot.metrics) == 0 {
		return
	}

	_, _ = m.store.StoreSnapshot(ctx, snapshot.projectID, snapshot.userID, snapshot.metrics)
}

func (m *MetricsBootstrap) clusterForCollector(name string) (RuntimeCluster, bool) {
	m.mu.RLock()
	cluster, ok := m.collectors[name]
	m.mu.RUnlock()

	return cluster, ok
}

func collectorNameForCluster(clusterID uuid.UUID) string {
	return "kubernetes:" + clusterID.String()
}

func mapClusterSnapshotToRequests(cluster RuntimeCluster, snapshot metrics.CollectedMetrics) []dto.CreateMetricRequest {
	now := snapshot.CollectedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	source := "requested-capacity"
	if snapshot.Cluster.UsageSource != "" {
		source = snapshot.Cluster.UsageSource
	}
	labels := json.RawMessage(`{"source":"` + source + `"}`)
	metadata := json.RawMessage(`{}`)
	// Kubernetes-category rows (restarts/readiness/replica availability) are
	// never estimates - they come straight from the Kubernetes API objects
	// already collected above, so they always carry this fixed source label
	// regardless of whether CPU/memory used the live-usage or requests path.
	k8sLabels := json.RawMessage(`{"source":"kubernetes-api"}`)

	notReadyNodes := 0
	for _, node := range snapshot.Nodes {
		if !node.Ready {
			notReadyNodes++
		}
	}

	var podRestarts int64
	notReadyPods := 0
	for _, pod := range snapshot.Pods {
		podRestarts += int64(pod.Restarts)
		if !pod.Ready {
			notReadyPods++
		}
	}

	var availableReplicas, unavailableReplicas int64
	for _, deployment := range snapshot.Deployments {
		availableReplicas += int64(deployment.AvailableReplicas)
		unavailableReplicas += int64(deployment.UnavailableReplicas)
	}

	requests := []dto.CreateMetricRequest{
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeCPU,
			MetricName:   "cluster.cpu.capacity.millicores",
			Value:        float64(snapshot.Cluster.CPUCapacityMilliCores),
			Unit:         "m",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeCPU,
			MetricName:   "cluster.cpu.usage.millicores",
			Value:        float64(snapshot.Cluster.CPUUsageMilliCores),
			Unit:         "m",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeMemory,
			MetricName:   "cluster.memory.capacity.bytes",
			Value:        float64(snapshot.Cluster.MemoryCapacityBytes),
			Unit:         "bytes",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeMemory,
			MetricName:   "cluster.memory.usage.bytes",
			Value:        float64(snapshot.Cluster.MemoryUsageBytes),
			Unit:         "bytes",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeStorage,
			MetricName:   "cluster.storage.capacity.bytes",
			Value:        float64(snapshot.Cluster.StorageCapacityBytes),
			Unit:         "bytes",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeStorage,
			MetricName:   "cluster.storage.usage.bytes",
			Value:        float64(snapshot.Cluster.StorageUsageBytes),
			Unit:         "bytes",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeReplicaCount,
			MetricName:   "cluster.deployment.count",
			Value:        float64(snapshot.Cluster.DeploymentCount),
			Unit:         "count",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeAvailability,
			MetricName:   "cluster.node.count",
			Value:        float64(snapshot.Cluster.NodeCount),
			Unit:         "count",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeAvailability,
			MetricName:   "cluster.pod.count",
			Value:        float64(snapshot.Cluster.PodCount),
			Unit:         "count",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeAvailability,
			MetricName:   "cluster.namespace.count",
			Value:        float64(snapshot.Cluster.NamespaceCount),
			Unit:         "count",
			Timestamp:    now,
			Labels:       labels,
			Metadata:     metadata,
		},
		// Kubernetes-category health, derived from data already collected
		// above (no additional API calls) - real, not estimated.
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeAvailability,
			MetricName:   "cluster.node.not_ready.count",
			Value:        float64(notReadyNodes),
			Unit:         "count",
			Timestamp:    now,
			Labels:       k8sLabels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeRestartCount,
			MetricName:   "cluster.pod.restarts.total",
			Value:        float64(podRestarts),
			Unit:         "count",
			Timestamp:    now,
			Labels:       k8sLabels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeAvailability,
			MetricName:   "cluster.pod.not_ready.count",
			Value:        float64(notReadyPods),
			Unit:         "count",
			Timestamp:    now,
			Labels:       k8sLabels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeReplicaCount,
			MetricName:   "cluster.deployment.available_replicas.total",
			Value:        float64(availableReplicas),
			Unit:         "count",
			Timestamp:    now,
			Labels:       k8sLabels,
			Metadata:     metadata,
		},
		{
			ProjectID:    cluster.ProjectID,
			ClusterID:    cluster.ID,
			ResourceID:   cluster.ID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeReplicaCount,
			MetricName:   "cluster.deployment.unavailable_replicas.total",
			Value:        float64(unavailableReplicas),
			Unit:         "count",
			Timestamp:    now,
			Labels:       k8sLabels,
			Metadata:     metadata,
		},
	}

	return requests
}
