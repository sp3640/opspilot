package pods

import (
	"context"

	"github.com/sp3640/opspilot/backend/internal/dto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricsClientsetFactory builds a metrics.k8s.io client from a kubeconfig,
// mirroring ClientsetFactory. Kept as a separate factory (rather than reusing
// ClientsetFactory) because metrics-server is a distinct, optional API group
// that many clusters do not install.
type MetricsClientsetFactory func(kubeconfig []byte) (metricsclientset.Interface, error)

// fetchPodMetrics returns live per-container CPU/memory usage for a single
// pod. Any error (metrics-server not installed, API unavailable, pod metrics
// not yet collected) is treated as "not available" rather than a request
// failure, since usage data is inherently best-effort.
func fetchPodMetrics(ctx context.Context, factory MetricsClientsetFactory, kubeconfig []byte, namespace, name string) *dto.PodMetricsResponse {
	if factory == nil {
		return nil
	}

	clientset, err := factory(kubeconfig)
	if err != nil {
		return nil
	}

	podMetrics, err := clientset.MetricsV1beta1().PodMetricses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	containers := make([]dto.ContainerMetricsResponse, 0, len(podMetrics.Containers))
	for _, container := range podMetrics.Containers {
		cpu := container.Usage[corev1.ResourceCPU]
		memory := container.Usage[corev1.ResourceMemory]
		containers = append(containers, dto.ContainerMetricsResponse{
			Name:   container.Name,
			CPU:    cpu.String(),
			Memory: memory.String(),
		})
	}

	return &dto.PodMetricsResponse{
		Timestamp:  podMetrics.Timestamp.Time,
		Window:     podMetrics.Window.Duration.String(),
		Containers: containers,
	}
}
