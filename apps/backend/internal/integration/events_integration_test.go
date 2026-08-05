package integration

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	k8sevents "github.com/sp3640/opspilot/backend/internal/kubernetes/events"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestKubernetesEventsAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachKubernetesEventsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return clientset, nil
	})

	adminAToken := registerAndLogin(t, app.router, "Events Admin A", "events-admin-a@opspilot.dev", "password123")
	memberAToken := registerAndLogin(t, app.router, "Events Member A", "events-member-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Events Admin B", "events-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "events-admin-a@opspilot.dev")
	memberA := mustGetUserByEmail(t, app.userRepo, "events-member-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "events-admin-b@opspilot.dev")
	if adminA.OrganizationID == nil || adminB.OrganizationID == nil {
		t.Fatalf("expected organizations for admins")
	}

	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, organizationB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "events-admin-b@opspilot.dev", "password123")
	if err := app.userRepo.AssignOrganizationAndRole(memberA.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member A role: %v", err)
	}
	memberAToken = loginOnly(t, app.router, "events-member-a@opspilot.dev", "password123")

	projectA := createProject(t, app.router, adminAToken, "Events Project A")
	projectB := createProject(t, app.router, adminBToken, "Events Project B")
	createCluster(t, app.router, adminAToken, projectA, "events-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "events-cluster-b")

	appAID := createApplicationForEvents(t, app.router, adminAToken, projectA, "Events App A")
	appBID := createApplicationForEvents(t, app.router, adminBToken, projectB, "Events App B")

	seedEvent(t, clientset, appAID, organizationA, "default", "event-a")
	seedEvent(t, clientset, appBID, organizationB, "default", "event-b")

	listAsMemberRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/events", memberAToken, nil)
	assertStatus(t, listAsMemberRec, http.StatusOK)
	listItems := decodeEventListItems(t, listAsMemberRec)
	if len(listItems) != 1 {
		t.Fatalf("expected one event for app A, got %d", len(listItems))
	}

	listAsAdminRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/events", adminAToken, nil)
	assertStatus(t, listAsAdminRec, http.StatusOK)

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/events/default/event-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	eventData := decodeDataMap(t, getRec)
	if eventData["name"] != "event-a" {
		t.Fatalf("expected event name event-a")
	}

	crossOrgRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/events/default/event-a?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, crossOrgRec, http.StatusForbidden)

	missingApplicationQueryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/events/default/event-a", memberAToken, nil)
	assertStatus(t, missingApplicationQueryRec, http.StatusBadRequest)

	invalidApplicationRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/not-a-uuid/events", memberAToken, nil)
	assertStatus(t, invalidApplicationRec, http.StatusBadRequest)

	applicationNotFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+uuid.NewString()+"/events", memberAToken, nil)
	assertStatus(t, applicationNotFoundRec, http.StatusNotFound)

	eventNotFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/events/default/not-found?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, eventNotFoundRec, http.StatusNotFound)

	clientset.PrependReactor("get", "events", func(action ktesting.Action) (bool, runtime.Object, error) {
		getAction, ok := action.(ktesting.GetAction)
		if !ok {
			return false, nil, nil
		}
		if getAction.GetNamespace() != "missing-ns" {
			return false, nil, nil
		}

		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespaces"}, "missing-ns")
	})
	missingNamespaceRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/events/missing-ns/event-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, missingNamespaceRec, http.StatusNotFound)
}

func TestKubernetesEventsAPI_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesEventsRoutes(t, app, nil)

	adminToken := registerAndLogin(t, app.router, "Event Kubeconfig Admin", "event-kubeconfig-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Event Kubeconfig Project")
	createCluster(t, app.router, adminToken, projectID, "event-kubeconfig-cluster")
	applicationID := createApplicationForEvents(t, app.router, adminToken, projectID, "Event Kubeconfig App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/events", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func TestKubernetesEventsAPI_ClusterUnavailable(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesEventsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return nil, &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:443: i/o timeout")}
	})

	adminToken := registerAndLogin(t, app.router, "Event Cluster Admin", "event-cluster-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Event Cluster Project")
	createCluster(t, app.router, adminToken, projectID, "event-cluster-unavailable")
	applicationID := createApplicationForEvents(t, app.router, adminToken, projectID, "Event Cluster App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/events", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func attachKubernetesEventsRoutes(t *testing.T, app *rbacTestApp, factory k8sevents.ClientsetFactory) {
	t.Helper()

	runtimeEventService := k8sevents.NewEventService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t))
	if factory != nil {
		runtimeEventService.WithClientsetFactory(factory)
	}
	handler := handlers.NewKubernetesEventHandler(runtimeEventService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/applications/:id/events", handler.ListByApplication)
	api.GET("/events/:namespace/:name", handler.GetByNamespaceAndName)
}

func createApplicationForEvents(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
	t.Helper()

	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", token, map[string]any{
		"name":    name,
		"runtime": constants.ApplicationRuntimeGo,
		"port":    8080,
	})
	assertStatus(t, rec, http.StatusCreated)

	id := decodeDataMap(t, rec)["id"].(string)
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("parse application id: %v", err)
	}

	return parsed
}

func seedEvent(t *testing.T, clientset kubernetes.Interface, applicationID uuid.UUID, organizationID uuid.UUID, namespace, name string) {
	t.Helper()

	now := metav1.NewTime(time.Now().UTC())
	first := metav1.NewTime(now.Add(-5 * time.Minute))
	_, err := clientset.CoreV1().Events(namespace).Create(t.Context(), &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
			Annotations: map[string]string{"team": "platform"},
		},
		Reason:         "Started",
		Message:        "Container started",
		Type:           corev1.EventTypeNormal,
		Count:          1,
		FirstTimestamp: first,
		LastTimestamp:  now,
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Namespace: namespace, Name: "app-pod"},
		Source:         corev1.EventSource{Component: "kubelet", Host: "node-1"},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed event %s/%s: %v", namespace, name, err)
	}
}

func decodeEventListItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode event list payload: %v", err)
	}

	return payload.Items
}
