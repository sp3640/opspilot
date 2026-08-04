package integration

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRBACOrganizationLevelIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "RBAC Admin", "rbac-admin@opspilot.dev", "password123")
	userToken := registerAndLogin(t, app.router, "RBAC User", "rbac-user@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "rbac-admin@opspilot.dev")
	member := mustGetUserByEmail(t, app.userRepo, "rbac-user@opspilot.dev")

	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	orgID := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(member.ID, orgID, models.RoleUser); err != nil {
		t.Fatalf("assign user to admin organization: %v", err)
	}

	projectID := createProject(t, app.router, adminToken, "RBAC Project")
	clusterID := createCluster(t, app.router, adminToken, projectID, "rbac-cluster")
	resourceID := createResource(t, app.router, adminToken, projectID, "rbac-resource")
	incidentID := createIncident(t, app.router, adminToken, projectID, "RBAC Incident")
	alertID := createAlert(t, app.router, adminToken, projectID, "RBAC Alert", resourceID)

	now := time.Now().UTC()
	_, err := app.metricService.StoreSnapshot(
		t.Context(),
		projectID,
		admin.ID,
		[]dto.CreateMetricRequest{{
			ProjectID:    projectID,
			ClusterID:    clusterID,
			ResourceID:   resourceID,
			ResourceKind: constants.ResourceKindDeployment,
			MetricType:   constants.MetricTypeCPU,
			MetricName:   "cpu.usage.millicores",
			Value:        42,
			Unit:         "m",
			Timestamp:    now,
		}},
	)
	if err != nil {
		t.Fatalf("seed metric snapshot: %v", err)
	}

	// Admin can access admin endpoints
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/organizations", adminToken, map[string]any{
		"name":        "RBAC Org Extra",
		"description": "admin create",
	}), http.StatusCreated)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/teams", adminToken, map[string]any{
		"name":        "RBAC Team",
		"description": "team",
	}), http.StatusCreated)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
		"email": "invitee@opspilot.dev",
		"role":  models.RoleUser,
	}), http.StatusCreated)

	// User forbidden from admin endpoints
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/organizations/"+orgID.String(), userToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", userToken, map[string]any{
		"email": "blocked@opspilot.dev",
		"role":  models.RoleUser,
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/projects/"+projectID.String(), userToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters", userToken, map[string]any{
		"project_id":           projectID.String(),
		"name":                 "forbidden-cluster",
		"provider":             constants.ClusterProviderKubernetes,
		"status":               constants.ClusterStatusConnected,
		"connection_type":      constants.ClusterConnectionTypeKubeconfig,
		"kubeconfig_encrypted": "dummy",
		"api_endpoint":         "https://api.example",
		"region":               "us-east-1",
		"version":              "1.29",
		"validation_error":     "",
		"metadata":             map[string]any{"env": "test"},
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/clusters/"+clusterID.String(), userToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/resources/"+resourceID.String(), userToken, nil), http.StatusForbidden)

	// Member endpoints accessible
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/organizations", userToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects", userToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/resources", userToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts", userToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/incidents", userToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/metrics", userToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectID.String()+"/audit-logs", userToken, nil), http.StatusOK)

	// Invitation protection
	invList := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/invitations", adminToken, nil)
	assertStatus(t, invList, http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/invitations", userToken, nil), http.StatusForbidden)

	// Team protection (create/delete admin only, list members member accessible)
	team := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/teams", adminToken, map[string]any{
		"name":        "RBAC Team 2",
		"description": "team 2",
	})
	assertStatus(t, team, http.StatusCreated)
	teamData := decodeDataMap(t, team)
	teamID, _ := teamData["id"].(string)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/teams/"+teamID, userToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/teams/"+teamID+"/members", userToken, nil), http.StatusOK)

	// Organization protection
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/organizations/"+orgID.String(), userToken, map[string]any{
		"name":        "Nope",
		"description": "no",
	}), http.StatusForbidden)

	// Cluster protection
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/clusters/"+clusterID.String(), userToken, map[string]any{
		"project_id":           projectID.String(),
		"name":                 "forbidden-update",
		"provider":             constants.ClusterProviderKubernetes,
		"status":               constants.ClusterStatusConnected,
		"connection_type":      constants.ClusterConnectionTypeKubeconfig,
		"kubeconfig_encrypted": "dummy",
		"api_endpoint":         "https://api.example",
		"region":               "us-east-1",
		"version":              "1.29",
		"validation_error":     "",
		"metadata":             map[string]any{"env": "test"},
	}), http.StatusForbidden)

	// Resource protection
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/resources", userToken, map[string]any{
		"project_id":   projectID.String(),
		"kind":         constants.ResourceKindDeployment,
		"name":         "forbidden-resource",
		"display_name": "forbidden-resource",
		"external_id":  "deployment/default/forbidden-resource",
		"provider":     constants.ClusterProviderKubernetes,
		"region":       "us-east-1",
		"namespace":    "default",
		"cluster":      "rbac-cluster",
		"status":       constants.ResourceStatusActive,
		"health":       constants.ResourceHealthHealthy,
		"labels":       map[string]any{"app": "forbidden"},
		"annotations":  map[string]any{},
		"metadata":     map[string]any{"env": "test"},
	}), http.StatusForbidden)

	// Project protection
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/projects/"+projectID.String(), userToken, map[string]any{
		"name":        "forbidden-update",
		"description": "forbidden",
	}), http.StatusForbidden)

	// Audit protection (member access)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/incidents/"+itoa(incidentID)+"/audit-logs", userToken, nil), http.StatusOK)

	// Alert protection (admin write; member ack/resolve)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/alerts/"+itoa(alertID), userToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/acknowledge", userToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/resolve", userToken, nil), http.StatusOK)

	// Incident delete protected (admin only)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/incidents/"+itoa(incidentID), userToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/incidents/"+itoa(incidentID), adminToken, nil), http.StatusOK)
}

type rbacTestApp struct {
	router           *gin.Engine
	cfg              *config.Config
	userRepo         *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	metricService    *services.MetricService
}

func setupRBACApp(t *testing.T) *rbacTestApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:rbac_%d?mode=memory&cache=private", time.Now().UnixNano())
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

	return &rbacTestApp{
		router:           r,
		cfg:              cfg,
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
		metricService:    metricService,
	}
}

func itoa(id uint) string {
	return fmtUint(uint64(id))
}

func fmtUint(v uint64) string {
	return strconv.FormatUint(v, 10)
}
