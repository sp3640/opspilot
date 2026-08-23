package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	k8sconfigmaps "github.com/sp3640/opspilot/backend/internal/kubernetes/configmaps"
	k8snamespaces "github.com/sp3640/opspilot/backend/internal/kubernetes/namespaces"
	k8snodes "github.com/sp3640/opspilot/backend/internal/kubernetes/nodes"
	k8spods "github.com/sp3640/opspilot/backend/internal/kubernetes/pods"
	k8ssecrets "github.com/sp3640/opspilot/backend/internal/kubernetes/secrets"
	"github.com/sp3640/opspilot/backend/internal/metrics"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/router"
	"github.com/sp3640/opspilot/backend/internal/services"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// fakeClusterClientsets keys a set of prebuilt fake Kubernetes clientsets by
// the exact (plaintext) kubeconfig content a test cluster was created with,
// so different clusters in the same test can simulate different live
// states (populated, empty, unreachable) without touching a real cluster.
type fakeClusterClientsets struct {
	byMarker map[string]func() (kubernetes.Interface, error)
}

func newFakeClusterClientsets() *fakeClusterClientsets {
	return &fakeClusterClientsets{byMarker: map[string]func() (kubernetes.Interface, error){}}
}

func (f *fakeClusterClientsets) register(marker string, clientset kubernetes.Interface) {
	f.byMarker[marker] = func() (kubernetes.Interface, error) { return clientset, nil }
}

func (f *fakeClusterClientsets) registerError(marker string, err error) {
	f.byMarker[marker] = func() (kubernetes.Interface, error) { return nil, err }
}

func (f *fakeClusterClientsets) factory(kubeconfig []byte) (kubernetes.Interface, error) {
	marker := string(kubeconfig)
	builder, ok := f.byMarker[marker]
	if !ok {
		return nil, fmt.Errorf("no fake clientset registered for marker %q", marker)
	}

	return builder()
}

type resourceDiscoveryTestApp struct {
	router      *gin.Engine
	cfg         *config.Config
	userRepo    *repository.UserRepository
	clusterRepo *repository.ClusterRepository
	fakes       *fakeClusterClientsets
}

func setupResourceDiscoveryApp(t *testing.T) *resourceDiscoveryTestApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:resource_discovery_%s?mode=memory&cache=private", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}
	migrateIntegrationSchema(t, db)

	cfg := &config.Config{JWTSecret: "this-is-a-very-long-test-jwt-secret-1234567890", RateLimit: 1000}

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	applicationRepo := repository.NewApplicationRepository(db)
	clusterRepo := repository.NewClusterRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	userService := services.NewUserService(userRepo, organizationRepo, invitationRepo, cfg)
	organizationService := services.NewOrganizationService(organizationRepo)
	invitationService := services.NewInvitationService(invitationRepo, userRepo, organizationRepo)
	auditService := services.NewAuditService(auditRepo).WithProjectRepo(projectRepo)
	projectService := services.NewProjectService(projectRepo, userRepo, auditService)
	applicationService := services.NewApplicationService(applicationRepo, projectRepo)
	clusterService := services.NewClusterService(clusterRepo, auditService, testClusterCredentialCipher(t))

	fakes := newFakeClusterClientsets()

	podService := k8spods.NewPodService(applicationRepo, clusterRepo, testClusterCredentialCipher(t)).
		WithClientsetFactory(fakes.factory)
	configMapService := k8sconfigmaps.NewConfigMapService(applicationRepo, clusterRepo, testClusterCredentialCipher(t)).
		WithClientsetFactory(fakes.factory)
	secretService := k8ssecrets.NewSecretService(applicationRepo, clusterRepo, testClusterCredentialCipher(t)).
		WithClientsetFactory(fakes.factory)
	nodeService := k8snodes.NewNodeService(clusterRepo, testClusterCredentialCipher(t)).
		WithClientsetFactory(fakes.factory)
	namespaceService := k8snamespaces.NewNamespaceService(clusterRepo, testClusterCredentialCipher(t)).
		WithClientsetFactory(fakes.factory)

	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)
	projectHandler := handlers.NewProjectHandler(projectService)
	applicationHandler := handlers.NewApplicationHandler(applicationService)
	clusterHandler := handlers.NewClusterHandler(clusterService)
	podHandler := handlers.NewPodHandler(podService)
	configMapHandler := handlers.NewKubernetesConfigMapHandler(configMapService)
	secretHandler := handlers.NewKubernetesSecretHandler(secretService)
	nodeHandler := handlers.NewKubernetesNodeHandler(nodeService)
	namespaceHandler := handlers.NewKubernetesNamespaceHandler(namespaceService)
	healthHandler := handlers.NewHealthHandler(cfg, time.Now(), func(_ context.Context) error { return nil })
	collector := metrics.NewCollector()

	r := gin.New()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(
		middleware.RequestID(),
		middleware.SecurityHeaders(),
		collector.Middleware(),
		middleware.RequestLogger(),
		middleware.RateLimit(middleware.NewIPRateLimiterWithWindow(cfg.RateLimit, time.Minute)),
		middleware.Recovery(collector),
	)

	router.RegisterRoutes(
		r,
		cfg,
		authHandler,
		userHandler,
		organizationHandler,
		invitationHandler,
		projectHandler,
		applicationHandler,
		nil,
		podHandler,
		nil,
		configMapHandler,
		secretHandler,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		clusterHandler,
		nodeHandler,
		namespaceHandler,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		healthHandler,
		collector,
	)

	return &resourceDiscoveryTestApp{
		router:      r,
		cfg:         cfg,
		userRepo:    userRepo,
		clusterRepo: clusterRepo,
		fakes:       fakes,
	}
}

func TestClusterResourceDiscoveryIntegration(t *testing.T) {
	t.Parallel()

	app := setupResourceDiscoveryApp(t)

	adminToken := registerAndLogin(t, app.router, "Discovery Admin", "discovery-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "discovery-admin@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	projectID := createProject(t, app.router, adminToken, "Discovery Project")

	t.Run("resource discovery lists live pods, configmaps, nodes, and namespaces", func(t *testing.T) {
		marker := "populated-cluster"
		app.fakes.register(marker, k8sfake.NewSimpleClientset(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "web-1", Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning},
			},
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "app-config", Namespace: "default"},
				Data:       map[string]string{"key": "value"},
			},
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
					NodeInfo:   corev1.NodeSystemInfo{KubeletVersion: "v1.30.4"},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: "default"},
				Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
			},
		))
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "populated-cluster", marker)

		podsRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/pods", adminToken, nil)
		assertStatus(t, podsRec, http.StatusOK)
		podsData := decodeDataMap(t, podsRec)
		if total, _ := podsData["total"].(float64); total != 1 {
			t.Fatalf("expected 1 pod, got %v", podsData["total"])
		}

		configMapsRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/configmaps", adminToken, nil)
		assertStatus(t, configMapsRec, http.StatusOK)
		configMapsData := decodeDataMap(t, configMapsRec)
		if total, _ := configMapsData["total"].(float64); total != 1 {
			t.Fatalf("expected 1 configmap, got %v", configMapsData["total"])
		}

		nodesRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/nodes", adminToken, nil)
		assertStatus(t, nodesRec, http.StatusOK)
		nodesData := decodeDataMap(t, nodesRec)
		if total, _ := nodesData["total"].(float64); total != 1 {
			t.Fatalf("expected 1 node, got %v", nodesData["total"])
		}

		namespacesRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/namespaces", adminToken, nil)
		assertStatus(t, namespacesRec, http.StatusOK)
		namespacesData := decodeDataMap(t, namespacesRec)
		if total, _ := namespacesData["total"].(float64); total != 1 {
			t.Fatalf("expected 1 namespace, got %v", namespacesData["total"])
		}

		// Discovery timestamp/status: a successful live listing must be
		// reflected on the cluster itself (lastDiscoveryAt populated).
		getClusterRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String(), adminToken, nil)
		assertStatus(t, getClusterRec, http.StatusOK)
		clusterData := decodeDataMap(t, getClusterRec)
		if clusterData["lastDiscoveryAt"] == nil {
			t.Fatalf("expected lastDiscoveryAt to be set after successful resource discovery")
		}

		// Refresh capability: calling the same endpoint again re-queries
		// live state rather than returning a stale cache.
		podsRecAgain := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/pods", adminToken, nil)
		assertStatus(t, podsRecAgain, http.StatusOK)
	})

	t.Run("empty cluster returns an empty, successful list", func(t *testing.T) {
		marker := "empty-cluster"
		app.fakes.register(marker, k8sfake.NewSimpleClientset())
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "empty-cluster", marker)

		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/pods", adminToken, nil)
		assertStatus(t, rec, http.StatusOK)
		data := decodeDataMap(t, rec)
		if total, _ := data["total"].(float64); total != 0 {
			t.Fatalf("expected 0 pods for an empty cluster, got %v", data["total"])
		}
		items, _ := data["items"].([]any)
		if len(items) != 0 {
			t.Fatalf("expected empty items array, got %v", items)
		}
	})

	t.Run("disconnected cluster is handled gracefully", func(t *testing.T) {
		marker := "disconnected-cluster"
		app.fakes.registerError(marker, &intkube.ErrConnectionFailed{Err: errors.New("dial tcp: connection refused")})
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "disconnected-cluster", marker)

		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/pods", adminToken, nil)
		// A connectivity failure is a mapped, well-formed error response —
		// never a panic or an opaque 500.
		assertStatus(t, rec, http.StatusBadGateway)
		env := decodeEnvelope(t, rec)
		if env.Success {
			t.Fatalf("expected success=false for a disconnected cluster")
		}
	})

	t.Run("cross-organization access is forbidden", func(t *testing.T) {
		marker := "cross-org-cluster"
		app.fakes.register(marker, k8sfake.NewSimpleClientset())
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "cross-org-cluster", marker)

		otherToken := registerAndLogin(t, app.router, "Discovery Other Admin", "discovery-other-admin@opspilot.dev", "password123")

		for _, path := range []string{"/pods", "/configmaps", "/secrets", "/nodes", "/namespaces"} {
			rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+path, otherToken, nil)
			assertStatus(t, rec, http.StatusForbidden)
		}

		// Sanity check: the owning org can still read it.
		ownRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/pods", adminToken, nil)
		assertStatus(t, ownRec, http.StatusOK)
	})

	t.Run("secret listing never exposes secret values", func(t *testing.T) {
		marker := "secrets-cluster"
		app.fakes.register(marker, k8sfake.NewSimpleClientset(
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "db-credentials", Namespace: "default"},
				Data: map[string][]byte{
					"username": []byte("admin"),
					"password": []byte("super-secret-value"),
				},
			},
		))
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "secrets-cluster", marker)

		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/secrets", adminToken, nil)
		assertStatus(t, rec, http.StatusOK)

		body := rec.Body.String()
		if containsAny(body, "super-secret-value", "admin") {
			t.Fatalf("secret values leaked into the response body: %s", body)
		}

		data := decodeDataMap(t, rec)
		items, _ := data["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("expected 1 secret, got %d", len(items))
		}
		item, _ := items[0].(map[string]any)
		if _, hasData := item["data"]; hasData {
			t.Fatalf("secret response must never include a data field with values")
		}
		if _, hasStringData := item["stringData"]; hasStringData {
			t.Fatalf("secret response must never include a stringData field with values")
		}
		if keyCount, _ := item["dataKeyCount"].(float64); keyCount != 2 {
			t.Fatalf("expected dataKeyCount=2, got %v", item["dataKeyCount"])
		}
	})

	t.Run("namespace filter is respected", func(t *testing.T) {
		marker := "namespace-filter-cluster"
		app.fakes.register(marker, k8sfake.NewSimpleClientset(
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: "team-a"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}},
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: "team-b"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}},
		))
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "namespace-filter-cluster", marker)

		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/pods?namespace=team-a", adminToken, nil)
		assertStatus(t, rec, http.StatusOK)
		data := decodeDataMap(t, rec)
		if total, _ := data["total"].(float64); total != 1 {
			t.Fatalf("expected 1 pod in team-a, got %v", data["total"])
		}
	})
}

func containsAny(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(haystack, needle) {
			return true
		}
	}

	return false
}
