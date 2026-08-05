package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/auth"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	"github.com/sp3640/opspilot/backend/internal/metrics"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/resourcesync"
	"github.com/sp3640/opspilot/backend/internal/router"
	"github.com/sp3640/opspilot/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type mtAPIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type mtTestApp struct {
	router           *gin.Engine
	cfg              *config.Config
	userRepo         *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	projectRepo      *repository.ProjectRepository
	metricService    *services.MetricService
}

func TestMultiTenantSecurityIsolationIntegration(t *testing.T) {
	t.Parallel()

	app := setupMultiTenantSecurityApp(t)

	// Setup: Org A/User A and Org B/User B, then authenticate both.
	tokenA := registerAndLogin(t, app.router, "Tenant A Owner", "tenant-a@opspilot.dev", "password123")
	userA := mustGetUserByEmail(t, app.userRepo, "tenant-a@opspilot.dev")
	if userA.OrganizationID == nil {
		t.Fatalf("expected user A organization id")
	}
	orgAID := *userA.OrganizationID

	_ = registerAndLogin(t, app.router, "Tenant B Owner", "tenant-b@opspilot.dev", "password123")
	userB := mustGetUserByEmail(t, app.userRepo, "tenant-b@opspilot.dev")

	orgB := &models.Organization{
		Name:        "Tenant B Organization",
		Slug:        "tenant-b-organization",
		Description: "Organization B",
		OwnerID:     userB.ID,
	}
	if err := app.organizationRepo.Create(orgB); err != nil {
		t.Fatalf("create org B: %v", err)
	}
	if err := app.userRepo.AssignOrganizationAndRole(userB.ID, orgB.ID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign org B to user B: %v", err)
	}
	userB = mustGetUserByEmail(t, app.userRepo, "tenant-b@opspilot.dev")
	if userB.OrganizationID == nil {
		t.Fatalf("expected user B organization id")
	}
	orgBID := *userB.OrganizationID
	if orgAID == orgBID {
		t.Fatalf("expected different organizations for user A and user B")
	}
	tokenB := loginOnly(t, app.router, "tenant-b@opspilot.dev", "password123")

	claimsA, err := auth.ValidateToken(tokenA, app.cfg.JWTSecret)
	if err != nil {
		t.Fatalf("validate token A: %v", err)
	}
	claimsB, err := auth.ValidateToken(tokenB, app.cfg.JWTSecret)
	if err != nil {
		t.Fatalf("validate token B: %v", err)
	}
	if claimsA.OrganizationID != orgAID.String() {
		t.Fatalf("token A org claim mismatch: got %s want %s", claimsA.OrganizationID, orgAID.String())
	}
	if claimsB.OrganizationID != orgBID.String() {
		t.Fatalf("token B org claim mismatch: got %s want %s", claimsB.OrganizationID, orgBID.String())
	}

	// Seed tenant A business entities.
	projectAID := createProject(t, app.router, tokenA, "Tenant A Project")
	clusterAID := createCluster(t, app.router, tokenA, projectAID, "tenant-a-cluster")
	resourceAID := createResource(t, app.router, tokenA, projectAID, "tenant-a-resource")
	incidentAID := createIncident(t, app.router, tokenA, projectAID, "Tenant A Incident")
	alertAID := createAlert(t, app.router, tokenA, projectAID, "Tenant A Alert", resourceAID)

	now := time.Now().UTC()
	_, err = app.metricService.StoreSnapshot(
		t.Context(),
		projectAID,
		userA.ID,
		[]dto.CreateMetricRequest{{
			ProjectID:    projectAID,
			ClusterID:    clusterAID,
			ResourceID:   resourceAID,
			ResourceKind: constants.ResourceKindDeployment,
			MetricType:   constants.MetricTypeCPU,
			MetricName:   "cpu.usage.millicores",
			Value:        123,
			Unit:         "m",
			Timestamp:    now,
			Labels:       json.RawMessage(`{"app":"tenant-a"}`),
			Metadata:     json.RawMessage(`{"source":"integration"}`),
		}},
	)
	if err != nil {
		t.Fatalf("store tenant A metric: %v", err)
	}

	// Seed tenant B project so list isolation and JWT-claim behavior are testable.
	projectBID := createProject(t, app.router, tokenB, "Tenant B Project")

	// Projects: B cannot view/update/delete A project, and cannot list it.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectAID.String(), tokenB, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/projects/"+projectAID.String(), tokenB, map[string]any{
		"name":        "Tenant A Project Updated",
		"description": "no",
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/projects/"+projectAID.String(), tokenB, nil), http.StatusForbidden)
	assertListExcludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects", tokenB, nil), projectAID.String())

	// Clusters: repeat assertions.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterAID.String(), tokenB, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/clusters/"+clusterAID.String(), tokenB, map[string]any{
		"project_id":           projectAID.String(),
		"name":                 "tenant-a-cluster-updated",
		"provider":             constants.ClusterProviderKubernetes,
		"status":               constants.ClusterStatusConnected,
		"connection_type":      constants.ClusterConnectionTypeKubeconfig,
		"kubeconfig_encrypted": "dummy",
		"api_endpoint":         "https://api.example",
		"region":               "us-east-1",
		"version":              "1.29",
		"validation_error":     "",
		"metadata":             map[string]any{"env": "a"},
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/clusters/"+clusterAID.String(), tokenB, nil), http.StatusForbidden)
	assertListExcludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters", tokenB, nil), clusterAID.String())

	// Resources: repeat assertions.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/resources/"+resourceAID.String(), tokenB, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/resources/"+resourceAID.String(), tokenB, map[string]any{
		"project_id":   projectAID.String(),
		"kind":         constants.ResourceKindDeployment,
		"name":         "tenant-a-resource-updated",
		"display_name": "tenant-a-resource-updated",
		"external_id":  "deployment/default/tenant-a-resource-updated",
		"provider":     constants.ClusterProviderKubernetes,
		"region":       "us-east-1",
		"namespace":    "default",
		"cluster":      "tenant-a-cluster",
		"status":       constants.ResourceStatusActive,
		"health":       constants.ResourceHealthHealthy,
		"labels":       map[string]any{"app": "tenant-a"},
		"annotations":  map[string]any{},
		"metadata":     map[string]any{"env": "a"},
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/resources/"+resourceAID.String(), tokenB, nil), http.StatusForbidden)
	assertListExcludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/resources", tokenB, nil), resourceAID.String())

	// Incidents: repeat assertions.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/incidents/"+strconv.FormatUint(uint64(incidentAID), 10), tokenB, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/incidents/"+strconv.FormatUint(uint64(incidentAID), 10), tokenB, map[string]any{
		"title":       "Tenant A Incident Updated",
		"description": "no",
		"severity":    constants.SeverityP1,
		"status":      constants.StatusOpen,
		"project_id":  projectAID.String(),
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/incidents/"+strconv.FormatUint(uint64(incidentAID), 10), tokenB, nil), http.StatusForbidden)
	assertListExcludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/incidents", tokenB, nil), strconv.FormatUint(uint64(incidentAID), 10))

	// Alerts: repeat assertions.
	alertAIDStr := strconv.FormatUint(uint64(alertAID), 10)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts/"+alertAIDStr, tokenB, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/alerts/"+alertAIDStr, tokenB, map[string]any{
		"project_id":       projectAID.String(),
		"title":            "Tenant A Alert Updated",
		"description":      "no",
		"severity":         constants.AlertSeverityHigh,
		"status":           constants.AlertStatusOpen,
		"source":           constants.AlertSourceKubernetes,
		"resource_type":    constants.AlertResourceTypeDeployment,
		"resource_id":      resourceAID.String(),
		"fingerprint":      "unused-by-service",
		"occurrence_count": 1,
		"labels":           map[string]any{"app": "tenant-a"},
		"metadata":         map[string]any{"env": "a"},
		"first_seen_at":    now.Format(time.RFC3339),
		"last_seen_at":     now.Format(time.RFC3339),
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/alerts/"+alertAIDStr, tokenB, nil), http.StatusForbidden)
	assertListExcludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts", tokenB, nil), alertAIDStr)

	// Metrics: B cannot view or list A metrics.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet,
		"/api/v1/metrics/latest?projectId="+projectAID.String()+"&clusterId="+clusterAID.String()+"&resourceId="+resourceAID.String()+"&metricType="+constants.MetricTypeCPU+"&metricName=cpu.usage.millicores",
		tokenB,
		nil,
	), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet,
		"/api/v1/metrics/history?projectId="+projectAID.String()+"&clusterId="+clusterAID.String()+"&resourceId="+resourceAID.String()+"&metricType="+constants.MetricTypeCPU+"&metricName=cpu.usage.millicores",
		tokenB,
		nil,
	), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet,
		"/api/v1/metrics/aggregate?projectId="+projectAID.String()+"&metricType="+constants.MetricTypeCPU+"&metricName=cpu.usage.millicores&interval=hour&start="+now.Add(-time.Hour).Format(time.RFC3339)+"&end="+now.Add(time.Hour).Format(time.RFC3339),
		tokenB,
		nil,
	), http.StatusForbidden)
	assertListExcludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/metrics", tokenB, nil), projectAID.String())

	// Audit logs: each tenant only sees their own; cross-tenant access forbidden.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectAID.String()+"/audit-logs", tokenA, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectAID.String()+"/audit-logs", tokenB, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/incidents/"+strconv.FormatUint(uint64(incidentAID), 10)+"/audit-logs", tokenB, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectBID.String()+"/audit-logs", tokenB, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectBID.String()+"/audit-logs", tokenA, nil), http.StatusForbidden)

	// JWT claim is respected: same user identity with org B claim cannot access org A data.
	forgedToken, err := auth.GenerateToken(userA.ID, userA.Email, string(userA.Role), orgBID.String(), app.cfg.JWTSecret)
	if err != nil {
		t.Fatalf("forge token: %v", err)
	}
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectAID.String(), forgedToken, nil), http.StatusForbidden)
	assertListIncludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects", forgedToken, nil), projectBID.String())
	assertListExcludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects", forgedToken, nil), projectAID.String())
}

func setupMultiTenantSecurityApp(t *testing.T) *mtTestApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:multi_tenant_security_%s?mode=memory&cache=private", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}

	migrateIntegrationSchema(t, db)

	cfg := &config.Config{
		JWTSecret: "this-is-a-very-long-test-jwt-secret-1234567890",
		RateLimit: 1000,
	}

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	applicationRepo := repository.NewApplicationRepository(db)
	deploymentRepo := repository.NewDeploymentRepository(db)
	deploymentHistoryRepo := repository.NewDeploymentHistoryRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	projectTeamRepo := repository.NewProjectTeamRepository(db)
	teamMemberRepo := repository.NewTeamMemberRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	clusterRepo := repository.NewClusterRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	metricRepo := repository.NewMetricRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)

	userService := services.NewUserService(userRepo, organizationRepo, cfg)
	organizationService := services.NewOrganizationService(organizationRepo)
	invitationService := services.NewInvitationService(invitationRepo, userRepo)
	auditService := services.NewAuditService(auditRepo).WithProjectRepo(projectRepo).WithIncidentRepo(incidentRepo)
	projectService := services.NewProjectService(projectRepo, userRepo, auditService)
	applicationService := services.NewApplicationService(applicationRepo, projectRepo)
	deploymentHistoryService := services.NewDeploymentHistoryService(deploymentHistoryRepo, deploymentRepo)
	deploymentService := services.NewDeploymentService(deploymentRepo, applicationRepo, projectRepo, clusterRepo, deploymentHistoryService)
	teamService := services.NewTeamService(teamRepo, teamMemberRepo, userRepo)
	projectTeamService := services.NewProjectTeamService(projectTeamRepo, projectRepo, teamRepo)
	incidentService := services.NewIncidentService(incidentRepo, commentRepo, auditRepo, auditService)
	alertService := services.NewAlertService(alertRepo, incidentRepo, auditService)
	clusterService := services.NewClusterService(clusterRepo, auditService, testClusterCredentialCipher(t))
	resourceService := services.NewResourceService(resourceRepo, resourcesync.NewSyncEngine(resourceRepo), auditService)
	metricService := services.NewMetricService(metricRepo, auditService)
	commentService := services.NewCommentService(commentRepo, incidentRepo, auditService)
	dashboardService := services.NewDashboardService(dashboardRepo)

	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)
	projectHandler := handlers.NewProjectHandler(projectService)
	applicationHandler := handlers.NewApplicationHandler(applicationService)
	deploymentHandler := handlers.NewDeploymentHandler(deploymentService)
	deploymentHistoryHandler := handlers.NewDeploymentHistoryHandler(deploymentHistoryService)
	teamHandler := handlers.NewTeamHandler(teamService)
	projectTeamHandler := handlers.NewProjectTeamHandler(projectTeamService)
	incidentHandler := handlers.NewIncidentHandler(incidentService)
	alertHandler := handlers.NewAlertHandler(alertService)
	metricHandler := handlers.NewMetricHandler(metricService)
	clusterHandler := handlers.NewClusterHandler(clusterService)
	resourceHandler := handlers.NewResourceHandler(resourceService, clusterService, nil)
	commentHandler := handlers.NewCommentHandler(commentService)
	auditHandler := handlers.NewAuditHandler(auditService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
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
		deploymentHandler,
		deploymentHistoryHandler,
		teamHandler,
		projectTeamHandler,
		incidentHandler,
		alertHandler,
		metricHandler,
		clusterHandler,
		resourceHandler,
		commentHandler,
		auditHandler,
		dashboardHandler,
		healthHandler,
		collector,
	)

	return &mtTestApp{
		router:           r,
		cfg:              cfg,
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
		projectRepo:      projectRepo,
		metricService:    metricService,
	}
}

func registerAndLogin(t *testing.T, r *gin.Engine, name, email, password string) string {
	t.Helper()
	registerRec := doJSONRequest(t, r, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"name":     name,
		"email":    email,
		"password": password,
	})
	assertStatus(t, registerRec, http.StatusCreated)

	return loginOnly(t, r, email, password)
}

func loginOnly(t *testing.T, r *gin.Engine, email, password string) string {
	t.Helper()
	loginRec := doJSONRequest(t, r, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	assertStatus(t, loginRec, http.StatusOK)
	env := decodeEnvelope(t, loginRec)
	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode login data: %v", err)
	}
	token, _ := data["access_token"].(string)
	if token == "" {
		t.Fatalf("missing access token in login response")
	}
	return token
}

func mustGetUserByEmail(t *testing.T, repo *repository.UserRepository, email string) *models.User {
	t.Helper()
	user, err := repo.GetByEmail(email)
	if err != nil {
		t.Fatalf("load user by email %s: %v", email, err)
	}
	return user
}

func createProject(t *testing.T, r *gin.Engine, token, name string) uuid.UUID {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/projects", token, map[string]any{
		"name":        name,
		"description": "integration",
	})
	assertStatus(t, rec, http.StatusCreated)
	data := decodeDataMap(t, rec)
	id, _ := data["id"].(string)
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("parse project id: %v", err)
	}
	return parsed
}

func createCluster(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/clusters", token, map[string]any{
		"project_id":           projectID.String(),
		"name":                 name,
		"provider":             constants.ClusterProviderKubernetes,
		"status":               constants.ClusterStatusConnected,
		"connection_type":      constants.ClusterConnectionTypeKubeconfig,
		"kubeconfig_encrypted": "dummy",
		"api_endpoint":         "https://api.example",
		"region":               "us-east-1",
		"version":              "1.29",
		"validation_error":     "",
		"metadata":             map[string]any{"env": "test"},
	})
	assertStatus(t, rec, http.StatusCreated)
	data := decodeDataMap(t, rec)
	id, _ := data["id"].(string)
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("parse cluster id: %v", err)
	}
	return parsed
}

func createResource(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/resources", token, map[string]any{
		"project_id":   projectID.String(),
		"kind":         constants.ResourceKindDeployment,
		"name":         name,
		"display_name": name,
		"external_id":  "deployment/default/" + name,
		"provider":     constants.ClusterProviderKubernetes,
		"region":       "us-east-1",
		"namespace":    "default",
		"cluster":      "tenant-a-cluster",
		"status":       constants.ResourceStatusActive,
		"health":       constants.ResourceHealthHealthy,
		"labels":       map[string]any{"app": name},
		"annotations":  map[string]any{},
		"metadata":     map[string]any{"env": "test"},
	})
	assertStatus(t, rec, http.StatusCreated)
	data := decodeDataMap(t, rec)
	id, _ := data["id"].(string)
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("parse resource id: %v", err)
	}
	return parsed
}

func createIncident(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, title string) uint {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/incidents", token, map[string]any{
		"title":       title,
		"description": "integration",
		"severity":    constants.SeverityP1,
		"status":      constants.StatusOpen,
		"project_id":  projectID.String(),
	})
	assertStatus(t, rec, http.StatusCreated)
	data := decodeDataMap(t, rec)
	return toUint(t, data["id"])
}

func createAlert(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, title string, resourceID uuid.UUID) uint {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/alerts", token, map[string]any{
		"project_id":       projectID.String(),
		"title":            title,
		"description":      "integration",
		"severity":         constants.AlertSeverityHigh,
		"status":           constants.AlertStatusOpen,
		"source":           constants.AlertSourceKubernetes,
		"resource_type":    constants.AlertResourceTypeDeployment,
		"resource_id":      resourceID.String(),
		"fingerprint":      "unused-by-service",
		"occurrence_count": 1,
		"labels":           map[string]any{"app": "tenant-a"},
		"metadata":         map[string]any{"env": "test"},
		"first_seen_at":    now,
		"last_seen_at":     now,
	})
	assertStatus(t, rec, http.StatusCreated)
	data := decodeDataMap(t, rec)
	return toUint(t, data["id"])
}

func doJSONRequest(t *testing.T, r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		payload = encoded
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if rec.Code != expected {
		t.Fatalf("unexpected status code: got %d want %d; body=%s", rec.Code, expected, rec.Body.String())
	}
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) mtAPIResponse {
	t.Helper()
	var env mtAPIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response envelope: %v; body=%s", err, rec.Body.String())
	}
	return env
}

func decodeDataMap(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode response data map: %v; data=%s", err, string(env.Data))
	}
	return data
}

func assertListExcludesID(t *testing.T, rec *httptest.ResponseRecorder, disallowedID string) {
	t.Helper()
	assertStatus(t, rec, http.StatusOK)
	env := decodeEnvelope(t, rec)
	var data struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode list payload: %v; data=%s", err, string(env.Data))
	}
	for _, item := range data.Items {
		if id, ok := item["id"]; ok && fmt.Sprint(id) == disallowedID {
			t.Fatalf("list leaked disallowed id %s; body=%s", disallowedID, rec.Body.String())
		}
	}
}

func assertListIncludesID(t *testing.T, rec *httptest.ResponseRecorder, expectedID string) {
	t.Helper()
	assertStatus(t, rec, http.StatusOK)
	env := decodeEnvelope(t, rec)
	var data struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode list payload: %v; data=%s", err, string(env.Data))
	}
	for _, item := range data.Items {
		if id, ok := item["id"]; ok && fmt.Sprint(id) == expectedID {
			return
		}
	}
	t.Fatalf("expected list to include id %s; body=%s", expectedID, rec.Body.String())
}

func toUint(t *testing.T, value any) uint {
	t.Helper()
	switch v := value.(type) {
	case float64:
		return uint(v)
	case int:
		return uint(v)
	case int64:
		return uint(v)
	case uint:
		return v
	default:
		t.Fatalf("unsupported numeric type for uint conversion: %T", value)
		return 0
	}
}
