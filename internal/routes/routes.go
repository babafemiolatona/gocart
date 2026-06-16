package routes

import (
	"gocart/internal/handlers"
	"gocart/internal/middleware"
	"gocart/internal/repositories"
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

		productRepo := repositories.NewProductRepository(db)
		categoryRepo := repositories.NewCategoryRepository(db)

		productService := services.NewProductService(productRepo, categoryRepo)
		categoryService := services.NewCategoryService(categoryRepo)

		productHandler := handlers.NewProductHandler(productService)
		categoryHandler := handlers.NewCategoryHandler(categoryService)

		cartRepo := repositories.NewCartRepository(db)
		cartService := services.NewCartService(cartRepo, productRepo)
		cartHandler := handlers.NewCartHandler(cartService)

		public := v1.Group("")
		{
			products := public.Group("/products")
			{
				products.GET("", productHandler.GetProducts)
				products.GET("/:id", productHandler.GetProduct)
			}

			categories := public.Group("/categories")
			{
				categories.GET("", categoryHandler.GetCategories)
				categories.GET("/:id", categoryHandler.GetCategoryByID)
			}
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(userService))
		{
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
			}

			products := protected.Group("/products")
			{
				products.POST("", productHandler.CreateProduct)
				products.PUT("/:id", productHandler.UpdateProduct)
				products.DELETE("/:id", productHandler.DeleteProduct)
			}

			categories := protected.Group("/categories")
			{
				categories.POST("", categoryHandler.CreateCategory)
				categories.PUT("/:id", categoryHandler.UpdateCategory)
				categories.DELETE("/:id", categoryHandler.DeleteCategory)
			}

			cart := protected.Group("/cart")
			{
				cart.GET("", cartHandler.GetCart)
				cart.POST("/items", cartHandler.AddToCart)
				cart.PUT("/items/:itemID", cartHandler.UpdateCartItem)
				cart.DELETE("/items/:itemID", cartHandler.RemoveFromCart)
				cart.DELETE("", cartHandler.ClearCart)
			}
		}
	}
}
