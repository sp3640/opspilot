package router

import (
	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	"github.com/sp3640/opspilot/backend/internal/middleware"
)

func RegisterRoutes(
	r *gin.Engine,
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	incidentHandler *handlers.IncidentHandler,
) {
	r.GET("/health", handlers.HealthCheck)

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
		}
	}
}
