package integration

import (
	"net/http"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/handlers"
	k8sevents "github.com/sp3640/opspilot/backend/internal/kubernetes/events"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// TestPodEventsAPIIntegration verifies the Phase 10 pod-scoped events route:
// only real Kubernetes events whose involvedObject matches the requested pod
// are returned, ownership is enforced via the pod's own labels (events
// themselves carry no opspilot labels), and cross-org access is denied.
func TestPodEventsAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachPodsRoutes(t, app, clientset)
	attachPodEventsRoute(t, app, clientset)

	adminAToken := registerAndLogin(t, app.router, "Pod Events Admin A", "pod-events-admin-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Pod Events Admin B", "pod-events-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "pod-events-admin-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "pod-events-admin-b@opspilot.dev")
	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID

	projectA := createProject(t, app.router, adminAToken, "Pod Events Project A")
	projectB := createProject(t, app.router, adminBToken, "Pod Events Project B")
	createCluster(t, app.router, adminAToken, projectA, "pod-events-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "pod-events-cluster-b")

	appAID := createApplicationForPods(t, app.router, adminAToken, projectA, "Pod Events App A")
	appBID := createApplicationForPods(t, app.router, adminBToken, projectB, "Pod Events App B")

	seedPod(t, clientset, appAID, organizationA, "default", "target-pod")
	seedPod(t, clientset, appBID, organizationB, "default", "other-org-pod")

	// A real cluster event never carries opspilot labels (kubelet/controllers
	// create them, not opspilot) - the route must not rely on event labels.
	seedRawEvent(t, clientset, "default", "target-pod-scheduled", "target-pod", "Scheduled", "Successfully assigned default/target-pod to node-1")
	seedRawEvent(t, clientset, "default", "target-pod-pulled", "target-pod", "Pulled", "Container image already present on machine")
	seedRawEvent(t, clientset, "default", "other-pod-event", "other-org-pod", "Scheduled", "Successfully assigned default/other-org-pod to node-1")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/target-pod/events?applicationId="+appAID.String(), adminAToken, nil)
	assertStatus(t, rec, http.StatusOK)
	items := decodeEventListItems(t, rec)
	if len(items) != 2 {
		t.Fatalf("expected 2 events for target-pod, got %d: %v", len(items), items)
	}
	for _, item := range items {
		reason, _ := item["reason"].(string)
		if reason != "Scheduled" && reason != "Pulled" {
			t.Fatalf("unexpected event leaked into pod events list: %v", item)
		}
	}

	crossOrgRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/target-pod/events?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, crossOrgRec, http.StatusForbidden)

	missingQueryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/target-pod/events", adminAToken, nil)
	assertStatus(t, missingQueryRec, http.StatusBadRequest)

	notFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/does-not-exist/events?applicationId="+appAID.String(), adminAToken, nil)
	assertStatus(t, notFoundRec, http.StatusNotFound)
}

func attachPodEventsRoute(t *testing.T, app *rbacTestApp, clientset kubernetes.Interface) {
	t.Helper()

	eventService := k8sevents.NewEventService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t)).
		WithClientsetFactory(func(_ []byte) (kubernetes.Interface, error) { return clientset, nil })
	handler := handlers.NewKubernetesEventHandler(eventService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/pods/:namespace/:name/events", handler.GetPodEvents)
}

func seedRawEvent(t *testing.T, clientset kubernetes.Interface, namespace, name, involvedPodName, reason, message string) {
	t.Helper()

	_, err := clientset.CoreV1().Events(namespace).Create(t.Context(), &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Reason:         reason,
		Message:        message,
		Type:           corev1.EventTypeNormal,
		Count:          1,
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Namespace: namespace, Name: involvedPodName},
		Source:         corev1.EventSource{Component: "kubelet", Host: "node-1"},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed raw event %s/%s: %v", namespace, name, err)
	}
}
