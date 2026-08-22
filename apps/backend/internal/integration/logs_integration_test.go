package integration

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	k8slogs "github.com/sp3640/opspilot/backend/internal/kubernetes/logs"
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

func TestKubernetesLogsAPIIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachKubernetesLogsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return clientset, nil
	})

	adminAToken := registerAndLogin(t, app.router, "Logs Admin A", "logs-admin-a@opspilot.dev", "password123")
	memberAToken := registerAndLogin(t, app.router, "Logs Member A", "logs-member-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Logs Admin B", "logs-admin-b@opspilot.dev", "password123")

	adminA := mustGetUserByEmail(t, app.userRepo, "logs-admin-a@opspilot.dev")
	memberA := mustGetUserByEmail(t, app.userRepo, "logs-member-a@opspilot.dev")
	adminB := mustGetUserByEmail(t, app.userRepo, "logs-admin-b@opspilot.dev")
	if adminA.OrganizationID == nil || adminB.OrganizationID == nil {
		t.Fatalf("expected organizations for admins")
	}

	organizationA := *adminA.OrganizationID
	organizationB := *adminB.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, organizationB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "logs-admin-b@opspilot.dev", "password123")
	if err := app.userRepo.AssignOrganizationAndRole(memberA.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member A role: %v", err)
	}
	memberAToken = loginOnly(t, app.router, "logs-member-a@opspilot.dev", "password123")

	projectA := createProject(t, app.router, adminAToken, "Logs Project A")
	projectB := createProject(t, app.router, adminBToken, "Logs Project B")
	createCluster(t, app.router, adminAToken, projectA, "logs-cluster-a")
	createCluster(t, app.router, adminBToken, projectB, "logs-cluster-b")

	appAID := createApplicationForLogs(t, app.router, adminAToken, projectA, "Logs App A")
	appBID := createApplicationForLogs(t, app.router, adminBToken, projectB, "Logs App B")

	seedLogPod(t, clientset, appAID, organizationA, "default", "pod-a")
	seedLogPod(t, clientset, appBID, organizationB, "default", "pod-b")

	var logOptionsChecked bool
	clientset.PrependReactor("get", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "log" {
			return false, nil, nil
		}
		genericAction, ok := action.(ktesting.GenericAction)
		if !ok {
			t.Fatalf("expected generic action for pod logs")
		}
		opts, ok := genericAction.GetValue().(*corev1.PodLogOptions)
		if !ok {
			t.Fatalf("expected pod log options in action value")
		}
		if opts.Container != "sidecar" {
			return false, nil, nil
		}
		if opts.TailLines == nil || *opts.TailLines != 100 {
			t.Fatalf("expected tailLines=100")
		}
		if opts.SinceSeconds == nil || *opts.SinceSeconds != 120 {
			t.Fatalf("expected sinceSeconds=120")
		}
		if !opts.Timestamps {
			t.Fatalf("expected timestamps=true")
		}
		logOptionsChecked = true
		return false, nil, nil
	})

	readAsMemberRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-a/logs?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, readAsMemberRec, http.StatusOK)
	memberData := decodeDataMap(t, readAsMemberRec)
	if memberData["pod"] != "pod-a" {
		t.Fatalf("expected pod pod-a")
	}
	if memberData["container"] != "app" {
		t.Fatalf("expected default container app")
	}

	readSpecificContainerRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-a/logs?applicationId="+appAID.String()+"&container=sidecar&tailLines=100&sinceSeconds=120&timestamps=true", adminAToken, nil)
	assertStatus(t, readSpecificContainerRec, http.StatusOK)
	specificContainerData := decodeDataMap(t, readSpecificContainerRec)
	if specificContainerData["container"] != "sidecar" {
		t.Fatalf("expected sidecar container")
	}
	if specificContainerData["namespace"] != "default" {
		t.Fatalf("expected default namespace")
	}
	if _, exists := specificContainerData["retrievedAt"]; !exists {
		t.Fatalf("expected retrievedAt in response")
	}
	if !logOptionsChecked {
		t.Fatalf("expected pod log options reactor to be invoked")
	}

	invalidContainerRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-a/logs?applicationId="+appAID.String()+"&container=does-not-exist", memberAToken, nil)
	assertStatus(t, invalidContainerRec, http.StatusNotFound)

	invalidPodRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-missing/logs?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, invalidPodRec, http.StatusNotFound)

	clientset.PrependReactor("get", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "log" {
			return false, nil, nil
		}
		getAction, ok := action.(ktesting.GetAction)
		if !ok {
			return false, nil, nil
		}
		if getAction.GetNamespace() != "missing-ns" {
			return false, nil, nil
		}

		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespaces"}, "missing-ns")
	})
	missingNamespaceRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/missing-ns/pod-a/logs?applicationId="+appAID.String(), memberAToken, nil)
	assertStatus(t, missingNamespaceRec, http.StatusNotFound)

	organizationIsolationRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-a/logs?applicationId="+appBID.String(), adminBToken, nil)
	assertStatus(t, organizationIsolationRec, http.StatusForbidden)
}

// TestKubernetesLogsAPI_TailLinesDefaultAndClamp verifies Phase 11's
// production hardening: an omitted tailLines defaults to a bounded value
// instead of fetching unbounded log data, and an oversized request is
// clamped rather than honored as-is.
func TestKubernetesLogsAPI_TailLinesDefaultAndClamp(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	clientset := fake.NewSimpleClientset()
	attachKubernetesLogsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return clientset, nil
	})

	adminToken := registerAndLogin(t, app.router, "Log Bounds Admin", "log-bounds-admin@opspilot.dev", "password123")
	adminUser := mustGetUserByEmail(t, app.userRepo, "log-bounds-admin@opspilot.dev")
	organizationID := *adminUser.OrganizationID
	projectID := createProject(t, app.router, adminToken, "Log Bounds Project")
	createCluster(t, app.router, adminToken, projectID, "log-bounds-cluster")
	applicationID := createApplicationForLogs(t, app.router, adminToken, projectID, "Log Bounds App")

	seedLogPod(t, clientset, applicationID, organizationID, "default", "pod-bounds")

	var observedTailLines *int64
	var observedPrevious bool
	clientset.PrependReactor("get", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "log" {
			return false, nil, nil
		}
		genericAction, ok := action.(ktesting.GenericAction)
		if !ok {
			t.Fatalf("expected generic action for pod logs")
		}
		opts, ok := genericAction.GetValue().(*corev1.PodLogOptions)
		if !ok {
			t.Fatalf("expected pod log options in action value")
		}
		observedTailLines = opts.TailLines
		observedPrevious = opts.Previous
		return false, nil, nil
	})

	// No tailLines supplied -> defaulted, not unbounded.
	defaultRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-bounds/logs?applicationId="+applicationID.String(), adminToken, nil)
	assertStatus(t, defaultRec, http.StatusOK)
	if observedTailLines == nil || *observedTailLines != 1000 {
		t.Fatalf("expected default tailLines=1000, got %v", observedTailLines)
	}

	// An oversized request is clamped rather than honored as-is.
	clampedRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-bounds/logs?applicationId="+applicationID.String()+"&tailLines=1000000", adminToken, nil)
	assertStatus(t, clampedRec, http.StatusOK)
	if observedTailLines == nil || *observedTailLines != 5000 {
		t.Fatalf("expected clamped tailLines=5000, got %v", observedTailLines)
	}

	// previous=true is threaded through to the Kubernetes API call and echoed back.
	previousRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-bounds/logs?applicationId="+applicationID.String()+"&previous=true", adminToken, nil)
	assertStatus(t, previousRec, http.StatusOK)
	if !observedPrevious {
		t.Fatalf("expected previous=true to reach PodLogOptions")
	}
	previousData := decodeDataMap(t, previousRec)
	if previousData["previous"] != true {
		t.Fatalf("expected response to echo previous=true, got %v", previousData["previous"])
	}

	invalidPreviousRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/pod-bounds/logs?applicationId="+applicationID.String()+"&previous=not-a-bool", adminToken, nil)
	assertStatus(t, invalidPreviousRec, http.StatusBadRequest)
}

func TestKubernetesLogsAPI_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesLogsRoutes(t, app, nil)

	adminToken := registerAndLogin(t, app.router, "Log Kubeconfig Admin", "log-kubeconfig-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Log Kubeconfig Project")
	createCluster(t, app.router, adminToken, projectID, "log-kubeconfig-cluster")
	applicationID := createApplicationForLogs(t, app.router, adminToken, projectID, "Log Kubeconfig App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/any/logs?applicationId="+applicationID.String(), adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func TestKubernetesLogsAPI_ClusterUnavailable(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	attachKubernetesLogsRoutes(t, app, func(_ []byte) (kubernetes.Interface, error) {
		return nil, &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:443: i/o timeout")}
	})

	adminToken := registerAndLogin(t, app.router, "Log Cluster Admin", "log-cluster-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Log Cluster Project")
	createCluster(t, app.router, adminToken, projectID, "log-cluster-unavailable")
	applicationID := createApplicationForLogs(t, app.router, adminToken, projectID, "Log Cluster App")

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/pods/default/any/logs?applicationId="+applicationID.String(), adminToken, nil)
	assertStatus(t, rec, http.StatusBadGateway)
}

func attachKubernetesLogsRoutes(t *testing.T, app *rbacTestApp, factory k8slogs.ClientsetFactory) {
	t.Helper()

	runtimeLogService := k8slogs.NewLogService(app.applicationRepo, app.clusterRepo, testClusterCredentialCipher(t))
	if factory != nil {
		runtimeLogService.WithClientsetFactory(factory)
	}
	handler := handlers.NewKubernetesLogHandler(runtimeLogService)

	api := app.router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(app.cfg))
	api.GET("/pods/:namespace/:pod/logs", handler.GetPodLogs)
}

func createApplicationForLogs(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
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

func seedLogPod(t *testing.T, clientset kubernetes.Interface, applicationID uuid.UUID, organizationID uuid.UUID, namespace, name string) {
	t.Helper()

	_, err := clientset.CoreV1().Pods(namespace).Create(t.Context(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"opspilot/application-id":  applicationID.String(),
				"opspilot/organization-id": organizationID.String(),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app", Image: "ghcr.io/opspilot/app:v1"},
				{Name: "sidecar", Image: "ghcr.io/opspilot/sidecar:v1"},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("seed pod %s/%s: %v", namespace, name, err)
	}
}
