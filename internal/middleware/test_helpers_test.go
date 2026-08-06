package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gocart/internal/logger"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

var errBoom = errors.New("boom")

var captureUser uint
var captureRole models.Role
var capturedMerchantID uint

func init() {
	logger.Log = zerolog.Nop()
}

// newTestRouterWithCapture builds a router like newTestRouter but the terminal
// handler records the context userID/userRole set by middleware.
func newTestRouterWithCapture(handlers ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ErrorHandler())
	r.Use(handlers...)
	r.GET("/test", func(c *gin.Context) {
		if v, ok := c.Get("userID"); ok {
			captureUser, _ = v.(uint)
		}
		if v, ok := c.Get("userRole"); ok {
			captureRole, _ = v.(models.Role)
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r
}

func capturedContext() (uint, models.Role) {
	return captureUser, captureRole
}

// newTestRouterWithMerchantCapture builds a router whose terminal handler
// records the merchantID set by RequireMerchant middleware.
func newTestRouterWithMerchantCapture(handlers ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ErrorHandler())
	r.Use(handlers...)
	r.GET("/test", func(c *gin.Context) {
		if v, ok := c.Get("merchantID"); ok {
			capturedMerchantID, _ = v.(uint)
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r
}

// newTestRouter builds a router with the real ErrorHandler middleware so error
// responses render like production. `handlers` are applied in order; a stub
// terminal handler records whether it ran.
func newTestRouter(handlers ...gin.HandlerFunc) (*gin.Engine, *bool) {
	gin.SetMode(gin.TestMode)

	var called bool

	r := gin.New()
	r.Use(ErrorHandler())
	r.Use(handlers...)
	r.GET("/test", func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r, &called
}

func doRequest(r *gin.Engine, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) (code, message string) {
	t.Helper()

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	return resp.Error.Code, resp.Error.Message
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("expected status %d, got %d (body: %s)", want, w.Code, w.Body.String())
	}
}

// stubTokenVerifier implements TokenVerifier for tests.
type stubTokenVerifier struct {
	verifyFn func(tokenStr string) (*services.CustomClaims, error)
}

func (s *stubTokenVerifier) VerifyToken(tokenStr string) (*services.CustomClaims, error) {
	return s.verifyFn(tokenStr)
}

// stubMerchantRepo implements repositories.MerchantRepository for tests.
type stubMerchantRepo struct {
	getByUserFn func(userID uint) (*models.Merchant, error)
}

func (s *stubMerchantRepo) Create(merchant *models.Merchant) error { return nil }

func (s *stubMerchantRepo) GetByID(id uint) (*models.Merchant, error) {
	return nil, repositories.ErrRecordNotFound
}

func (s *stubMerchantRepo) GetByUserID(userID uint) (*models.Merchant, error) {
	if s.getByUserFn != nil {
		return s.getByUserFn(userID)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubMerchantRepo) Update(merchant *models.Merchant) error { return nil }

func claimsWith(subject string, role models.Role) *services.CustomClaims {
	return &services.CustomClaims{
		Role:             role,
		RegisteredClaims: jwt.RegisteredClaims{Subject: subject},
	}
}
