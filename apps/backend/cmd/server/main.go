package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/database"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/router"
	"github.com/sp3640/opspilot/backend/internal/services"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)
	repo := repository.NewUserRepository(database.DB)
	userService := services.NewUserService(repo,cfg)
	authHandler := handlers.NewAuthHandler(userService)

	r := gin.Default()

	router.RegisterRoutes(r, authHandler)

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