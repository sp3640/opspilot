package integration

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	k8sconfigmaps "github.com/sp3640/opspilot/backend/internal/kubernetes/configmaps"
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

func TestKubernetesConfigMapsAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachKubernetesConfigMapsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return clientset, nil
	})

	adminAToken := registerAndLogin(t, app.router, "ConfigMap Admin A", "configmap-admin-a@opspilot.dev", "password123")
	memberAToken := registerAndLogin(t, app.router, "ConfigMap Member A", "configmap-member-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "ConfigMap Admin B", "configmap-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "configmap-admin-a@opspilot.dev")
	memberA := mustGetUserByEmail(t, app.userRepo, "configmap-member-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "configmap-admin-b@opspilot.dev")
	if adminA.OrganizationID == nil || adminB.OrganizationID == nil {
		t.Fatalf("expected organizations for admins")
	}

	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, organizationB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "configmap-admin-b@opspilot.dev", "password123")
	if err := app.userRepo.AssignOrganizationAndRole(memberA.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member A role: %v", err)
	}
	memberAToken = loginOnly(t, app.router, "configmap-member-a@opspilot.dev", "password123")

	projectA := createProject(t, app.router, adminAToken, "ConfigMap Project A")
	projectB := createProject(t, app.router, adminBToken, "ConfigMap Project B")
	createCluster(t, app.router, adminAToken, projectA, "configmap-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "configmap-cluster-b")

	appAID := createApplicationForConfigMaps(t, app.router, adminAToken, projectA, "ConfigMap App A")
	appBID := createApplicationForConfigMaps(t, app.router, adminBToken, projectB, "ConfigMap App B")

	seedConfigMap(t, clientset, appAID, organizationA, "default", "cm-a")
	seedConfigMap(t, clientset, appBID, organizationB, "default", "cm-b")

	listAsMemberRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/configmaps", memberAToken, nil)
	assertStatus(t, listAsMemberRec, http.StatusOK)
	listItems := decodeConfigMapListItems(t, listAsMemberRec)
	if len(listItems) != 1 {
		t.Fatalf("expected one configmap for app A, got %d", len(listItems))
	}

	listAsAdminRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/configmaps", adminAToken, nil)
	assertStatus(t, listAsAdminRec, http.StatusOK)

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/configmaps/default/cm-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	configMapData := decodeDataMap(t, getRec)
	if configMapData["name"] != "cm-a" {
		t.Fatalf("expected configmap name cm-a")
	}
	if _, exists := configMapData["data"]; !exists {
		t.Fatalf("expected data in detail response")
	}
	if _, exists := configMapData["binaryData"]; !exists {
		t.Fatalf("expected binaryData metadata in detail response")
	}

	crossOrgRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/configmaps/default/cm-a?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, crossOrgRec, http.StatusForbidden)

	missingApplicationQueryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/configmaps/default/cm-a", memberAToken, nil)
	assertStatus(t, missingApplicationQueryRec, http.StatusBadRequest)

	configMapNotFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/configmaps/default/not-found?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, configMapNotFoundRec, http.StatusNotFound)

	clientset.PrependReactor("get", "configmaps", func(action ktesting.Action) (bool, runtime.Object, error) {
		getAction, ok := action.(ktesting.GetAction)
		if !ok {
			return false, nil, nil
		}
		if getAction.GetNamespace() != "missing-ns" {
			return false, nil, nil
		}

		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespaces"}, "missing-ns")
	})
	missingNamespaceRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/configmaps/missing-ns/cm-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, missingNamespaceRec, http.StatusNotFound)
}

func TestKubernetesConfigMapsAPI_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesConfigMapsRoutes(t, app, nil)

	adminToken := registerAndLogin(t, app.router, "ConfigMap Kubeconfig Admin", "configmap-kubeconfig-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "ConfigMap Kubeconfig Project")
	createCluster(t, app.router, adminToken, projectID, "configmap-kubeconfig-cluster")
	applicationID := createApplicationForConfigMaps(t, app.router, adminToken, projectID, "ConfigMap Kubeconfig App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/configmaps", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func TestKubernetesConfigMapsAPI_ClusterUnavailable(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesConfigMapsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return nil, &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:443: i/o timeout")}
	})

	adminToken := registerAndLogin(t, app.router, "ConfigMap Cluster Admin", "configmap-cluster-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "ConfigMap Cluster Project")
	createCluster(t, app.router, adminToken, projectID, "configmap-cluster-unavailable")
	applicationID := createApplicationForConfigMaps(t, app.router, adminToken, projectID, "ConfigMap Cluster App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/configmaps", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func attachKubernetesConfigMapsRoutes(t *testing.T, app *rbacTestApp, factory k8sconfigmaps.ClientsetFactory) {
	t.Helper()

	runtimeConfigMapService := k8sconfigmaps.NewConfigMapService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t))
	if factory != nil {
		runtimeConfigMapService.WithClientsetFactory(factory)
	}
	handler := handlers.NewKubernetesConfigMapHandler(runtimeConfigMapService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/applications/:id/configmaps", handler.ListByApplication)
	api.GET("/configmaps/:namespace/:name", handler.GetByNamespaceAndName)
}

func createApplicationForConfigMaps(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
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

func seedConfigMap(t *testing.T, clientset kubernetes.Interface, applicationID uuid.UUID, organizationID uuid.UUID, namespace, name string) {
	t.Helper()

	immutable := true
	_, err := clientset.CoreV1().ConfigMaps(namespace).Create(t.Context(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
			Annotations: map[string]string{"team": "platform"},
		},
		Data: map[string]string{
			"APP_ENV": "production",
			"TIMEOUT": "30",
		},
		BinaryData: map[string][]byte{
			"cert.pem": []byte("dummy-certificate-bytes"),
		},
		Immutable: &immutable,
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed configmap %s/%s: %v", namespace, name, err)
	}
}

func decodeConfigMapListItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode configmap list payload: %v", err)
	}

	return payload.Items
}
