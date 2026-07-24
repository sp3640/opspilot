package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"service":     "OpsPilot Backend",
		"environment": "development",
		"version":     "v1.0.0",
		"timestamp":   time.Now().UTC(),
	})
}
