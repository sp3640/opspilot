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
	k8sruntimedeployments "github.com/sp3640/opspilot/backend/internal/kubernetes/deployments"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestKubernetesRuntimeDeploymentsAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachKubernetesRuntimeDeploymentsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return clientset, nil
	})

	adminAToken := registerAndLogin(t, app.router, "Runtime Deployments Admin A", "runtime-deployments-admin-a@opspilot.dev", "password123")
	memberAToken := registerAndLogin(t, app.router, "Runtime Deployments Member A", "runtime-deployments-member-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Runtime Deployments Admin B", "runtime-deployments-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "runtime-deployments-admin-a@opspilot.dev")
	memberA := mustGetUserByEmail(t, app.userRepo, "runtime-deployments-member-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "runtime-deployments-admin-b@opspilot.dev")
	if adminA.OrganizationID == nil || adminB.OrganizationID == nil {
		t.Fatalf("expected organizations for admins")
	}

	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, organizationB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "runtime-deployments-admin-b@opspilot.dev", "password123")
	if err := app.userRepo.AssignOrganizationAndRole(memberA.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member A role: %v", err)
	}
	memberAToken = loginOnly(t, app.router, "runtime-deployments-member-a@opspilot.dev", "password123")

	projectA := createProject(t, app.router, adminAToken, "Runtime Deployments Project A")
	projectB := createProject(t, app.router, adminBToken, "Runtime Deployments Project B")
	createCluster(t, app.router, adminAToken, projectA, "runtime-deployments-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "runtime-deployments-cluster-b")

	appAID := createApplicationForRuntimeDeployments(t, app.router, adminAToken, projectA, "Runtime Deployments App A")
	appBID := createApplicationForRuntimeDeployments(t, app.router, adminBToken, projectB, "Runtime Deployments App B")

	seedRuntimeDeployment(t, clientset, appAID, organizationA, "default", "dep-healthy", 3, 3, 3, 3, 3, 0, appsv1.DeploymentCondition{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: "MinimumReplicasAvailable"})
	seedRuntimeDeployment(t, clientset, appAID, organizationA, "default", "dep-scaling", 2, 1, 1, 1, 1, 1, appsv1.DeploymentCondition{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: "ReplicaSetUpdated"})
	seedRuntimeDeployment(t, clientset, appAID, organizationA, "default", "dep-failed", 2, 2, 0, 1, 0, 2, appsv1.DeploymentCondition{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionFalse, Reason: "ProgressDeadlineExceeded"})
	seedRuntimeDeployment(t, clientset, appBID, organizationB, "default", "dep-other-org", 1, 1, 1, 1, 1, 0, appsv1.DeploymentCondition{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: "MinimumReplicasAvailable"})

	listAsMemberRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/runtime/deployments", memberAToken, nil)
	assertStatus(t, listAsMemberRec, http.StatusOK)
	listItems := decodeDeploymentRuntimeListItems(t, listAsMemberRec)
	if len(listItems) != 3 {
		t.Fatalf("expected three deployments for app A, got %d", len(listItems))
	}

	listAsAdminRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+appAID.String()+"/runtime/deployments", adminAToken, nil)
	assertStatus(t, listAsAdminRec, http.StatusOK)

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/runtime/deployments/default/dep-healthy?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	deploymentData := decodeDataMap(t, getRec)
	if deploymentData["status"] != k8sruntimedeployments.DeploymentStatusHealthy {
		t.Fatalf("expected healthy status")
	}

	scalingRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/runtime/deployments/default/dep-scaling?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, scalingRec, http.StatusOK)
	scalingData := decodeDataMap(t, scalingRec)
	if scalingData["status"] != k8sruntimedeployments.DeploymentStatusScaling {
		t.Fatalf("expected scaling status")
	}

	failedRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/runtime/deployments/default/dep-failed?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, failedRec, http.StatusOK)
	failedData := decodeDataMap(t, failedRec)
	if failedData["status"] != k8sruntimedeployments.DeploymentStatusFailed {
		t.Fatalf("expected failed status")
	}

	crossOrgRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/runtime/deployments/default/dep-healthy?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, crossOrgRec, http.StatusForbidden)

	missingApplicationQueryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/runtime/deployments/default/dep-healthy", memberAToken, nil)
	assertStatus(t, missingApplicationQueryRec, http.StatusBadRequest)

	deploymentNotFoundRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/runtime/deployments/default/not-found?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, deploymentNotFoundRec, http.StatusNotFound)

	clientset.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
		getAction, ok := action.(ktesting.GetAction)
		if !ok {
			return false, nil, nil
		}
		if getAction.GetNamespace() != "missing-ns" {
			return false, nil, nil
		}

		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespaces"}, "missing-ns")
	})
	missingNamespaceRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/runtime/deployments/missing-ns/dep-healthy?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, missingNamespaceRec, http.StatusNotFound)
}

func TestKubernetesRuntimeDeploymentsAPI_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesRuntimeDeploymentsRoutes(t, app, nil)

	adminToken := registerAndLogin(t, app.router, "Runtime Deployment Kubeconfig Admin", "runtime-deployment-kubeconfig-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Runtime Deployment Kubeconfig Project")
	createCluster(t, app.router, adminToken, projectID, "runtime-deployment-kubeconfig-cluster")
	applicationID := createApplicationForRuntimeDeployments(t, app.router, adminToken, projectID, "Runtime Deployment Kubeconfig App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/runtime/deployments", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func TestKubernetesRuntimeDeploymentsAPI_ClusterUnavailable(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesRuntimeDeploymentsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return nil, &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:443: i/o timeout")}
	})

	adminToken := registerAndLogin(t, app.router, "Runtime Deployment Cluster Admin", "runtime-deployment-cluster-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Runtime Deployment Cluster Project")
	createCluster(t, app.router, adminToken, projectID, "runtime-deployment-cluster-unavailable")
	applicationID := createApplicationForRuntimeDeployments(t, app.router, adminToken, projectID, "Runtime Deployment Cluster App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/runtime/deployments", adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func attachKubernetesRuntimeDeploymentsRoutes(t *testing.T, app *rbacTestApp, factory k8sruntimedeployments.ClientsetFactory) {
	t.Helper()

	runtimeService := k8sruntimedeployments.NewDeploymentRuntimeService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t))
	if factory != nil {
		runtimeService.WithClientsetFactory(factory)
	}
	handler := handlers.NewKubernetesRuntimeDeploymentHandler(runtimeService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/applications/:id/runtime/deployments", handler.ListByApplication)
	api.GET("/runtime/deployments/:namespace/:name", handler.GetByNamespaceAndName)
}

func createApplicationForRuntimeDeployments(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
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

func seedRuntimeDeployment(
	t *testing.T,
	clientset kubernetes.Interface,
	applicationID uuid.UUID,
	organizationID uuid.UUID,
	namespace, name string,
	desiredReplicas, statusReplicas, readyReplicas, updatedReplicas, availableReplicas, unavailableReplicas int32,
	condition appsv1.DeploymentCondition,
) {
	t.Helper()

	strategy := appsv1.RollingUpdateDeploymentStrategyType
	revisionHistoryLimit := int32(10)
	progressDeadline := int32(600)
	minReadySeconds := int32(5)

	_, err := clientset.AppsV1().Deployments(namespace).Create(t.Context(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
				"app":                      name,
			},
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "3",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &desiredReplicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "ghcr.io/opspilot/app:v1"}}},
			},
			Strategy:                appsv1.DeploymentStrategy{Type: strategy},
			RevisionHistoryLimit:    &revisionHistoryLimit,
			ProgressDeadlineSeconds: &progressDeadline,
			MinReadySeconds:         minReadySeconds,
		},
		Status: appsv1.DeploymentStatus{
			Replicas:            statusReplicas,
			ReadyReplicas:       readyReplicas,
			UpdatedReplicas:     updatedReplicas,
			AvailableReplicas:   availableReplicas,
			UnavailableReplicas: unavailableReplicas,
			ObservedGeneration:  1,
			Conditions:          []appsv1.DeploymentCondition{condition},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed runtime deployment %s/%s: %v", namespace, name, err)
	}
}

func decodeDeploymentRuntimeListItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode runtime deployment list payload: %v", err)
	}

	return payload.Items
}
