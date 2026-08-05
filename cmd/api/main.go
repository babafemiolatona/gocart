package main

// @title GoCart API
// @version 1.0
// @description E-commerce REST API built with Gin.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Initialize database
	db, err := repositories.InitDB(cfg)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to initialize database")
	}

	logger.Log.Info().
		Str("endpoint", cfg.MinioEndpoint).
		Str("bucket", cfg.MinioBucket).
		Bool("use_ssl", cfg.MinioUseSSL).
		Msg("minio configuration")

	// Initialize MinIO
	minioStorage, err := storage.NewMinioStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to initialize minio")
	}

	// Create service layer
	authRepo := repositories.NewAuthRepository(db)

	if err := seed.SeedAdmin(
		authRepo,
		cfg.SeedAdminEmail,
		cfg.SeedAdminPassword,
	); err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to seed admin")
	}

	AuthService := services.NewAuthService(authRepo, cfg)

	// Set Gin mode
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup routes
	routes.SetupRoutes(router, db, AuthService, minioStorage)

	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: router,
	}

	go func() {
		logger.Log.Info().Str("port", cfg.ServerPort).Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Log.Info().Msg("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	logger.Log.Info().Msg("server exited gracefully")
}
