package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sp3640/opspilot/backend/internal/handlers"
	k8spods "github.com/sp3640/opspilot/backend/internal/kubernetes/pods"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsfake "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

// TestPodHealthDetail_EnrichedFields verifies that a single pod fetch surfaces
// the richer Phase 10 fields: reason-when-unhealthy, per-container crash
// state/exit code, resource requests/limits, and probe presence.
func TestPodHealthDetail_EnrichedFields(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachPodsRoutes(t, app, clientset)

	adminToken := registerAndLogin(t, app.router, "Pod Health Admin", "pod-health-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Pod Health Project")
	createCluster(t, app.router, adminToken, projectID, "pod-health-cluster")
	applicationID := createApplicationForPods(t, app.router, adminToken, projectID, "Pod Health App")

	adminUser := mustGetUserByEmail(t, app.userRepo, "pod-health-admin@opspilot.dev")
	organizationID := *adminUser.OrganizationID

	exitCode := int32(137)
	lastFinishedAt := metav1.NewTime(time.Now().Add(-10 * time.Minute))
	_, err := clientset.CoreV1().Pods("default").Create(t.Context(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crash-pod",
			Namespace: "default",
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "app",
				Image: "ghcr.io/opspilot/app:v2",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("500m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
				ReadinessProbe: &corev1.Probe{},
				LivenessProbe:  &corev1.Probe{},
			}},
		},
		Status: corev1.PodStatus{
			Phase:  corev1.PodFailed,
			Reason: "Evicted",
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:         "app",
				Image:        "ghcr.io/opspilot/app:v2",
				Ready:        false,
				RestartCount: 4,
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{
						Reason:  "CrashLoopBackOff",
						Message: "back-off 5m0s restarting failed container",
					},
				},
				LastTerminationState: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						Reason:     "OOMKilled",
						ExitCode:   exitCode,
						FinishedAt: lastFinishedAt,
					},
				},
			}},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed crash pod: %v", err)
	}

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/crash-pod?applicationId="+applicationID.String(), adminToken, nil)
	assertStatus(t, rec, http.StatusOK)
	pod := decodeDataMap(t, rec)

	if pod["reason"] != "Evicted" {
		t.Fatalf("expected pod reason Evicted, got %v", pod["reason"])
	}
	if pod["readyContainerCount"].(float64) != 0 {
		t.Fatalf("expected 0 ready containers, got %v", pod["readyContainerCount"])
	}

	containerStatuses, ok := pod["containerStatuses"].([]any)
	if !ok || len(containerStatuses) != 1 {
		t.Fatalf("expected 1 container status, got %v", pod["containerStatuses"])
	}
	status := containerStatuses[0].(map[string]any)

	if status["state"] != "CrashLoopBackOff" {
		t.Fatalf("expected state CrashLoopBackOff, got %v", status["state"])
	}
	if status["stateMessage"] != "back-off 5m0s restarting failed container" {
		t.Fatalf("expected state message, got %v", status["stateMessage"])
	}
	if status["lastTerminationReason"] != "OOMKilled" {
		t.Fatalf("expected last termination reason OOMKilled, got %v", status["lastTerminationReason"])
	}
	if status["lastTerminationExitCode"].(float64) != 137 {
		t.Fatalf("expected last termination exit code 137, got %v", status["lastTerminationExitCode"])
	}
	if status["cpuRequest"] != "100m" || status["cpuLimit"] != "500m" {
		t.Fatalf("expected cpu request/limit, got %v/%v", status["cpuRequest"], status["cpuLimit"])
	}
	if status["memoryRequest"] != "128Mi" || status["memoryLimit"] != "256Mi" {
		t.Fatalf("expected memory request/limit, got %v/%v", status["memoryRequest"], status["memoryLimit"])
	}
	if status["hasReadinessProbe"] != true || status["hasLivenessProbe"] != true {
		t.Fatalf("expected readiness/liveness probes to be reported as configured")
	}

	// No metrics-server wired in this test -> metrics must be absent, not fabricated.
	if _, present := pod["metrics"]; present {
		t.Fatalf("expected no metrics field when metrics-server is unavailable, got %v", pod["metrics"])
	}
}

// TestPodHealthDetail_LiveMetrics verifies that when a metrics-server-backed
// client is available, real CPU/memory usage is attached to the pod detail.
func TestPodHealthDetail_LiveMetrics(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()

	podMetrics := &metricsv1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{Name: "metered-pod", Namespace: "default"},
		Timestamp:  metav1.Now(),
		Window:     metav1.Duration{Duration: 30 * time.Second},
		Containers: []metricsv1beta1.ContainerMetrics{{
			Name: "app",
			Usage: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("42m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
		}},
	}
	// The fake clientset's object tracker guesses a resource name from the Go
	// type ("podmetricses"), but the real typed client requests "pods" under
	// the metrics.k8s.io group - seed the tracker with the exact GVR the
	// client will ask for so the fake actually resolves the lookup.
	metricsClientset := metricsfake.NewSimpleClientset()
	if err := metricsClientset.Tracker().Create(
		schema.GroupVersionResource{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "pods"},
		podMetrics,
		"default",
	); err != nil {
		t.Fatalf("seed pod metrics: %v", err)
	}

	podService := k8spods.NewPodService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t)).
		WithClientsetFactory(func(_ []byte) (kubernetes.Interface, error) { return clientset, nil }).
		WithMetricsClientsetFactory(func(_ []byte) (metricsclientset.Interface, error) { return metricsClientset, nil })
	podHandler := handlers.NewPodHandler(podService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/applications/:id/pods", podHandler.ListByApplication)
	api.GET("/pods/:namespace/:name", podHandler.GetByNamespaceAndName)

	adminToken := registerAndLogin(t, app.router, "Pod Metrics Admin", "pod-metrics-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Pod Metrics Project")
	createCluster(t, app.router, adminToken, projectID, "pod-metrics-cluster")
	applicationID := createApplicationForPods(t, app.router, adminToken, projectID, "Pod Metrics App")

	adminUser := mustGetUserByEmail(t, app.userRepo, "pod-metrics-admin@opspilot.dev")
	organizationID := *adminUser.OrganizationID

	_, err := clientset.CoreV1().Pods("default").Create(context.Background(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metered-pod",
			Namespace: "default",
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "ghcr.io/opspilot/app:v1"}}},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "app", Ready: true,
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
			}},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed metered pod: %v", err)
	}

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/metered-pod?applicationId="+applicationID.String(), adminToken, nil)
	assertStatus(t, rec, http.StatusOK)
	pod := decodeDataMap(t, rec)

	metrics, ok := pod["metrics"].(map[string]any)
	if !ok {
		t.Fatalf("expected metrics to be present, got %v", pod["metrics"])
	}
	containers, ok := metrics["containers"].([]any)
	if !ok || len(containers) != 1 {
		t.Fatalf("expected 1 container metric, got %v", metrics["containers"])
	}
	container := containers[0].(map[string]any)
	if container["cpu"] != "42m" || container["memory"] != "64Mi" {
		t.Fatalf("expected cpu=42m memory=64Mi, got cpu=%v memory=%v", container["cpu"], container["memory"])
	}
}
