package handlers

import (
	"mime/multipart"

	"gocart/internal/dto"
	"gocart/internal/query"
)

// Consumer-side service interfaces so handlers can be unit-tested with stubs.
// The concrete *services.* implementations satisfy these interfaces.

type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
}

type CartService interface {
	GetCartResponse(userID uint) (*dto.CartResponse, error)
	AddToCart(userID uint, req *dto.AddToCartRequest) (*dto.CartResponse, error)
	UpdateCartItem(userID, itemID uint, qty int) (*dto.CartResponse, error)
	RemoveFromCart(userID, itemID uint) (*dto.CartResponse, error)
	ClearCart(userID uint) error
}

type CategoryService interface {
	CreateCategory(req *dto.CategoryRequest) (*dto.CategoryResponse, error)
	GetCategoryByID(id uint) (*dto.CategoryResponse, error)
	GetAllCategories() ([]dto.CategoryResponse, error)
	UpdateCategory(req *dto.UpdateCategoryRequest, id uint) (*dto.CategoryResponse, error)
	DeleteCategory(id uint) error
}

type MerchantService interface {
	RegisterMerchant(userID uint, req *dto.MerchantRegisterRequest) (*dto.MerchantResponse, error)
	GetProfile(merchantID uint) (*dto.MerchantResponse, error)
	UpdateProfile(merchantID uint, req *dto.UpdateMerchantRequest) (*dto.MerchantResponse, error)
	GetOrders(merchantID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error)
	GetOrder(merchantID uint, orderID uint) (*dto.MerchantOrderResponse, error)
	UpdateOrderStatus(merchantID, orderID uint, req *dto.UpdateOrderStatusRequest) error
	GetDashboard(merchantID uint) (*dto.MerchantDashboardResponse, error)
}

type OrderService interface {
	ProcessCheckout(userID uint, shippingAddress, idempotencyKey string) (*dto.CheckoutResponse, error)
	GetUserOrders(userID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error)
	GetOrder(userID, orderID uint) (*dto.OrderDetailsResponse, error)
	CancelOrder(userID, orderID uint) error
}

type PaymentService interface {
	ProcessPayment(userID uint, reference string) (*dto.PaymentResponse, error)
	GetPayment(userID uint, reference string) (*dto.PaymentResponse, error)
	ProcessWebhook(body []byte, signature string) (*dto.PaymentResponse, error)
	SimulateWebhook(reference string, status string) (*dto.PaymentResponse, error)
}

type ProductService interface {
	CreateProduct(merchantID uint, req *dto.CreateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error)
	GetProducts(query *dto.PaginationQuery, filters *query.ProductFilters) (*dto.PaginatedResponse, error)
	GetProduct(id uint) (*dto.ProductResponse, error)
	UpdateProduct(merchantID, id uint, req *dto.UpdateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error)
	DeleteProduct(merchantID, id uint) error
	GetMerchantProduct(merchantID, id uint) (*dto.ProductResponse, error)
}

type UserService interface {
	GetMe(userID uint) (*dto.UserResponse, error)
}
