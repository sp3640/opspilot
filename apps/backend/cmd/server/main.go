package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/bootstrap"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/connector"
	githubconnector "github.com/sp3640/opspilot/backend/internal/connector/github"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/database"
	"github.com/sp3640/opspilot/backend/internal/discovery"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	k8sconfigmaps "github.com/sp3640/opspilot/backend/internal/kubernetes/configmaps"
	k8sruntimedeployments "github.com/sp3640/opspilot/backend/internal/kubernetes/deployments"
	k8sevents "github.com/sp3640/opspilot/backend/internal/kubernetes/events"
	k8sexecutor "github.com/sp3640/opspilot/backend/internal/kubernetes/executor"
	k8singresses "github.com/sp3640/opspilot/backend/internal/kubernetes/ingresses"
	k8slogs "github.com/sp3640/opspilot/backend/internal/kubernetes/logs"
	k8snamespaces "github.com/sp3640/opspilot/backend/internal/kubernetes/namespaces"
	k8snodes "github.com/sp3640/opspilot/backend/internal/kubernetes/nodes"
	k8spods "github.com/sp3640/opspilot/backend/internal/kubernetes/pods"
	k8sreplicasets "github.com/sp3640/opspilot/backend/internal/kubernetes/replicasets"
	k8ssecrets "github.com/sp3640/opspilot/backend/internal/kubernetes/secrets"
	k8sservices "github.com/sp3640/opspilot/backend/internal/kubernetes/services"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/metrics"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/notification"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/resourcesync"
	"github.com/sp3640/opspilot/backend/internal/router"
	"github.com/sp3640/opspilot/backend/internal/security"
	"github.com/sp3640/opspilot/backend/internal/services"
)

const shutdownTimeout = 30 * time.Second

type discoveryClusterLoader struct {
	db   *gorm.DB
	repo *repository.ClusterRepository
}

func (l *discoveryClusterLoader) LoadCluster(_ context.Context, clusterID uuid.UUID) (*discovery.ClusterDescriptor, error) {
	if l == nil || l.db == nil {
		return nil, nil
	}

	var cluster models.Cluster
	if err := l.db.Where("id = ?", clusterID).First(&cluster).Error; err != nil {
		return nil, err
	}

	return &discovery.ClusterDescriptor{
		ID:                  cluster.ID,
		OrganizationID:      cluster.OrganizationID,
		ProjectID:           cluster.ProjectID,
		Name:                cluster.Name,
		Provider:            cluster.Provider,
		Status:              cluster.Status,
		KubeconfigEncrypted: []byte(cluster.KubeconfigEncrypted),
		CreatedBy:           cluster.CreatedBy,
	}, nil
}

func (l *discoveryClusterLoader) LoadEnabledClusters(_ context.Context) ([]bootstrap.RuntimeCluster, error) {
	if l == nil || l.db == nil {
		return []bootstrap.RuntimeCluster{}, nil
	}

	clusters := make([]models.Cluster, 0)
	if err := l.db.Where("deleted_at IS NULL").Find(&clusters).Error; err != nil {
		return nil, err
	}

	runtimeClusters := make([]bootstrap.RuntimeCluster, 0, len(clusters))
	for _, cluster := range clusters {
		runtimeClusters = append(runtimeClusters, bootstrap.RuntimeCluster{
			ID:                  cluster.ID,
			ProjectID:           cluster.ProjectID,
			Name:                cluster.Name,
			Provider:            cluster.Provider,
			Status:              cluster.Status,
			KubeconfigEncrypted: []byte(cluster.KubeconfigEncrypted),
			CreatedBy:           cluster.CreatedBy,
		})
	}

	return runtimeClusters, nil
}

func main() {
	logger.Configure()
	if err := run(); err != nil {
		logger.Error(context.Background(), "application terminated", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	processContext := context.Background()
	startedAt := time.Now()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if _, err := database.Connect(cfg); err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer func() {
		logger.Info(processContext, "closing PostgreSQL connection pool")
		if err := database.Close(); err != nil {
			logger.Error(processContext, "failed to close PostgreSQL connection pool", slog.Any("error", err))
			return
		}
		logger.Info(processContext, "PostgreSQL connection pool closed")
	}()

	userRepo := repository.NewUserRepository(database.DB)
	organizationRepo := repository.NewOrganizationRepository(database.DB)
	invitationRepo := repository.NewInvitationRepository(database.DB)
	projectRepo := repository.NewProjectRepository(database.DB)
	applicationRepo := repository.NewApplicationRepository(database.DB)
	deploymentRepo := repository.NewDeploymentRepository(database.DB)
	deploymentHistoryRepo := repository.NewDeploymentHistoryRepository(database.DB)
	teamRepo := repository.NewTeamRepository(database.DB)
	projectTeamRepo := repository.NewProjectTeamRepository(database.DB)
	applicationTeamRepo := repository.NewApplicationTeamRepository(database.DB)
	teamMemberRepo := repository.NewTeamMemberRepository(database.DB)
	incidentRepo := repository.NewIncidentRepository(database.DB)
	alertRepo := repository.NewAlertRepository(database.DB)
	clusterRepo := repository.NewClusterRepository(database.DB)
	resourceRepo := repository.NewResourceRepository(database.DB)
	metricRepo := repository.NewMetricRepository(database.DB)
	commentRepo := repository.NewCommentRepository(database.DB)
	auditRepo := repository.NewAuditRepository(database.DB)
	dashboardRepo := repository.NewDashboardRepository(database.DB)
	notificationChannelRepo := repository.NewNotificationChannelRepository(database.DB)
	applicationSLORepo := repository.NewApplicationSLORepository(database.DB)
	integrationRepo := repository.NewIntegrationRepository(database.DB)
	githubRepositoryRepo := repository.NewGitHubRepositoryRepository(database.DB)
	applicationGitHubRepositoryRepo := repository.NewApplicationGitHubRepositoryRepository(database.DB)

	organizationService := services.NewOrganizationService(organizationRepo)
	auditService := services.NewAuditService(auditRepo).
		WithProjectRepo(projectRepo).
		WithIncidentRepo(incidentRepo)
	userService := services.NewUserService(userRepo, organizationRepo, invitationRepo, cfg).
		WithAuditService(auditService)
	invitationService := services.NewInvitationService(invitationRepo, userRepo, organizationRepo).
		WithAuditService(auditService)
	projectService := services.NewProjectService(projectRepo, userRepo, auditService)
	applicationService := services.NewApplicationService(applicationRepo, projectRepo).
		WithAuditService(auditService)
	deploymentHistoryService := services.NewDeploymentHistoryService(deploymentHistoryRepo, deploymentRepo)
	deploymentService := services.NewDeploymentService(deploymentRepo, applicationRepo, projectRepo, clusterRepo, deploymentHistoryService).
		WithAuditService(auditService)
	teamService := services.NewTeamService(teamRepo, teamMemberRepo, userRepo).
		WithAuditService(auditService)
	projectTeamService := services.NewProjectTeamService(projectTeamRepo, projectRepo, teamRepo).
		WithAuditService(auditService)
	applicationTeamService := services.NewApplicationTeamService(applicationTeamRepo, applicationRepo, teamRepo).
		WithAuditService(auditService)
	incidentService := services.NewIncidentService(incidentRepo, commentRepo, auditRepo, auditService).
		WithApplicationRepo(applicationRepo).
		WithTeamRepo(teamRepo).
		WithUserRepo(userRepo)
	alertService := services.NewAlertService(alertRepo, incidentRepo, auditService)
	clusterCredentialCipher, err := security.NewClusterCredentialCipher(cfg.ClusterCredentialEncryptionKey)
	if err != nil {
		return fmt.Errorf("initialize cluster credential cipher: %w", err)
	}
	notificationProviders := map[string]notification.Provider{
		constants.NotificationChannelEmail: notification.NewEmailProvider(notification.EmailConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
		}),
		constants.NotificationChannelSlack:   notification.NewSlackProvider(),
		constants.NotificationChannelTeams:   notification.NewTeamsProvider(),
		constants.NotificationChannelWebhook: notification.NewWebhookProvider(),
	}
	notificationService := services.NewNotificationService(notificationChannelRepo, clusterCredentialCipher, notificationProviders).
		WithDeploymentRepo(deploymentRepo).
		WithTeamRepo(teamRepo).
		WithAuditService(auditService)
	incidentService.WithNotificationService(notificationService)
	alertService.WithNotificationService(notificationService)
	sreMetricsService := services.NewSREMetricsService(applicationRepo, incidentRepo, applicationSLORepo).
		WithAuditService(auditService)
	rcaService := services.NewRCAService(incidentRepo, alertRepo, metricRepo, deploymentRepo, auditRepo, applicationRepo)
	// connectorRegistry has a real GitHub connector registered (Sprint 28) -
	// every other integration type still resolves to a "not implemented"
	// connector (see internal/connector) until a future sprint implements
	// Slack/Email/Prometheus/Loki/OpenTelemetry/Azure. The same
	// cluster-credential cipher already used for kubeconfigs and
	// notification channel targets is reused for integration credentials,
	// per CLUSTER_CREDENTIAL_ENCRYPTION_KEY - no second encryption key.
	connectorRegistry := connector.NewRegistry()
	githubClient := githubconnector.NewClient("", nil)
	connectorRegistry.Register(constants.IntegrationTypeGitHub, githubconnector.NewConnector(githubClient))
	integrationService := services.NewIntegrationService(integrationRepo, clusterCredentialCipher, connectorRegistry).
		WithAuditService(auditService)
	githubOAuthConfig := githubconnector.NewOAuthConfig(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.GitHubOAuthRedirectURL)
	githubService := services.NewGitHubService(
		integrationService,
		githubClient,
		githubOAuthConfig,
		githubRepositoryRepo,
		applicationGitHubRepositoryRepo,
		applicationRepo,
		deploymentRepo,
		cfg.JWTSecret,
	).WithAuditService(auditService)
	podService := k8spods.NewPodService(applicationRepo, clusterRepo, clusterCredentialCipher)
	kubernetesLogRuntime := k8slogs.NewLogService(applicationRepo, clusterRepo, clusterCredentialCipher)
	kubernetesConfigMapRuntime := k8sconfigmaps.NewConfigMapService(applicationRepo, clusterRepo, clusterCredentialCipher)
	kubernetesSecretRuntime := k8ssecrets.NewSecretService(applicationRepo, clusterRepo, clusterCredentialCipher)
	kubernetesReplicaSetRuntime := k8sreplicasets.NewReplicaSetService(
		applicationRepo,
		clusterRepo,
		clusterCredentialCipher,
	)
	kubernetesRuntimeDeploymentRuntime := k8sruntimedeployments.NewDeploymentRuntimeService(applicationRepo, clusterRepo, clusterCredentialCipher)
	applicationHealthService := services.NewApplicationHealthService(applicationRepo, deploymentRepo, alertRepo, incidentRepo).
		WithPodService(podService).
		WithRuntimeDeploymentService(kubernetesRuntimeDeploymentRuntime)
	kubernetesServiceRuntime := k8sservices.NewServiceService(applicationRepo, clusterRepo, clusterCredentialCipher)
	kubernetesIngressRuntime := k8singresses.NewIngressService(applicationRepo, clusterRepo, clusterCredentialCipher)
	kubernetesEventRuntime := k8sevents.NewEventService(applicationRepo, clusterRepo, clusterCredentialCipher)
	rcaService.
		WithPodService(podService).
		WithEventService(kubernetesEventRuntime).
		WithLogService(kubernetesLogRuntime)
	kubernetesNodeRuntime := k8snodes.NewNodeService(clusterRepo, clusterCredentialCipher)
	kubernetesNamespaceRuntime := k8snamespaces.NewNamespaceService(clusterRepo, clusterCredentialCipher)
	clusterService := services.NewClusterService(clusterRepo, auditService, clusterCredentialCipher)
	deploymentStatusUpdater := k8sexecutor.NewDeploymentStatusUpdater(deploymentRepo, deploymentHistoryService, auditService).
		WithNotifier(notificationService)
	deploymentManifestBuilder := k8sexecutor.NewDeploymentManifestBuilder()
	deploymentExecutor := k8sexecutor.NewDeploymentExecutor(
		deploymentRepo,
		applicationRepo,
		clusterRepo,
		clusterCredentialCipher,
		deploymentManifestBuilder,
		deploymentStatusUpdater,
		k8sexecutor.NewClientsetFactory(),
		90*time.Second,
	)
	deploymentService.WithExecutor(deploymentExecutor)
	resourceSyncEngine := resourcesync.NewSyncEngine(resourceRepo)
	resourceService := services.NewResourceService(resourceRepo, resourceSyncEngine, auditService)
	metricService := services.NewMetricService(metricRepo, auditService).WithResourceRepo(resourceRepo)
	metricsStatusService := services.NewMetricsStatusService(clusterRepo, metricRepo, clusterCredentialCipher)
	kubernetesProviderFactory := bootstrap.NewKubernetesDiscoveryProviderFactory()
	runtimeClusterCatalog := &discoveryClusterLoader{db: database.DB, repo: clusterRepo}
	discoveryWorker := discovery.NewDiscoveryWorker(
		nil,
		runtimeClusterCatalog,
		resourceService,
		kubernetesProviderFactory,
		auditService,
	)
	discoveryBootstrap := bootstrap.NewDiscoveryBootstrap(0, 1, discoveryWorker, kubernetesProviderFactory)
	metricsBootstrap := bootstrap.NewMetricsBootstrap(nil, metricService, 0).WithAlertReconciler(alertService)
	runtimeBootstrap := bootstrap.NewRuntime(bootstrap.RuntimeDependencies{
		Catalog:            runtimeClusterCatalog,
		DiscoveryBootstrap: discoveryBootstrap,
		MetricsBootstrap:   metricsBootstrap,
	})
	commentService := services.NewCommentService(commentRepo, incidentRepo, auditService)
	dashboardService := services.NewDashboardService(dashboardRepo)

	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)
	projectHandler := handlers.NewProjectHandler(projectService)
	applicationHandler := handlers.NewApplicationHandler(applicationService)
	applicationHealthHandler := handlers.NewApplicationHealthHandler(applicationHealthService)
	podHandler := handlers.NewPodHandler(podService)
	kubernetesLogHandler := handlers.NewKubernetesLogHandler(kubernetesLogRuntime)
	kubernetesConfigMapHandler := handlers.NewKubernetesConfigMapHandler(kubernetesConfigMapRuntime)
	kubernetesSecretHandler := handlers.NewKubernetesSecretHandler(kubernetesSecretRuntime)
	kubernetesReplicaSetHandler := handlers.NewKubernetesReplicaSetHandler(
		kubernetesReplicaSetRuntime,
	)
	kubernetesRuntimeDeploymentHandler := handlers.NewKubernetesRuntimeDeploymentHandler(kubernetesRuntimeDeploymentRuntime)
	kubernetesServiceHandler := handlers.NewKubernetesServiceHandler(kubernetesServiceRuntime)
	kubernetesIngressHandler := handlers.NewKubernetesIngressHandler(kubernetesIngressRuntime)
	kubernetesEventHandler := handlers.NewKubernetesEventHandler(kubernetesEventRuntime)
	kubernetesNodeHandler := handlers.NewKubernetesNodeHandler(kubernetesNodeRuntime)
	kubernetesNamespaceHandler := handlers.NewKubernetesNamespaceHandler(kubernetesNamespaceRuntime)
	deploymentHandler := handlers.NewDeploymentHandler(deploymentService)
	deploymentHistoryHandler := handlers.NewDeploymentHistoryHandler(deploymentHistoryService)
	teamHandler := handlers.NewTeamHandler(teamService)
	projectTeamHandler := handlers.NewProjectTeamHandler(projectTeamService)
	applicationTeamHandler := handlers.NewApplicationTeamHandler(applicationTeamService)
	incidentHandler := handlers.NewIncidentHandler(incidentService)
	alertHandler := handlers.NewAlertHandler(alertService)
	metricHandler := handlers.NewMetricHandler(metricService, metricsStatusService)
	clusterHandler := handlers.NewClusterHandler(clusterService)
	resourceHandler := handlers.NewResourceHandler(resourceService, clusterService, discoveryWorker)
	commentHandler := handlers.NewCommentHandler(commentService)
	auditHandler := handlers.NewAuditHandler(auditService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	sreHandler := handlers.NewSREHandler(sreMetricsService)
	rcaHandler := handlers.NewRCAHandler(rcaService)
	integrationHandler := handlers.NewIntegrationHandler(integrationService)
	githubHandler := handlers.NewGitHubHandler(githubService, cfg)
	healthHandler := handlers.NewHealthHandler(cfg, startedAt, database.Ping)
	collector := metrics.NewCollector()

	r := gin.New()

	if err := r.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("configure trusted proxies: %w", err)
	}

	// -----------------------------
	// CORS Middleware
	// -----------------------------
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(
		middleware.RequestID(),
		middleware.SecurityHeaders(),
		collector.Middleware(),
		middleware.RequestLogger(),
		middleware.RateLimit(
			middleware.NewIPRateLimiterWithWindow(
				cfg.RateLimit,
				cfg.RateLimitWindow,
			),
		),
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
		applicationHealthHandler,
		podHandler,
		kubernetesLogHandler,
		kubernetesConfigMapHandler,
		kubernetesSecretHandler,
		kubernetesReplicaSetHandler,
		kubernetesRuntimeDeploymentHandler,
		kubernetesServiceHandler,
		kubernetesIngressHandler,
		kubernetesEventHandler,
		deploymentHandler,
		deploymentHistoryHandler,
		teamHandler,
		projectTeamHandler,
		applicationTeamHandler,
		incidentHandler,
		alertHandler,
		metricHandler,
		clusterHandler,
		kubernetesNodeHandler,
		kubernetesNamespaceHandler,
		resourceHandler,
		commentHandler,
		auditHandler,
		dashboardHandler,
		notificationHandler,
		sreHandler,
		rcaHandler,
		integrationHandler,
		githubHandler,
		healthHandler,
		collector,
	)
	healthHandler.SetInitialized(true)
	if err := runtimeBootstrap.Startup(processContext); err != nil {
		return fmt.Errorf("start runtime bootstrap: %w", err)
	}
	defer func() {
		runtimeShutdownContext, cancelRuntimeShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancelRuntimeShutdown()
		if err := runtimeBootstrap.Shutdown(runtimeShutdownContext); err != nil {
			logger.Error(processContext, "runtime bootstrap shutdown failed", slog.Any("error", err))
		}
	}()

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	logger.Info(
		processContext,
		"application started",
		slog.String("app_name", cfg.AppName),
		slog.String("environment", cfg.AppEnv),
		slog.String("version", cfg.Version),
		slog.String("address", server.Addr),
	)

	signalContext, stopSignals := signal.NotifyContext(
		processContext,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stopSignals()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil

	case <-signalContext.Done():
	}

	logger.Info(processContext, "shutdown signal received; draining active requests")

	shutdownContext, cancelShutdown := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("graceful HTTP shutdown: %w", err)
	}

	logger.Info(processContext, "HTTP server shutdown complete")

	return nil
}
