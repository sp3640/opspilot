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
) {
	r.GET("/health", handlers.HealthCheck)

	api := r.Group("/api/v1")
	{
		// Public Routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected Routes
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg))
		{
			users.GET("/me", userHandler.Me)
		}
	}
}