package main

import (
	"log"

	"gocart/internal/config"
	"gocart/internal/repositories"
	"gocart/internal/routes"
	"gocart/internal/seed"
	"gocart/internal/services"
	"gocart/internal/storage"

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

	log.Printf("Endpoint: %q", config.CFG.MinioEndpoint)
	log.Printf("AccessKey: %q", config.CFG.MinioAccessKey)
	log.Printf("Bucket: %q", config.CFG.MinioBucket)
	log.Printf("UseSSL: %v", config.CFG.MinioUseSSL)

	// Initialize MinIO
	minioStorage, err := storage.NewMinioStorage(
		config.CFG.MinioEndpoint,
		config.CFG.MinioAccessKey,
		config.CFG.MinioSecretKey,
		config.CFG.MinioBucket,
		config.CFG.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}

	// Create service layer
	userRepo := repositories.NewUserRepository(db)

	if err := seed.SeedAdmin(userRepo); err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	userService := services.NewUserService(userRepo, config.CFG)

	// Set Gin mode
	if config.CFG.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, db, userService, minioStorage)

	// Start server
	log.Printf("Starting server on %s", config.CFG.ServerPort)
	if err := router.Run(config.CFG.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
