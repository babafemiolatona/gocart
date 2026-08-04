package main

// @title GoCart API
// @version 1.0
// @description E-commerce REST API built with Gin.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	_ "gocart/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gocart/internal/config"
	"gocart/internal/logger"
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
		logger.Log.Fatal().Err(err).Msg("failed to initialize database")
	}

	logger.Log.Info().
		Str("endpoint", config.CFG.MinioEndpoint).
		Str("access_key", config.CFG.MinioAccessKey).
		Str("bucket", config.CFG.MinioBucket).
		Bool("use_ssl", config.CFG.MinioUseSSL).
		Msg("minio configuration")

	// Initialize MinIO
	minioStorage, err := storage.NewMinioStorage(
		config.CFG.MinioEndpoint,
		config.CFG.MinioAccessKey,
		config.CFG.MinioSecretKey,
		config.CFG.MinioBucket,
		config.CFG.MinioUseSSL,
	)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to initialize minio")
	}

	// Create service layer
	authRepo := repositories.NewAuthRepository(db)

	if err := seed.SeedAdmin(
		authRepo,
		config.CFG.SeedAdminEmail,
		config.CFG.SeedAdminPassword,
	); err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to seed admin")
	}

	AuthService := services.NewAuthService(authRepo, config.CFG)

	// Set Gin mode
	if config.CFG.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup routes
	routes.SetupRoutes(router, db, AuthService, minioStorage)

	// Start server
	logger.Log.Info().Str("port", config.CFG.ServerPort).Msg("starting server")
	if err := router.Run(config.CFG.ServerPort); err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to start server")
	}
}
