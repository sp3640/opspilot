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
	k8singresses "github.com/sp3640/opspilot/backend/internal/kubernetes/ingresses"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestKubernetesIngressesAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachKubernetesIngressesRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return clientset, nil
	})

	adminAToken := registerAndLogin(t, app.router, "Ingress Admin A", "ingress-admin-a@opspilot.dev", "password123")
	memberAToken := registerAndLogin(t, app.router, "Ingress Member A", "ingress-member-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Ingress Admin B", "ingress-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "ingress-admin-a@opspilot.dev")
	memberA := mustGetUserByEmail(t, app.userRepo, "ingress-member-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "ingress-admin-b@opspilot.dev")
	if adminA.OrganizationID == nil || adminB.OrganizationID == nil {
		t.Fatalf("expected organizations for admins")
	}

	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, organizationB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "ingress-admin-b@opspilot.dev", "password123")
	if err := app.userRepo.AssignOrganizationAndRole(memberA.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member A role: %v", err)
	}
	memberAToken = loginOnly(t, app.router, "ingress-member-a@opspilot.dev", "password123")

	projectA := createProject(t, app.router, adminAToken, "Ingress Project A")
	projectB := createProject(t, app.router, adminBToken, "Ingress Project B")
	createCluster(t, app.router, adminAToken, projectA, "ingress-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "ingress-cluster-b")

	appAID := createApplicationForIngresses(t, app.router, adminAToken, projectA, "Ingress App A")
	appBID := createApplicationForIngresses(t, app.router, adminBToken, projectB, "Ingress App B")

	seedIngress(t, clientset, appAID, organizationA, "default", "ing-a")
	seedIngress(t, clientset, appBID, organizationB, "default", "ing-b")

	listAsMemberRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/ingresses", memberAToken, nil)
	assertStatus(t, listAsMemberRec, http.StatusOK)
	listItems := decodeIngressListItems(t, listAsMemberRec)
	if len(listItems) != 1 {
		t.Fatalf("expected one ingress for app A, got %d", len(listItems))
	}

	listAsAdminRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/ingresses", adminAToken, nil)
	assertStatus(t, listAsAdminRec, http.StatusOK)

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/ingresses/default/ing-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	ingressData := decodeDataMap(t, getRec)
	if ingressData["name"] != "ing-a" {
		t.Fatalf("expected ingress name ing-a")
	}

	crossOrgRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/ingresses/default/ing-a?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, crossOrgRec, http.StatusForbidden)

	missingApplicationQueryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/ingresses/default/ing-a", memberAToken, nil)
	assertStatus(t, missingApplicationQueryRec, http.StatusBadRequest)

	ingressNotFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/ingresses/default/not-found?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, ingressNotFoundRec, http.StatusNotFound)

	clientset.PrependReactor("get", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
		getAction, ok := action.(ktesting.GetAction)
		if !ok {
			return false, nil, nil
		}
		if getAction.GetNamespace() != "missing-ns" {
			return false, nil, nil
		}

		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespaces"}, "missing-ns")
	})
	missingNamespaceRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/ingresses/missing-ns/ing-a?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, missingNamespaceRec, http.StatusNotFound)
}

func TestKubernetesIngressesAPI_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesIngressesRoutes(t, app, nil)

	adminToken := registerAndLogin(t, app.router, "Ingress Kubeconfig Admin", "ingress-kubeconfig-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Ingress Kubeconfig Project")
	createCluster(t, app.router, adminToken, projectID, "ingress-kubeconfig-cluster")
	applicationID := createApplicationForIngresses(t, app.router, adminToken, projectID, "Ingress Kubeconfig App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/ingresses", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func TestKubernetesIngressesAPI_ClusterUnavailable(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesIngressesRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return nil, &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:443: i/o timeout")}
	})

	adminToken := registerAndLogin(t, app.router, "Ingress Cluster Admin", "ingress-cluster-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Ingress Cluster Project")
	createCluster(t, app.router, adminToken, projectID, "ingress-cluster-unavailable")
	applicationID := createApplicationForIngresses(t, app.router, adminToken, projectID, "Ingress Cluster App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/ingresses", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func attachKubernetesIngressesRoutes(t *testing.T, app *rbacTestApp, factory k8singresses.ClientsetFactory) {
	t.Helper()

	runtimeIngressService := k8singresses.NewIngressService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t))
	if factory != nil {
		runtimeIngressService.WithClientsetFactory(factory)
	}
	handler := handlers.NewKubernetesIngressHandler(runtimeIngressService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/applications/:id/ingresses", handler.ListByApplication)
	api.GET("/ingresses/:namespace/:name", handler.GetByNamespaceAndName)
}

func createApplicationForIngresses(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
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

func seedIngress(t *testing.T, clientset kubernetes.Interface, applicationID uuid.UUID, organizationID uuid.UUID, namespace, name string) {
	t.Helper()

	pathType := networkingv1.PathTypePrefix
	_, err := clientset.NetworkingV1().Ingresses(namespace).Create(t.Context(), &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
			Annotations: map[string]string{"team": "platform"},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: "example.test",
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: "svc-app",
									Port: networkingv1.ServiceBackendPort{Number: 80},
								},
							},
						}},
					},
				},
			}},
		},
		Status: networkingv1.IngressStatus{
			LoadBalancer: networkingv1.IngressLoadBalancerStatus{
				Ingress: []networkingv1.IngressLoadBalancerIngress{{
					Hostname: "lb.example.test",
				}},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed ingress %s/%s: %v", namespace, name, err)
	}
}

func decodeIngressListItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode ingress list payload: %v", err)
	}

	return payload.Items
}
