package bootstrap

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/metrics"
)

// TestMapClusterSnapshotToRequests_PersistsKubernetesHealthMetrics verifies
// Phase 12's fix: per-node/per-pod/per-deployment health data the collector
// already gathers (Ready, Restarts, AvailableReplicas, ...) used to be
// discarded here, leaving only 10 cluster-wide aggregate rows. It must now
// also be summarized into real (not fabricated) Kubernetes-category rows.
func TestMapClusterSnapshotToRequests_PersistsKubernetesHealthMetrics(t *testing.T) {
	cluster := RuntimeCluster{ID: uuid.New(), ProjectID: uuid.New(), Provider: constants.ClusterProviderKubernetes, CreatedBy: 1}

	snapshot := metrics.CollectedMetrics{
		CollectedAt: time.Now().UTC(),
		Cluster:     metrics.ClusterMetrics{UsageSource: "metrics-server"},
		Nodes: []metrics.NodeMetrics{
			{Name: "node-1", Ready: true},
			{Name: "node-2", Ready: false},
		},
		Pods: []metrics.PodMetrics{
			{Name: "pod-1", Ready: true, Restarts: 2},
			{Name: "pod-2", Ready: false, Restarts: 5},
		},
		Deployments: []metrics.DeploymentMetrics{
			{Name: "app-1", AvailableReplicas: 3, UnavailableReplicas: 1},
			{Name: "app-2", AvailableReplicas: 2, UnavailableReplicas: 0},
		},
	}

	requests := mapClusterSnapshotToRequests(cluster, snapshot)

	byName := make(map[string]float64, len(requests))
	labelsByName := make(map[string]string, len(requests))
	for _, request := range requests {
		byName[request.MetricName] = request.Value
		labelsByName[request.MetricName] = string(request.Labels)
	}

	cases := map[string]float64{
		"cluster.node.not_ready.count":                  1,
		"cluster.pod.restarts.total":                    7,
		"cluster.pod.not_ready.count":                   1,
		"cluster.deployment.available_replicas.total":   5,
		"cluster.deployment.unavailable_replicas.total": 1,
	}
	for name, want := range cases {
		got, ok := byName[name]
		if !ok {
			t.Fatalf("expected metric %q to be present, got %v", name, byName)
		}
		if got != want {
			t.Fatalf("metric %q: expected %v, got %v", name, want, got)
		}

		var labels map[string]string
		if err := json.Unmarshal([]byte(labelsByName[name]), &labels); err != nil {
			t.Fatalf("decode labels for %q: %v", name, err)
		}
		if labels["source"] != "kubernetes-api" {
			t.Fatalf("metric %q: expected source=kubernetes-api, got %v", name, labels)
		}
	}

	// CPU/memory/storage rows must carry the collector's reported usage
	// source, not a hardcoded label, so the frontend can tell live usage
	// from a requests-based estimate.
	var cpuLabels map[string]string
	if err := json.Unmarshal([]byte(labelsByName["cluster.cpu.usage.millicores"]), &cpuLabels); err != nil {
		t.Fatalf("decode cpu labels: %v", err)
	}
	if cpuLabels["source"] != "metrics-server" {
		t.Fatalf("expected cpu usage source metrics-server, got %v", cpuLabels)
	}
}

func TestMapClusterSnapshotToRequests_DefaultsUsageSourceWhenUnset(t *testing.T) {
	cluster := RuntimeCluster{ID: uuid.New(), ProjectID: uuid.New(), Provider: constants.ClusterProviderKubernetes, CreatedBy: 1}
	snapshot := metrics.CollectedMetrics{CollectedAt: time.Now().UTC()}

	requests := mapClusterSnapshotToRequests(cluster, snapshot)

	for _, request := range requests {
		if request.MetricType != constants.MetricTypeCPU {
			continue
		}
		var labels map[string]string
		if err := json.Unmarshal(request.Labels, &labels); err != nil {
			t.Fatalf("decode labels: %v", err)
		}
		if labels["source"] != "requested-capacity" {
			t.Fatalf("expected default source requested-capacity, got %v", labels)
		}
	}
}
