package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/auth"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/discovery"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/monitoring"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/resourcesync"
	"github.com/sp3640/opspilot/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type staticDiscoveryProvider struct {
	name      string
	provider  monitoring.ProviderType
	resources []models.Resource
}

func (p *staticDiscoveryProvider) Validate(context.Context) error {
	return nil
}

func (p *staticDiscoveryProvider) Discover(context.Context) (*discovery.DiscoveryResult, error) {
	return &discovery.DiscoveryResult{
		Resources: p.resources,
		Warnings:  []string{},
		Errors:    []string{},
		Duration:  10 * time.Millisecond,
	}, nil
}

func (p *staticDiscoveryProvider) Name() string {
	return p.name
}

func (p *staticDiscoveryProvider) Provider() monitoring.ProviderType {
	return p.provider
}

func TestOpsPilotCompleteWorkflowIntegration(t *testing.T) {
	t.Parallel()

	db := setupSQLiteIntegrationDB(t)
	runCompleteWorkflowIntegration(t, db)
}

func runCompleteWorkflowIntegration(t *testing.T, db *gorm.DB) {
	t.Helper()

	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	clusterRepo := repository.NewClusterRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	metricRepo := repository.NewMetricRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)

	testConfig := &config.Config{JWTSecret: "this-is-a-very-long-test-jwt-secret-1234567890", ClusterCredentialEncryptionKey: testClusterEncryptionKey()}

	userService := services.NewUserService(userRepo, organizationRepo, invitationRepo, testConfig)
	auditService := services.NewAuditService(auditRepo).
		WithProjectRepo(projectRepo).
		WithIncidentRepo(incidentRepo)
	projectService := services.NewProjectService(projectRepo, userRepo, auditService)
	incidentService := services.NewIncidentService(incidentRepo, commentRepo, auditRepo, auditService)
	alertService := services.NewAlertService(alertRepo, incidentRepo, auditService)
	clusterService := services.NewClusterService(clusterRepo, auditService, testClusterCredentialCipher(t))
	resourceService := services.NewResourceService(resourceRepo, resourcesync.NewSyncEngine(resourceRepo), auditService)
	metricService := services.NewMetricService(metricRepo, auditService)
	dashboardService := services.NewDashboardService(dashboardRepo)

	// 1) User authentication
	if err := userService.Register("Ops Owner", "owner@opspilot.dev", "password123", ""); err != nil {
		t.Fatalf("register user: %v", err)
	}
	token, err := userService.Login("owner@opspilot.dev", "password123")
	if err != nil {
		t.Fatalf("login user: %v", err)
	}
	claims, err := auth.ValidateToken(token, testConfig.JWTSecret)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	owner, err := userRepo.GetByEmail("owner@opspilot.dev")
	if err != nil {
		t.Fatalf("load owner by email: %v", err)
	}
	if claims.UserID != owner.ID {
		t.Fatalf("unexpected token user id: got %d want %d", claims.UserID, owner.ID)
	}
	if owner.OrganizationID == nil {
		t.Fatalf("expected owner organization id")
	}
	organizationID := *owner.OrganizationID

	// 2) Project creation
	projectResp, err := projectService.Create(ctx, "OpsPilot Integration", "Workflow integration project", owner.ID, organizationID)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	projectID, err := uuid.Parse(projectResp.ID)
	if err != nil {
		t.Fatalf("parse project id: %v", err)
	}

	// 3) Cluster registration
	clusterResp, err := clusterService.CreateCluster(
		ctx,
		projectID,
		"primary-cluster",
		constants.ClusterProviderKubernetes,
		constants.ClusterConnectionTypeKubeconfig,
		"not-a-valid-kubeconfig",
		"",
		"us-east-1",
		json.RawMessage(`{"environment":"test"}`),
		owner.ID,
		organizationID,
	)
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	clusterID, err := uuid.Parse(clusterResp.ID)
	if err != nil {
		t.Fatalf("parse cluster id: %v", err)
	}

	// 4) Cluster validation — "not-a-valid-kubeconfig" is not real kubeconfig
	// YAML, so real connectivity validation correctly reports it as
	// unreachable/invalid rather than healthy; the workflow under test here is
	// that validation runs and persists an outcome, not that this particular
	// placeholder credential is connectable.
	validationResult, err := clusterService.ValidateClusterCredential(ctx, clusterID, organizationID)
	if err != nil {
		t.Fatalf("validate cluster credential: %v", err)
	}
	if validationResult.Connected {
		t.Fatalf("expected placeholder kubeconfig to fail real validation: %+v", validationResult)
	}
	validatedCluster := validationResult.Cluster
	if validatedCluster == nil || validatedCluster.LastValidatedAt == nil {
		t.Fatalf("expected last validated at to be set")
	}
	if validatedCluster.Status != constants.ClusterStatusInvalid {
		t.Fatalf("expected cluster status INVALID after failed validation, got %s", validatedCluster.Status)
	}

	// 5) Resource discovery
	discoveredResource := models.Resource{
		ProjectID:   projectID,
		Kind:        constants.ResourceKindDeployment,
		Name:        "ops-api",
		DisplayName: "ops-api",
		ExternalID:  "deployment/default/ops-api",
		Provider:    constants.ClusterProviderKubernetes,
		Region:      "us-east-1",
		Namespace:   "default",
		Cluster:     validatedCluster.Name,
		Status:      constants.ResourceStatusActive,
		Health:      constants.ResourceHealthHealthy,
		Labels:      json.RawMessage(`{"app":"ops-api"}`),
		Annotations: json.RawMessage(`{}`),
		Metadata:    json.RawMessage(`{"discoveredBy":"integration-test"}`),
		CreatedBy:   owner.ID,
	}

	provider := &staticDiscoveryProvider{
		name:      validatedCluster.Name,
		provider:  monitoring.ProviderKubernetes,
		resources: []models.Resource{discoveredResource},
	}

	discoveryManager := discovery.NewDiscoveryManager(nil)
	discoveryManager.RegisterProvider(provider, projectID, clusterID)

	execution, err := discoveryManager.ExecuteDiscovery(ctx, provider.Provider(), provider.Name(), projectID, clusterID)
	if err != nil {
		t.Fatalf("execute discovery: %v", err)
	}
	if execution == nil || execution.Result == nil || len(execution.Result.Resources) == 0 {
		t.Fatalf("expected discovered resources")
	}

	// 6) Resource synchronization
	syncResult, err := resourceService.SyncResources(ctx, projectID, owner.ID, organizationID, execution.Result.Resources)
	if err != nil {
		t.Fatalf("sync resources: %v", err)
	}
	if syncResult.Created < 1 {
		t.Fatalf("expected at least one created resource, got %d", syncResult.Created)
	}

	syncedResources, err := resourceRepo.ListByProject(projectID, organizationID)
	if err != nil {
		t.Fatalf("list synced resources: %v", err)
	}
	if len(syncedResources) == 0 {
		t.Fatalf("expected synced resources in storage")
	}
	primaryResource := syncedResources[0]

	// 7) Metrics collection
	metricSnapshots := []dto.CreateMetricRequest{
		{
			ProjectID:    projectID,
			ClusterID:    clusterID,
			ResourceID:   primaryResource.ID,
			ResourceKind: primaryResource.Kind,
			MetricType:   constants.MetricTypeCPU,
			MetricName:   "cpu.usage.millicores",
			Value:        120,
			Unit:         "m",
			Timestamp:    time.Now().UTC(),
			Labels:       json.RawMessage(`{"container":"api"}`),
			Metadata:     json.RawMessage(`{"sample":"integration"}`),
		},
	}

	storedMetrics, err := metricService.StoreSnapshot(ctx, projectID, owner.ID, metricSnapshots)
	if err != nil {
		t.Fatalf("store metrics snapshot: %v", err)
	}
	if len(storedMetrics) != 1 {
		t.Fatalf("expected one stored metric, got %d", len(storedMetrics))
	}

	latestMetric, err := metricService.GetLatest(organizationID, projectID, &clusterID, &primaryResource.ID, constants.MetricTypeCPU, "cpu.usage.millicores")
	if err != nil {
		t.Fatalf("get latest metric: %v", err)
	}
	if latestMetric == nil {
		t.Fatalf("expected latest metric response")
	}

	// 8) Alert creation
	now := time.Now().UTC()
	alertResp, err := alertService.CreateAlert(
		ctx,
		projectID,
		nil,
		"CPU saturation",
		"CPU usage exceeded threshold",
		constants.AlertSeverityHigh,
		constants.AlertStatusOpen,
		constants.AlertSourceKubernetes,
		constants.AlertResourceTypeDeployment,
		primaryResource.ID.String(),
		json.RawMessage(`{"resource":"ops-api"}`),
		json.RawMessage(`{"threshold":90}`),
		&now,
		&now,
		owner.ID,
		organizationID,
	)
	if err != nil {
		t.Fatalf("create alert: %v", err)
	}

	// 9) Alert acknowledgement
	acknowledgedAlert, err := alertService.AcknowledgeAlert(ctx, alertResp.ID, owner.ID, organizationID)
	if err != nil {
		t.Fatalf("acknowledge alert: %v", err)
	}
	if acknowledgedAlert.Status != constants.AlertStatusAcknowledged {
		t.Fatalf("unexpected alert status after acknowledgement: %s", acknowledgedAlert.Status)
	}

	// 10) Incident creation
	incidentResp, err := incidentService.CreateIncident(
		ctx,
		"High CPU incident",
		"Automated incident for sustained CPU saturation",
		constants.SeverityP1,
		constants.StatusOpen,
		projectID,
		owner.ID,
		organizationID,
	)
	if err != nil {
		t.Fatalf("create incident: %v", err)
	}

	// 11) Alert -> Incident linking
	linkedAlert, err := alertService.AttachIncident(ctx, alertResp.ID, incidentResp.ID, owner.ID, organizationID)
	if err != nil {
		t.Fatalf("attach alert to incident: %v", err)
	}
	if linkedAlert.IncidentID == nil || *linkedAlert.IncidentID != incidentResp.ID {
		t.Fatalf("expected alert to reference incident %d", incidentResp.ID)
	}

	// 12) Alert resolution
	resolvedAlert, err := alertService.ResolveAlert(ctx, alertResp.ID, owner.ID, organizationID)
	if err != nil {
		t.Fatalf("resolve alert: %v", err)
	}
	if resolvedAlert.Status != constants.AlertStatusResolved {
		t.Fatalf("unexpected alert status after resolve: %s", resolvedAlert.Status)
	}

	// 13) Dashboard aggregation
	summary, err := dashboardService.GetSummary(owner.ID)
	if err != nil {
		t.Fatalf("dashboard summary: %v", err)
	}
	if summary.TotalProjects < 1 {
		t.Fatalf("expected summary total projects >= 1, got %d", summary.TotalProjects)
	}

	stats, err := dashboardService.GetStats(owner.ID)
	if err != nil {
		t.Fatalf("dashboard stats: %v", err)
	}
	if stats.Incidents.Total < 1 {
		t.Fatalf("expected incident stats total >= 1, got %d", stats.Incidents.Total)
	}

	// 14) Audit log generation
	auditLogs, err := auditRepo.GetByProjectID(projectID, organizationID)
	if err != nil {
		t.Fatalf("list project audit logs: %v", err)
	}
	if len(auditLogs) == 0 {
		t.Fatalf("expected audit logs to be generated")
	}

	var hasProjectCreate bool
	var hasClusterUpdate bool
	var hasResourceCreate bool
	var hasMetricSnapshot bool
	var hasAlertCreate bool
	var hasIncidentCreate bool

	for _, log := range auditLogs {
		switch {
		case log.EntityType == "project" && log.Action == models.AuditActionCreate:
			hasProjectCreate = true
		case log.EntityType == "cluster" && log.Action == models.AuditActionUpdate:
			hasClusterUpdate = true
		case log.EntityType == "resource" && log.Action == models.AuditActionCreate:
			hasResourceCreate = true
		case log.EntityType == "metric_snapshot" && log.Action == models.AuditActionCreate:
			hasMetricSnapshot = true
		case log.EntityType == "alert" && log.Action == models.AuditActionCreate:
			hasAlertCreate = true
		case log.EntityType == "incident" && log.Action == models.AuditActionCreate:
			hasIncidentCreate = true
		}
	}

	if !hasProjectCreate {
		t.Fatalf("expected project create audit log")
	}
	if !hasClusterUpdate {
		t.Fatalf("expected cluster update audit log")
	}
	if !hasResourceCreate {
		t.Fatalf("expected resource create audit log")
	}
	if !hasMetricSnapshot {
		t.Fatalf("expected metric snapshot audit log")
	}
	if !hasAlertCreate {
		t.Fatalf("expected alert create audit log")
	}
	if !hasIncidentCreate {
		t.Fatalf("expected incident create audit log")
	}
}

func setupSQLiteIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:opspilot_workflow?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}

	migrateIntegrationSchema(t, db)

	return db
}

func migrateIntegrationSchema(t *testing.T, db *gorm.DB) {
	t.Helper()

	err := db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.Invitation{},
		&models.Project{},
		&models.Application{},
		&models.Deployment{},
		&models.DeploymentHistory{},
		&models.Team{},
		&models.ProjectTeam{},
		&models.ApplicationTeam{},
		&models.TeamMember{},
		&models.Incident{},
		&models.Alert{},
		&models.Cluster{},
		&models.Resource{},
		&models.Metric{},
		&models.Comment{},
		&models.AuditLog{},
	)
	if err != nil {
		t.Fatalf("auto migrate schema: %v", err)
	}
}
