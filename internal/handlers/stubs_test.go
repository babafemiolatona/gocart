package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gocart/internal/dto"
	"gocart/internal/logger"
	"gocart/internal/middleware"
	"gocart/internal/query"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func init() {
	logger.Log = zerolog.Nop()
}

func newTestRouter(userID uint, merchantID *uint) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(func(c *gin.Context) {
		if userID != 0 {
			c.Set("userID", userID)
		}
		if merchantID != nil {
			c.Set("merchantID", *merchantID)
		}
		c.Next()
	})

	return r
}

func registerHandler(r *gin.Engine, method, path string, handler gin.HandlerFunc) {
	r.Handle(method, path, handler)
}

func doRequest(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeJSON(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) (code, message string) {
	t.Helper()

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	decodeJSON(t, w.Body.Bytes(), &resp)

	return resp.Error.Code, resp.Error.Message
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("expected status %d, got %d (body: %s)", want, w.Code, w.Body.String())
	}
}

func buildMultipartForm(t *testing.T, fields map[string]string) (body *bytes.Buffer, contentType string) {
	t.Helper()

	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	for key, value := range fields {
		if err := w.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	return buf, w.FormDataContentType()
}

func newRequestWithBody(t *testing.T, method, path string, body *bytes.Buffer, contentType string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", contentType)
	return req
}

// ---- stub services ----

type stubAuthService struct {
	registerFn func(req *dto.RegisterRequest) (*dto.UserResponse, error)
	loginFn    func(req *dto.LoginRequest) (*dto.AuthResponse, error)
}

func (s *stubAuthService) Register(req *dto.RegisterRequest) (*dto.UserResponse, error) {
	return s.registerFn(req)
}

func (s *stubAuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	return s.loginFn(req)
}

type stubCartService struct {
	getCartFn        func(userID uint) (*dto.CartResponse, error)
	addToCartFn      func(userID uint, req *dto.AddToCartRequest) (*dto.CartResponse, error)
	updateCartItemFn func(userID, itemID uint, qty int) (*dto.CartResponse, error)
	removeFromCartFn func(userID, itemID uint) (*dto.CartResponse, error)
	clearCartFn      func(userID uint) error
}

func (s *stubCartService) GetCartResponse(userID uint) (*dto.CartResponse, error) {
	return s.getCartFn(userID)
}

func (s *stubCartService) AddToCart(userID uint, req *dto.AddToCartRequest) (*dto.CartResponse, error) {
	return s.addToCartFn(userID, req)
}

func (s *stubCartService) UpdateCartItem(userID, itemID uint, qty int) (*dto.CartResponse, error) {
	return s.updateCartItemFn(userID, itemID, qty)
}

func (s *stubCartService) RemoveFromCart(userID, itemID uint) (*dto.CartResponse, error) {
	return s.removeFromCartFn(userID, itemID)
}

func (s *stubCartService) ClearCart(userID uint) error {
	return s.clearCartFn(userID)
}

type stubCategoryService struct {
	createFn  func(req *dto.CategoryRequest) (*dto.CategoryResponse, error)
	getByIDFn func(id uint) (*dto.CategoryResponse, error)
	getAllFn  func() ([]dto.CategoryResponse, error)
	updateFn  func(req *dto.UpdateCategoryRequest, id uint) (*dto.CategoryResponse, error)
	deleteFn  func(id uint) error
}

func (s *stubCategoryService) CreateCategory(req *dto.CategoryRequest) (*dto.CategoryResponse, error) {
	return s.createFn(req)
}

func (s *stubCategoryService) GetCategoryByID(id uint) (*dto.CategoryResponse, error) {
	return s.getByIDFn(id)
}

func (s *stubCategoryService) GetAllCategories() ([]dto.CategoryResponse, error) {
	return s.getAllFn()
}

func (s *stubCategoryService) UpdateCategory(req *dto.UpdateCategoryRequest, id uint) (*dto.CategoryResponse, error) {
	return s.updateFn(req, id)
}

func (s *stubCategoryService) DeleteCategory(id uint) error {
	return s.deleteFn(id)
}

type stubMerchantService struct {
	registerFn      func(userID uint, req *dto.MerchantRegisterRequest) (*dto.MerchantResponse, error)
	getProfileFn    func(merchantID uint) (*dto.MerchantResponse, error)
	updateProfileFn func(merchantID uint, req *dto.UpdateMerchantRequest) (*dto.MerchantResponse, error)
	getOrdersFn     func(merchantID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error)
	getOrderFn      func(merchantID, orderID uint) (*dto.MerchantOrderResponse, error)
	updateStatusFn  func(merchantID, orderID uint, req *dto.UpdateOrderStatusRequest) error
	getDashboardFn  func(merchantID uint) (*dto.MerchantDashboardResponse, error)
}

func (s *stubMerchantService) RegisterMerchant(userID uint, req *dto.MerchantRegisterRequest) (*dto.MerchantResponse, error) {
	return s.registerFn(userID, req)
}

func (s *stubMerchantService) GetProfile(merchantID uint) (*dto.MerchantResponse, error) {
	return s.getProfileFn(merchantID)
}

func (s *stubMerchantService) UpdateProfile(merchantID uint, req *dto.UpdateMerchantRequest) (*dto.MerchantResponse, error) {
	return s.updateProfileFn(merchantID, req)
}

func (s *stubMerchantService) GetOrders(merchantID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error) {
	return s.getOrdersFn(merchantID, p)
}

func (s *stubMerchantService) GetOrder(merchantID, orderID uint) (*dto.MerchantOrderResponse, error) {
	return s.getOrderFn(merchantID, orderID)
}

func (s *stubMerchantService) UpdateOrderStatus(merchantID, orderID uint, req *dto.UpdateOrderStatusRequest) error {
	return s.updateStatusFn(merchantID, orderID, req)
}

func (s *stubMerchantService) GetDashboard(merchantID uint) (*dto.MerchantDashboardResponse, error) {
	return s.getDashboardFn(merchantID)
}

type stubOrderService struct {
	checkoutFn    func(userID uint, shippingAddress, idempotencyKey string) (*dto.CheckoutResponse, error)
	getUserOrders func(userID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error)
	getOrderFn    func(userID, orderID uint) (*dto.OrderDetailsResponse, error)
	cancelFn      func(userID, orderID uint) error
}

func (s *stubOrderService) ProcessCheckout(userID uint, shippingAddress, idempotencyKey string) (*dto.CheckoutResponse, error) {
	return s.checkoutFn(userID, shippingAddress, idempotencyKey)
}

func (s *stubOrderService) GetUserOrders(userID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error) {
	return s.getUserOrders(userID, p)
}

func (s *stubOrderService) GetOrder(userID, orderID uint) (*dto.OrderDetailsResponse, error) {
	return s.getOrderFn(userID, orderID)
}

func (s *stubOrderService) CancelOrder(userID, orderID uint) error {
	return s.cancelFn(userID, orderID)
}

type stubPaymentService struct {
	processFn         func(userID uint, reference string) (*dto.PaymentResponse, error)
	getFn             func(userID uint, reference string) (*dto.PaymentResponse, error)
	processWebhookFn  func(body []byte, signature string) (*dto.PaymentResponse, error)
	simulateWebhookFn func(reference string, status string) (*dto.PaymentResponse, error)
}

func (s *stubPaymentService) ProcessPayment(userID uint, reference string) (*dto.PaymentResponse, error) {
	return s.processFn(userID, reference)
}

func (s *stubPaymentService) GetPayment(userID uint, reference string) (*dto.PaymentResponse, error) {
	return s.getFn(userID, reference)
}

func (s *stubPaymentService) ProcessWebhook(body []byte, signature string) (*dto.PaymentResponse, error) {
	return s.processWebhookFn(body, signature)
}

func (s *stubPaymentService) SimulateWebhook(reference string, status string) (*dto.PaymentResponse, error) {
	return s.simulateWebhookFn(reference, status)
}

type stubProductService struct {
	createFn      func(merchantID uint, req *dto.CreateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error)
	getProductsFn func(query *dto.PaginationQuery, filters *query.ProductFilters) (*dto.PaginatedResponse, error)
	getProductFn  func(id uint) (*dto.ProductResponse, error)
	updateFn      func(merchantID, id uint, req *dto.UpdateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error)
	deleteFn      func(merchantID, id uint) error
	getMerchantFn func(merchantID, id uint) (*dto.ProductResponse, error)
}

func (s *stubProductService) CreateProduct(merchantID uint, req *dto.CreateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error) {
	return s.createFn(merchantID, req, images)
}

func (s *stubProductService) GetProducts(query *dto.PaginationQuery, filters *query.ProductFilters) (*dto.PaginatedResponse, error) {
	return s.getProductsFn(query, filters)
}

func (s *stubProductService) GetProduct(id uint) (*dto.ProductResponse, error) {
	return s.getProductFn(id)
}

func (s *stubProductService) UpdateProduct(merchantID, id uint, req *dto.UpdateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error) {
	return s.updateFn(merchantID, id, req, images)
}

func (s *stubProductService) DeleteProduct(merchantID, id uint) error {
	return s.deleteFn(merchantID, id)
}

func (s *stubProductService) GetMerchantProduct(merchantID, id uint) (*dto.ProductResponse, error) {
	return s.getMerchantFn(merchantID, id)
}

type stubUserService struct {
	getMeFn func(userID uint) (*dto.UserResponse, error)
}

func (s *stubUserService) GetMe(userID uint) (*dto.UserResponse, error) {
	return s.getMeFn(userID)
}
