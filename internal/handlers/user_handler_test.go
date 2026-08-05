package handlers

import (
	"net/http"
	"testing"

	"gocart/internal/dto"

	"github.com/gin-gonic/gin"
)

func TestGetMeSuccess(t *testing.T) {
	svc := &stubUserService{
		getMeFn: func(userID uint) (*dto.UserResponse, error) {
			return &dto.UserResponse{ID: 7, Username: "chris", Email: "chris@example.com"}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/users/me", NewUserHandler(svc).GetMe)

	w := doRequest(t, r, http.MethodGet, "/users/me", "")
	assertStatus(t, w, http.StatusOK)

	var resp dto.UserResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.ID != 7 || resp.Username != "chris" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetMeUnauthorized(t *testing.T) {
	svc := &stubUserService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/users/me", NewUserHandler(svc).GetMe)

	w := doRequest(t, r, http.MethodGet, "/users/me", "")
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "unauthorized" {
		t.Errorf("expected unauthorized, got %q", code)
	}
}

func TestGetMeNotFound(t *testing.T) {
	svc := &stubUserService{
		getMeFn: func(userID uint) (*dto.UserResponse, error) {
			return nil, appErr(http.StatusNotFound, "user_not_found", "user not found")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/users/me", NewUserHandler(svc).GetMe)

	w := doRequest(t, r, http.MethodGet, "/users/me", "")
	assertStatus(t, w, http.StatusNotFound)
}

func TestHealthCheck(t *testing.T) {
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/health", HealthCheck)

	w := doRequest(t, r, http.MethodGet, "/health", "")
	assertStatus(t, w, http.StatusOK)

	var resp struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.Status != "OK" || resp.Version != "1.0.0" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetMeWrongTypeInContext(t *testing.T) {
	svc := &stubUserService{}
	r := newTestRouter(0, nil)
	r.Use(func(c *gin.Context) {
		c.Set("userID", "not-a-uint")
		c.Next()
	})
	registerHandler(r, http.MethodGet, "/users/me", NewUserHandler(svc).GetMe)

	w := doRequest(t, r, http.MethodGet, "/users/me", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetMerchantIDWrongType(t *testing.T) {
	svc := &stubMerchantService{}
	r := newTestRouter(7, nil)
	r.Use(func(c *gin.Context) {
		c.Set("merchantID", "not-a-uint")
		c.Next()
	})
	registerHandler(r, http.MethodGet, "/merchants/me", NewMerchantHandler(svc).GetMe)

	w := doRequest(t, r, http.MethodGet, "/merchants/me", "")
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "merchant_required" {
		t.Errorf("expected merchant_required, got %q", code)
	}
}
