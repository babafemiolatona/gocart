package main

import (
	"log"

	"gocart/internal/config"
	"gocart/internal/repositories"
	"gocart/internal/routes"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize database
	db, err := repositories.InitDB(config.CFG)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create service layer
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo, config.CFG)

	// Set Gin mode
	if config.CFG.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, db, userService)

	// Start server
	log.Printf("Starting server on %s", config.CFG.ServerPort)
	if err := router.Run(config.CFG.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
