package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	k8spods "github.com/sp3640/opspilot/backend/internal/kubernetes/pods"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestKubernetesPodsAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachPodsRoutes(t, app, clientset)

	adminAToken := registerAndLogin(t, app.router, "Pods Admin A", "pods-admin-a@opspilot.dev", "password123")
	memberAToken := registerAndLogin(t, app.router, "Pods Member A", "pods-member-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Pods Admin B", "pods-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "pods-admin-a@opspilot.dev")
	memberA := mustGetUserByEmail(t, app.userRepo, "pods-member-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "pods-admin-b@opspilot.dev")
	if adminA.OrganizationID == nil || adminB.OrganizationID == nil {
		t.Fatalf("expected organizations for admins")
	}
	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, organizationB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "pods-admin-b@opspilot.dev", "password123")
	if err := app.userRepo.AssignOrganizationAndRole(memberA.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member A to org A: %v", err)
	}
	memberAToken = loginOnly(t, app.router, "pods-member-a@opspilot.dev", "password123")

	projectA := createProject(t, app.router, adminAToken, "Pods Project A")
	projectB := createProject(t, app.router, adminBToken, "Pods Project B")
	createCluster(t, app.router, adminAToken, projectA, "pods-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "pods-cluster-b")

	appAID := createApplicationForPods(t, app.router, adminAToken, projectA, "Pods App A")
	appBID := createApplicationForPods(t, app.router, adminBToken, projectB, "Pods App B")

	seedPod(t, clientset, appAID, organizationA, "default", "pods-a-1")
	seedPod(t, clientset, appAID, organizationA, "staging", "pods-a-2")
	seedPod(t, clientset, appBID, organizationB, "default", "pods-b-1")

	listARec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/pods", memberAToken, nil)
	assertStatus(t, listARec, http.StatusOK)
	listAItems := decodePodListItems(t, listARec)
	if len(listAItems) != 2 {
		t.Fatalf("expected 2 pods for app A, got %d", len(listAItems))
	}

	listANamespaceRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/pods?namespace=default", memberAToken, nil)
	assertStatus(t, listANamespaceRec, http.StatusOK)
	listANamespaceItems := decodePodListItems(t, listANamespaceRec)
	if len(listANamespaceItems) != 1 {
		t.Fatalf("expected 1 pod for app A in default namespace, got %d", len(listANamespaceItems))
	}

	getPodRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pods-a-1?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, getPodRec, http.StatusOK)
	pod := decodeDataMap(t, getPodRec)
	if pod["name"] != "pods-a-1" {
		t.Fatalf("expected pod name pods-a-1")
	}

	crossTenantRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pods-a-1?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, crossTenantRec, http.StatusForbidden)

	missingQueryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pods-a-1", memberAToken, nil)
	assertStatus(t, missingQueryRec, http.StatusBadRequest)

	invalidAppIDRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/not-a-uuid/pods", memberAToken, nil)
	assertStatus(t, invalidAppIDRec, http.StatusBadRequest)

	notFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/does-not-exist?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, notFoundRec, http.StatusNotFound)
}

func attachPodsRoutes(t *testing.T, app *rbacTestApp, clientset kubernetes.Interface) {
	t.Helper()

	podService := k8spods.NewPodService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t)).WithClientsetFactory(
		func(_ []byte) (kubernetes.Interface, error) {
			return clientset, nil
		},
	)
	podHandler := handlers.NewPodHandler(podService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/applications/:id/pods", podHandler.ListByApplication)
	api.GET("/pods/:namespace/:name", podHandler.GetByNamespaceAndName)
}

func createApplicationForPods(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
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

func seedPod(t *testing.T, clientset kubernetes.Interface, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) {
	t.Helper()

	_, err := clientset.CoreV1().Pods(namespace).Create(t.Context(), &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  "app",
				Image: "ghcr.io/opspilot/app:v1",
			}},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{{
				Name:         "app",
				Image:        "ghcr.io/opspilot/app:v1",
				Ready:        true,
				RestartCount: 1,
				State:        v1.ContainerState{Running: &v1.ContainerStateRunning{}},
			}},
			Conditions: []v1.PodCondition{{
				Type:   v1.PodReady,
				Status: v1.ConditionTrue,
			}},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed pod %s/%s: %v", namespace, name, err)
	}
}

func decodePodListItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode pod list payload: %v", err)
	}

	return payload.Items
}
