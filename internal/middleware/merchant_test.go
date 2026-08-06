package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"

	"gocart/internal/models"
	"gocart/internal/repositories"
)

func TestRequireMerchantSuccess(t *testing.T) {
	repo := &stubMerchantRepo{
		getByUserFn: func(userID uint) (*models.Merchant, error) {
			return &models.Merchant{ID: 9, UserID: userID}, nil
		},
	}

	r := newTestRouterWithMerchantCapture(func(c *gin.Context) {
		c.Set("userID", uint(7))
		c.Next()
	}, RequireMerchant(repo))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusOK)

	if capturedMerchantID != 9 {
		t.Errorf("expected merchantID 9, got %d", capturedMerchantID)
	}
}

func TestRequireMerchantMissingUserID(t *testing.T) {
	r, _ := newTestRouter(RequireMerchant(&stubMerchantRepo{}))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "unauthorized" {
		t.Errorf("expected unauthorized, got %q", code)
	}
}

func TestRequireMerchantWrongTypeUserID(t *testing.T) {
	r, _ := newTestRouter(func(c *gin.Context) {
		c.Set("userID", "not-a-uint")
		c.Next()
	}, RequireMerchant(&stubMerchantRepo{}))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestRequireMerchantNotFound(t *testing.T) {
	r, called := newTestRouter(func(c *gin.Context) {
		c.Set("userID", uint(7))
		c.Next()
	}, RequireMerchant(&stubMerchantRepo{}))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusForbidden)

	code, _ := decodeError(t, w)
	if code != "merchant_required" {
		t.Errorf("expected merchant_required, got %q", code)
	}
	if *called {
		t.Error("handler should not run without a merchant")
	}
}

func TestRequireMerchantRepoError(t *testing.T) {
	repo := &stubMerchantRepo{
		getByUserFn: func(userID uint) (*models.Merchant, error) {
			return nil, errBoom
		},
	}
	r, _ := newTestRouter(func(c *gin.Context) {
		c.Set("userID", uint(7))
		c.Next()
	}, RequireMerchant(repo))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusInternalServerError)

	code, _ := decodeError(t, w)
	if code != "verify_merchant_failed" {
		t.Errorf("expected verify_merchant_failed, got %q", code)
	}
}

func TestRequireMerchantGormNotFound(t *testing.T) {
	repo := &stubMerchantRepo{
		getByUserFn: func(userID uint) (*models.Merchant, error) {
			return nil, repositories.ErrRecordNotFound
		},
	}
	r, _ := newTestRouter(func(c *gin.Context) {
		c.Set("userID", uint(7))
		c.Next()
	}, RequireMerchant(repo))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusForbidden)

	code, _ := decodeError(t, w)
	if code != "merchant_required" {
		t.Errorf("expected merchant_required, got %q", code)
	}
}
