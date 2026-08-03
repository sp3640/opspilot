package router

import (
	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	"github.com/sp3640/opspilot/backend/internal/metrics"
	"github.com/sp3640/opspilot/backend/internal/middleware"
)

func RegisterRoutes(
	r *gin.Engine,
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	incidentHandler *handlers.IncidentHandler,
	clusterHandler *handlers.ClusterHandler,
	resourceHandler *handlers.ResourceHandler,
	commentHandler *handlers.CommentHandler,
	auditHandler *handlers.AuditHandler,
	dashboardHandler *handlers.DashboardHandler,
	healthHandler *handlers.HealthHandler,
	collector *metrics.Collector,
) {
	if healthHandler != nil {
		r.GET("/health", healthHandler.HealthCheck)
		r.GET("/ready", healthHandler.Ready)
		r.GET("/live", healthHandler.Live)
	}
	if collector != nil {
		r.GET("/metrics", collector.Handler)
	}

	api := r.Group("/api/v1")
	{
		// =========================
		// Public Routes
		// =========================
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// =========================
		// Protected User Routes
		// =========================
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg))
		{
			users.GET("/me", userHandler.Me)
		}

		// =========================
		// Protected Project Routes
		// =========================
		projects := api.Group("/projects")
		projects.Use(middleware.AuthMiddleware(cfg))
		{
			projects.POST("", projectHandler.Create)
			projects.GET("", projectHandler.List)
			projects.GET("/:id", projectHandler.GetByID)
			projects.PUT("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)
			projects.GET("/:id/audit-logs", auditHandler.GetProjectAuditLogs)
		}

		// =========================
		// Protected Incident Routes
		// =========================
		incidents := api.Group("/incidents")
		incidents.Use(middleware.AuthMiddleware(cfg))
		{
			incidents.POST("", incidentHandler.Create)
			incidents.GET("", incidentHandler.List)
			incidents.GET("/:id", incidentHandler.GetByID)
			incidents.PUT("/:id", incidentHandler.Update)
			incidents.DELETE("/:id", incidentHandler.Delete)
			incidents.POST("/:id/comments", commentHandler.Create)
			incidents.GET("/:id/comments", commentHandler.List)
			incidents.GET("/:id/audit-logs", auditHandler.GetIncidentAuditLogs)
		}

		clusters := api.Group("/clusters")
		clusters.Use(middleware.AuthMiddleware(cfg))
		{
			clusters.POST("", clusterHandler.Create)
			clusters.GET("", clusterHandler.List)
			clusters.GET("/:id", clusterHandler.GetByID)
			clusters.PUT("/:id", clusterHandler.Update)
			clusters.DELETE("/:id", clusterHandler.Delete)
			clusters.POST("/:id/validate", clusterHandler.Validate)
			clusters.POST("/:id/default", clusterHandler.SetDefault)
		}

		resources := api.Group("/resources")
		resources.Use(middleware.AuthMiddleware(cfg))
		{
			resources.POST("/sync", resourceHandler.Sync)
			resources.POST("", resourceHandler.Create)
			resources.GET("", resourceHandler.List)
			resources.GET("/:id", resourceHandler.GetByID)
			resources.PUT("/:id", resourceHandler.Update)
			resources.DELETE("/:id", resourceHandler.Delete)
		}

		comments := api.Group("/comments")
		comments.Use(middleware.AuthMiddleware(cfg))
		{
			comments.PUT("/:id", commentHandler.Update)
			comments.DELETE("/:id", commentHandler.Delete)
		}

		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware(cfg))
		{
			dashboard.GET("/summary", dashboardHandler.Summary)
			dashboard.GET("/recent-incidents", dashboardHandler.RecentIncidents)
			dashboard.GET("/activity", dashboardHandler.Activity)
			dashboard.GET("/stats", dashboardHandler.Stats)
			dashboard.GET("/services", dashboardHandler.Services)
		}
	}
}
