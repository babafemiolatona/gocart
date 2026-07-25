package routes

import (
	"gocart/internal/handlers"
	"gocart/internal/middleware"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"gocart/internal/services"
	"gocart/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(
	router *gin.Engine,
	db *gorm.DB,
	userService *services.UserService,
	storage storage.Storage,
) {

	router.Use(middleware.ErrorHandler())

	router.GET("/health", handlers.HealthCheck)

	v1 := router.Group("/api/v1")
	{
		userHandler := handlers.NewUserHandler(userService)

		// ----------------------------------
		// Authentication
		// ----------------------------------

		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// ----------------------------------
		// Repositories
		// ----------------------------------

		userRepo := repositories.NewUserRepository(db)
		productRepo := repositories.NewProductRepository(db)
		categoryRepo := repositories.NewCategoryRepository(db)
		cartRepo := repositories.NewCartRepository(db)
		orderRepo := repositories.NewOrderRepository(db)
		paymentRepo := repositories.NewPaymentRepository(db)
		productImageRepo := repositories.NewProductImageRepository(db)
		merchantRepo := repositories.NewMerchantRepository(db)

		// ----------------------------------
		// Services
		// ----------------------------------

		productService := services.NewProductService(
			productRepo,
			categoryRepo,
			productImageRepo,
			storage,
		)

		categoryService := services.NewCategoryService(categoryRepo)

		cartService := services.NewCartService(
			cartRepo,
			productRepo,
		)

		orderService := services.NewOrderService(
			orderRepo,
			cartRepo,
			productRepo,
			paymentRepo,
		)

		paymentService := services.NewPaymentService(
			paymentRepo,
			orderRepo,
			cartRepo,
			productRepo,
		)

		merchantService := services.NewMerchantService(
			merchantRepo,
			userRepo,
		)

		// ----------------------------------
		// Handlers
		// ----------------------------------

		productHandler := handlers.NewProductHandler(productService)
		categoryHandler := handlers.NewCategoryHandler(categoryService)
		cartHandler := handlers.NewCartHandler(cartService)
		orderHandler := handlers.NewOrderHandler(orderService)
		paymentHandler := handlers.NewPaymentHandler(paymentService)
		merchantHandler := handlers.NewMerchantHandler(merchantService)

		// ----------------------------------
		// Public Routes
		// ----------------------------------

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

		// ----------------------------------
		// Authenticated Routes
		// ----------------------------------

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(userService))
		{
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
			}

			cart := protected.Group("/cart")
			{
				cart.GET("", cartHandler.GetCart)
				cart.POST("/items", cartHandler.AddToCart)
				cart.PUT("/items/:itemID", cartHandler.UpdateCartItem)
				cart.DELETE("/items/:itemID", cartHandler.RemoveFromCart)
				cart.DELETE("", cartHandler.ClearCart)
			}

			orders := protected.Group("/orders")
			{
				orders.POST("/checkout", orderHandler.Checkout)
				orders.GET("", orderHandler.GetMyOrders)
				orders.GET("/:id", orderHandler.GetOrder)
				orders.PUT("/:id/cancel", orderHandler.CancelOrder)
			}

			payments := protected.Group("/payments")
			{
				payments.POST("/:reference/process", paymentHandler.ProcessPayment)
				payments.GET("/:reference", paymentHandler.GetPayment)
			}

			// ----------------------------------
			// Merchant Registration
			// (Customer -> Merchant)
			// ----------------------------------

			merchant := protected.Group("/merchant")
			{
				merchant.POST("/register", merchantHandler.RegisterMerchant)
			}

			// ----------------------------------
			// Merchant-only Routes
			// ----------------------------------

			merchantProtected := protected.Group("/merchant")
			merchantProtected.Use(
				middleware.RequireMerchant(merchantRepo),
			)
			{
				merchantProtected.GET("/profile", merchantHandler.GetProfile)
				merchantProtected.PUT("/profile", merchantHandler.UpdateProfile)

				products := merchantProtected.Group("/products")
				{
					products.GET("", productHandler.GetMerchantProducts)
					products.GET("/:id", productHandler.GetMerchantProduct)
					products.POST("", productHandler.CreateProduct)
					products.PUT("/:id", productHandler.UpdateProduct)
					products.DELETE("/:id", productHandler.DeleteProduct)
				}
			}
		}

		// ----------------------------------
		// Admin Routes
		// ----------------------------------

		admin := v1.Group("/admin")
		admin.Use(
			middleware.AuthMiddleware(userService),
			middleware.RequireRole(models.RoleAdmin),
		)
		{
			categories := admin.Group("/categories")
			{
				categories.POST("", categoryHandler.CreateCategory)
				categories.PUT("/:id", categoryHandler.UpdateCategory)
				categories.DELETE("/:id", categoryHandler.DeleteCategory)
			}
		}
	}
}
