package routes

import (
	"gocart/internal/handlers"
	"gocart/internal/middleware"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, userService *services.UserService) {
	router.GET("/health", handlers.HealthCheck)

	v1 := router.Group("/api/v1")
	{
		userHandler := handlers.NewUserHandler(userService)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(userService))
		{
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
			}
		}
	}
}
