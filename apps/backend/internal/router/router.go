package router

import (
	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/handlers"
)

func RegisterRoutes(
	r *gin.Engine,
	authHandler *handlers.AuthHandler,
) {
	r.GET("/health", handlers.HealthCheck)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
	}
}