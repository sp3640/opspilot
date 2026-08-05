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
	k8sservices "github.com/sp3640/opspilot/backend/internal/kubernetes/services"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestKubernetesServicesAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachKubernetesServicesRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return clientset, nil
	})

	adminAToken := registerAndLogin(t, app.router, "Services Admin A", "services-admin-a@opspilot.dev", "password123")
	memberAToken := registerAndLogin(t, app.router, "Services Member A", "services-member-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Services Admin B", "services-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "services-admin-a@opspilot.dev")
	memberA := mustGetUserByEmail(t, app.userRepo, "services-member-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "services-admin-b@opspilot.dev")
	if adminA.OrganizationID == nil || adminB.OrganizationID == nil {
		t.Fatalf("expected organizations for admins")
	}

	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, organizationB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "services-admin-b@opspilot.dev", "password123")
	if err := app.userRepo.AssignOrganizationAndRole(memberA.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member A role: %v", err)
	}
	memberAToken = loginOnly(t, app.router, "services-member-a@opspilot.dev", "password123")

	projectA := createProject(t, app.router, adminAToken, "Services Project A")
	projectB := createProject(t, app.router, adminBToken, "Services Project B")
	createCluster(t, app.router, adminAToken, projectA, "services-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "services-cluster-b")

	appAID := createApplicationForServices(t, app.router, adminAToken, projectA, "Services App A")
	appBID := createApplicationForServices(t, app.router, adminBToken, projectB, "Services App B")

	seedService(t, clientset, appAID, organizationA, "default", "svc-a", corev1.ServiceTypeClusterIP)
	seedService(t, clientset, appBID, organizationB, "default", "svc-b", corev1.ServiceTypeClusterIP)

	listAsMemberRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/services", memberAToken, nil)
	assertStatus(t, listAsMemberRec, http.StatusOK)
	listItems := decodeServiceListItems(t, listAsMemberRec)
	if len(listItems) != 1 {
		t.Fatalf("expected one service for app A, got %d", len(listItems))
	}

	listAsAdminRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/services", adminAToken, nil)
	assertStatus(t, listAsAdminRec, http.StatusOK)

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/services/default/svc-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	serviceData := decodeDataMap(t, getRec)
	if serviceData["name"] != "svc-a" {
		t.Fatalf("expected service name svc-a")
	}

	crossOrgRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/services/default/svc-a?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, crossOrgRec, http.StatusForbidden)

	missingApplicationQueryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/services/default/svc-a", memberAToken, nil)
	assertStatus(t, missingApplicationQueryRec, http.StatusBadRequest)

	invalidApplicationRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/not-a-uuid/services", memberAToken, nil)
	assertStatus(t, invalidApplicationRec, http.StatusBadRequest)

	applicationNotFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+uuid.NewString()+"/services", memberAToken, nil)
	assertStatus(t, applicationNotFoundRec, http.StatusNotFound)

	serviceNotFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/services/default/not-found?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, serviceNotFoundRec, http.StatusNotFound)

	clientset.PrependReactor("get", "services", func(action ktesting.Action) (bool, runtime.Object, error) {
		getAction, ok := action.(ktesting.GetAction)
		if !ok {
			return false, nil, nil
		}
		if getAction.GetNamespace() != "missing-ns" {
			return false, nil, nil
		}

		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespaces"}, "missing-ns")
	})
	missingNamespaceRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/services/missing-ns/svc-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, missingNamespaceRec, http.StatusNotFound)
}

func TestKubernetesServicesAPI_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesServicesRoutes(t, app, nil)

	adminToken := registerAndLogin(t, app.router, "Kubeconfig Admin", "kubeconfig-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Kubeconfig Project")
	createCluster(t, app.router, adminToken, projectID, "kubeconfig-cluster")
	applicationID := createApplicationForServices(t, app.router, adminToken, projectID, "Kubeconfig App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/services", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func TestKubernetesServicesAPI_ClusterUnavailable(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesServicesRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return nil, &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:443: i/o timeout")}
	})

	adminToken := registerAndLogin(t, app.router, "Cluster Admin", "cluster-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Cluster Project")
	createCluster(t, app.router, adminToken, projectID, "cluster-unavailable")
	applicationID := createApplicationForServices(t, app.router, adminToken, projectID, "Cluster App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/services", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func attachKubernetesServicesRoutes(t *testing.T, app *rbacTestApp, factory k8sservices.ClientsetFactory) {
	t.Helper()

	runtimeService := k8sservices.NewServiceService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t))
	if factory != nil {
		runtimeService.WithClientsetFactory(factory)
	}
	handler := handlers.NewKubernetesServiceHandler(runtimeService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/applications/:id/services", handler.ListByApplication)
	api.GET("/services/:namespace/:name", handler.GetByNamespaceAndName)
}

func createApplicationForServices(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
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

func seedService(t *testing.T, clientset kubernetes.Interface, applicationID uuid.UUID, organizationID uuid.UUID, namespace, name string, serviceType corev1.ServiceType) {
	t.Helper()

	_, err := clientset.CoreV1().Services(namespace).Create(t.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
			Annotations: map[string]string{"team": "platform"},
		},
		Spec: corev1.ServiceSpec{
			Type:      serviceType,
			ClusterIP: "10.96.0.10",
			Selector:  map[string]string{"app": name},
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Protocol:   corev1.ProtocolTCP,
			}},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed service %s/%s: %v", namespace, name, err)
	}
}

func decodeServiceListItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode service list payload: %v", err)
	}

	return payload.Items
}
