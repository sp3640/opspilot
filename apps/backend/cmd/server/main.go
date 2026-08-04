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
	"github.com/sp3640/opspilot/backend/internal/database"
	"github.com/sp3640/opspilot/backend/internal/discovery"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/metrics"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
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
	teamRepo := repository.NewTeamRepository(database.DB)
	projectTeamRepo := repository.NewProjectTeamRepository(database.DB)
	teamMemberRepo := repository.NewTeamMemberRepository(database.DB)
	incidentRepo := repository.NewIncidentRepository(database.DB)
	alertRepo := repository.NewAlertRepository(database.DB)
	clusterRepo := repository.NewClusterRepository(database.DB)
	resourceRepo := repository.NewResourceRepository(database.DB)
	metricRepo := repository.NewMetricRepository(database.DB)
	commentRepo := repository.NewCommentRepository(database.DB)
	auditRepo := repository.NewAuditRepository(database.DB)
	dashboardRepo := repository.NewDashboardRepository(database.DB)

	userService := services.NewUserService(userRepo, organizationRepo, cfg)
	organizationService := services.NewOrganizationService(organizationRepo)
	invitationService := services.NewInvitationService(invitationRepo, userRepo)
	auditService := services.NewAuditService(auditRepo).
		WithProjectRepo(projectRepo).
		WithIncidentRepo(incidentRepo)
	projectService := services.NewProjectService(projectRepo, userRepo, auditService)
	applicationService := services.NewApplicationService(applicationRepo, projectRepo)
	teamService := services.NewTeamService(teamRepo, teamMemberRepo, userRepo)
	projectTeamService := services.NewProjectTeamService(projectTeamRepo, projectRepo, teamRepo)
	incidentService := services.NewIncidentService(incidentRepo, commentRepo, auditRepo, auditService)
	alertService := services.NewAlertService(alertRepo, incidentRepo, auditService)
	clusterCredentialCipher, err := security.NewClusterCredentialCipher(cfg.ClusterCredentialEncryptionKey)
	if err != nil {
		return fmt.Errorf("initialize cluster credential cipher: %w", err)
	}
	clusterService := services.NewClusterService(clusterRepo, auditService, clusterCredentialCipher)
	resourceSyncEngine := resourcesync.NewSyncEngine(resourceRepo)
	resourceService := services.NewResourceService(resourceRepo, resourceSyncEngine, auditService)
	metricService := services.NewMetricService(metricRepo, auditService)
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
	metricsBootstrap := bootstrap.NewMetricsBootstrap(nil, metricService, 0)
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
	teamHandler := handlers.NewTeamHandler(teamService)
	projectTeamHandler := handlers.NewProjectTeamHandler(projectTeamService)
	incidentHandler := handlers.NewIncidentHandler(incidentService)
	alertHandler := handlers.NewAlertHandler(alertService)
	metricHandler := handlers.NewMetricHandler(metricService)
	clusterHandler := handlers.NewClusterHandler(clusterService)
	resourceHandler := handlers.NewResourceHandler(resourceService, clusterService, discoveryWorker)
	commentHandler := handlers.NewCommentHandler(commentService)
	auditHandler := handlers.NewAuditHandler(auditService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
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
