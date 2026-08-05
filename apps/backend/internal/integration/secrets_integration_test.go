package integration

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	k8ssecrets "github.com/sp3640/opspilot/backend/internal/kubernetes/secrets"
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

func TestKubernetesSecretsAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachKubernetesSecretsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return clientset, nil
	})

	adminAToken := registerAndLogin(t, app.router, "Secrets Admin A", "secrets-admin-a@opspilot.dev", "password123")
	memberAToken := registerAndLogin(t, app.router, "Secrets Member A", "secrets-member-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Secrets Admin B", "secrets-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "secrets-admin-a@opspilot.dev")
	memberA := mustGetUserByEmail(t, app.userRepo, "secrets-member-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "secrets-admin-b@opspilot.dev")
	if adminA.OrganizationID == nil || adminB.OrganizationID == nil {
		t.Fatalf("expected organizations for admins")
	}

	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, organizationB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "secrets-admin-b@opspilot.dev", "password123")
	if err := app.userRepo.AssignOrganizationAndRole(memberA.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member A role: %v", err)
	}
	memberAToken = loginOnly(t, app.router, "secrets-member-a@opspilot.dev", "password123")

	projectA := createProject(t, app.router, adminAToken, "Secrets Project A")
	projectB := createProject(t, app.router, adminBToken, "Secrets Project B")
	createCluster(t, app.router, adminAToken, projectA, "secrets-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "secrets-cluster-b")

	appAID := createApplicationForSecrets(t, app.router, adminAToken, projectA, "Secrets App A")
	appBID := createApplicationForSecrets(t, app.router, adminBToken, projectB, "Secrets App B")

	seedSecret(t, clientset, appAID, organizationA, "default", "secret-a")
	seedSecret(t, clientset, appBID, organizationB, "default", "secret-b")

	listAsMemberRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/secrets", memberAToken, nil)
	assertStatus(t, listAsMemberRec, http.StatusOK)
	listItems := decodeSecretListItems(t, listAsMemberRec)
	if len(listItems) != 1 {
		t.Fatalf("expected one secret for app A, got %d", len(listItems))
	}

	listAsAdminRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/secrets", adminAToken, nil)
	assertStatus(t, listAsAdminRec, http.StatusOK)

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/secrets/default/secret-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	secretData := decodeDataMap(t, getRec)
	if secretData["name"] != "secret-a" {
		t.Fatalf("expected secret name secret-a")
	}
	if _, exists := secretData["keys"]; !exists {
		t.Fatalf("expected keys in detail response")
	}
	if _, exists := secretData["data"]; exists {
		t.Fatalf("secret data must never be exposed")
	}
	if _, exists := secretData["binaryData"]; exists {
		t.Fatalf("secret binaryData must never be exposed")
	}
	if _, exists := secretData["stringData"]; exists {
		t.Fatalf("secret stringData must never be exposed")
	}
	if strings.Contains(getRec.Body.String(), "s3cr3t") || strings.Contains(getRec.Body.String(), "czNjcjN0") {
		t.Fatalf("secret value leaked in response body")
	}

	crossOrgRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/secrets/default/secret-a?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, crossOrgRec, http.StatusForbidden)

	missingApplicationQueryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/secrets/default/secret-a", memberAToken, nil)
	assertStatus(t, missingApplicationQueryRec, http.StatusBadRequest)

	secretNotFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/secrets/default/not-found?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, secretNotFoundRec, http.StatusNotFound)

	clientset.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
		getAction, ok := action.(ktesting.GetAction)
		if !ok {
			return false, nil, nil
		}
		if getAction.GetNamespace() != "missing-ns" {
			return false, nil, nil
		}

		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespaces"}, "missing-ns")
	})
	missingNamespaceRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/secrets/missing-ns/secret-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, missingNamespaceRec, http.StatusNotFound)
}

func TestKubernetesSecretsAPI_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesSecretsRoutes(t, app, nil)

	adminToken := registerAndLogin(t, app.router, "Secret Kubeconfig Admin", "secret-kubeconfig-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Secret Kubeconfig Project")
	createCluster(t, app.router, adminToken, projectID, "secret-kubeconfig-cluster")
	applicationID := createApplicationForSecrets(t, app.router, adminToken, projectID, "Secret Kubeconfig App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/secrets", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func TestKubernetesSecretsAPI_ClusterUnavailable(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesSecretsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return nil, &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:443: i/o timeout")}
	})

	adminToken := registerAndLogin(t, app.router, "Secret Cluster Admin", "secret-cluster-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Secret Cluster Project")
	createCluster(t, app.router, adminToken, projectID, "secret-cluster-unavailable")
	applicationID := createApplicationForSecrets(t, app.router, adminToken, projectID, "Secret Cluster App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/secrets", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func attachKubernetesSecretsRoutes(t *testing.T, app *rbacTestApp, factory k8ssecrets.ClientsetFactory) {
	t.Helper()

	runtimeSecretService := k8ssecrets.NewSecretService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t))
	if factory != nil {
		runtimeSecretService.WithClientsetFactory(factory)
	}
	handler := handlers.NewKubernetesSecretHandler(runtimeSecretService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/applications/:id/secrets", handler.ListByApplication)
	api.GET("/secrets/:namespace/:name", handler.GetByNamespaceAndName)
}

func createApplicationForSecrets(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
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

func seedSecret(t *testing.T, clientset kubernetes.Interface, applicationID uuid.UUID, organizationID uuid.UUID, namespace, name string) {
	t.Helper()

	immutable := true
	_, err := clientset.CoreV1().Secrets(namespace).Create(t.Context(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
			Annotations: map[string]string{"team": "platform"},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": []byte("s3cr3t"),
			"token":    []byte("abc123"),
		},
		Immutable: &immutable,
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed secret %s/%s: %v", namespace, name, err)
	}
}

func decodeSecretListItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode secret list payload: %v", err)
	}

	return payload.Items
}
