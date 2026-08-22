package kubernetes

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsfake "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

type fakeMetricsCollectorClient struct {
	clientset kubernetes.Interface
}

func (f *fakeMetricsCollectorClient) RESTConfig() (*rest.Config, error) {
	return &rest.Config{}, nil
}

func (f *fakeMetricsCollectorClient) Clientset() (kubernetes.Interface, error) {
	return f.clientset, nil
}

func seedCollectorNode(t *testing.T, clientset kubernetes.Interface, name string, cpuCapacity, memoryCapacity string) {
	t.Helper()

	_, err := clientset.CoreV1().Nodes().Create(t.Context(), &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cpuCapacity),
				corev1.ResourceMemory: resource.MustParse(memoryCapacity),
			},
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed node %s: %v", name, err)
	}
}

func seedCollectorPod(t *testing.T, clientset kubernetes.Interface, name, node, cpuRequest, memoryRequest string) {
	t.Helper()

	_, err := clientset.CoreV1().Pods("default").Create(t.Context(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: corev1.PodSpec{
			NodeName: node,
			Containers: []corev1.Container{{
				Name: "app",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(cpuRequest),
						corev1.ResourceMemory: resource.MustParse(memoryRequest),
					},
				},
			}},
		},
		Status: corev1.PodStatus{
			Phase:      corev1.PodRunning,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed pod %s: %v", name, err)
	}
}

// seedNodeMetrics/seedPodMetrics work around the fake metrics clientset's
// object tracker guessing the wrong plural resource name for these types -
// the real typed client requests "nodes"/"pods" under metrics.k8s.io, so the
// tracker must be seeded with that exact GVR (see Phase 10's pod_health
// integration test for the same workaround).
func seedNodeMetrics(t *testing.T, clientset *metricsfake.Clientset, name string, cpuUsage, memoryUsage string) {
	t.Helper()

	err := clientset.Tracker().Create(
		schema.GroupVersionResource{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "nodes"},
		&metricsv1beta1.NodeMetrics{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Timestamp:  metav1.Now(),
			Usage: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cpuUsage),
				corev1.ResourceMemory: resource.MustParse(memoryUsage),
			},
		},
		"",
	)
	if err != nil {
		t.Fatalf("seed node metrics %s: %v", name, err)
	}
}

func seedPodMetrics(t *testing.T, clientset *metricsfake.Clientset, name, namespace, cpuUsage, memoryUsage string) {
	t.Helper()

	err := clientset.Tracker().Create(
		schema.GroupVersionResource{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "pods"},
		&metricsv1beta1.PodMetrics{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Timestamp:  metav1.Now(),
			Window:     metav1.Duration{Duration: 30 * time.Second},
			Containers: []metricsv1beta1.ContainerMetrics{{
				Name: "app",
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(cpuUsage),
					corev1.ResourceMemory: resource.MustParse(memoryUsage),
				},
			}},
		},
		namespace,
	)
	if err != nil {
		t.Fatalf("seed pod metrics %s/%s: %v", namespace, name, err)
	}
}

func TestMetricsCollector_FallsBackToRequestedCapacity_WhenMetricsServerUnavailable(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	seedCollectorNode(t, clientset, "node-1", "2", "4Gi")
	seedCollectorPod(t, clientset, "pod-1", "node-1", "500m", "1Gi")

	collector := NewMetricsCollector("test", &fakeMetricsCollectorClient{clientset: clientset})
	// No WithMetricsClientsetFactory override: the default factory will try
	// to reach a bogus empty REST config and fail, exercising the fallback.

	collected, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	if collected.Cluster.UsageSource != usageSourceRequestedEstimate {
		t.Fatalf("expected usage source %q, got %q", usageSourceRequestedEstimate, collected.Cluster.UsageSource)
	}
	if len(collected.Pods) != 1 || collected.Pods[0].CPUMilliCores != 500 {
		t.Fatalf("expected pod CPU to reflect the requested 500m, got %+v", collected.Pods)
	}
}

func TestMetricsCollector_UsesLiveUsage_WhenMetricsServerAvailable(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	seedCollectorNode(t, clientset, "node-1", "2", "4Gi")
	seedCollectorPod(t, clientset, "pod-1", "node-1", "500m", "1Gi")

	metricsClientset := metricsfake.NewSimpleClientset()
	seedNodeMetrics(t, metricsClientset, "node-1", "250m", "512Mi")
	seedPodMetrics(t, metricsClientset, "pod-1", "default", "42m", "64Mi")

	collector := NewMetricsCollector("test", &fakeMetricsCollectorClient{clientset: clientset}).
		WithMetricsClientsetFactory(func(_ *rest.Config) (metricsclientset.Interface, error) {
			return metricsClientset, nil
		})

	collected, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	if collected.Cluster.UsageSource != usageSourceMetricsServer {
		t.Fatalf("expected usage source %q, got %q", usageSourceMetricsServer, collected.Cluster.UsageSource)
	}
	if len(collected.Pods) != 1 || collected.Pods[0].CPUMilliCores != 42 {
		t.Fatalf("expected pod CPU to reflect live usage 42m, got %+v", collected.Pods)
	}
	if len(collected.Nodes) != 1 || collected.Nodes[0].CPUPercent == 0 {
		t.Fatalf("expected node CPU percent derived from live usage, got %+v", collected.Nodes)
	}
	// 250m / 2000m capacity = 12.5%
	if got := collected.Nodes[0].CPUPercent; got < 12 || got > 13 {
		t.Fatalf("expected node CPU percent ~12.5, got %v", got)
	}
}
