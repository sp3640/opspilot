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

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/database"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/metrics"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/router"
	"github.com/sp3640/opspilot/backend/internal/services"
)

const shutdownTimeout = 30 * time.Second

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
	projectRepo := repository.NewProjectRepository(database.DB)
	incidentRepo := repository.NewIncidentRepository(database.DB)
	commentRepo := repository.NewCommentRepository(database.DB)
	auditRepo := repository.NewAuditRepository(database.DB)
	dashboardRepo := repository.NewDashboardRepository(database.DB)

	userService := services.NewUserService(userRepo, cfg)
	auditService := services.NewAuditService(auditRepo).
		WithProjectRepo(projectRepo).
		WithIncidentRepo(incidentRepo)
	projectService := services.NewProjectService(projectRepo, userRepo, auditService)
	incidentService := services.NewIncidentService(incidentRepo, auditService)
	commentService := services.NewCommentService(commentRepo, incidentRepo, auditService)
	dashboardService := services.NewDashboardService(dashboardRepo)

	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	projectHandler := handlers.NewProjectHandler(projectService)
	incidentHandler := handlers.NewIncidentHandler(incidentService)
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
		MaxAge: 12 * time.Hour,
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
		projectHandler,
		incidentHandler,
		commentHandler,
		auditHandler,
		dashboardHandler,
		healthHandler,
		collector,
	)

	healthHandler.SetInitialized(true)

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