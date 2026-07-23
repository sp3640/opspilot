package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/router"
)

func main() {
	cfg := config.Load()

	r := gin.Default()

	router.RegisterRoutes(r)

	log.Printf(
		"🚀 %s running in %s mode on port %s",
		cfg.AppName,
		cfg.AppEnv,
		cfg.Port,
	)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}