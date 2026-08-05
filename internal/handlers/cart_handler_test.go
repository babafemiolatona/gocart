package handlers

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
)

func TestGetCartSuccess(t *testing.T) {
	svc := &stubCartService{
		getCartFn: func(userID uint) (*dto.CartResponse, error) {
			return &dto.CartResponse{ID: 5, ItemCount: 2, Total: 29.98}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/cart", NewCartHandler(svc).GetCart)

	w := doRequest(t, r, http.MethodGet, "/cart", "")
	assertStatus(t, w, http.StatusOK)

	var resp dto.CartResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.ID != 5 || resp.ItemCount != 2 || resp.Total != 29.98 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetCartUnauthorized(t *testing.T) {
	svc := &stubCartService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/cart", NewCartHandler(svc).GetCart)

	w := doRequest(t, r, http.MethodGet, "/cart", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetCartServiceError(t *testing.T) {
	svc := &stubCartService{
		getCartFn: func(userID uint) (*dto.CartResponse, error) {
			return nil, appErr(http.StatusInternalServerError, "fetch_cart_failed", "failed to fetch cart")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/cart", NewCartHandler(svc).GetCart)

	w := doRequest(t, r, http.MethodGet, "/cart", "")
	assertStatus(t, w, http.StatusInternalServerError)

	code, _ := decodeError(t, w)
	if code != "fetch_cart_failed" {
		t.Errorf("unexpected error code: %q", code)
	}
}

func TestAddToCartSuccess(t *testing.T) {
	svc := &stubCartService{
		addToCartFn: func(userID uint, req *dto.AddToCartRequest) (*dto.CartResponse, error) {
			if req.ProductID != 3 || req.Quantity != 2 {
				t.Errorf("unexpected request: %+v", req)
			}
			return &dto.CartResponse{ID: 5, ItemCount: 2}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/cart/items", NewCartHandler(svc).AddToCart)

	w := doRequest(t, r, http.MethodPost, "/cart/items", `{"product_id":3,"quantity":2}`)
	assertStatus(t, w, http.StatusOK)
}

func TestAddToCartValidationError(t *testing.T) {
	svc := &stubCartService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/cart/items", NewCartHandler(svc).AddToCart)

	w := doRequest(t, r, http.MethodPost, "/cart/items", `{"quantity":0}`)
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "validation_error" {
		t.Errorf("expected validation_error, got %q", code)
	}
}

func TestUpdateCartItemSuccess(t *testing.T) {
	svc := &stubCartService{
		updateCartItemFn: func(userID, itemID uint, qty int) (*dto.CartResponse, error) {
			if userID != 7 || itemID != 9 || qty != 3 {
				t.Errorf("unexpected args: userID=%d itemID=%d qty=%d", userID, itemID, qty)
			}
			return &dto.CartResponse{ID: 5, ItemCount: 3}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPut, "/cart/items/:itemID", NewCartHandler(svc).UpdateCartItem)

	w := doRequest(t, r, http.MethodPut, "/cart/items/9", `{"quantity":3}`)
	assertStatus(t, w, http.StatusOK)
}

func TestUpdateCartItemInvalidID(t *testing.T) {
	svc := &stubCartService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPut, "/cart/items/:itemID", NewCartHandler(svc).UpdateCartItem)

	w := doRequest(t, r, http.MethodPut, "/cart/items/abc", `{"quantity":3}`)
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "invalid_cart_item_id" {
		t.Errorf("expected invalid_cart_item_id, got %q", code)
	}
}

func TestRemoveFromCartSuccess(t *testing.T) {
	svc := &stubCartService{
		removeFromCartFn: func(userID, itemID uint) (*dto.CartResponse, error) {
			if userID != 7 || itemID != 9 {
				t.Errorf("unexpected args: userID=%d itemID=%d", userID, itemID)
			}
			return &dto.CartResponse{ID: 5, ItemCount: 1}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodDelete, "/cart/items/:itemID", NewCartHandler(svc).RemoveFromCart)

	w := doRequest(t, r, http.MethodDelete, "/cart/items/9", "")
	assertStatus(t, w, http.StatusOK)
}

func TestRemoveFromCartInvalidID(t *testing.T) {
	svc := &stubCartService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodDelete, "/cart/items/:itemID", NewCartHandler(svc).RemoveFromCart)

	w := doRequest(t, r, http.MethodDelete, "/cart/items/xyz", "")
	assertStatus(t, w, http.StatusBadRequest)
}

func TestClearCartSuccess(t *testing.T) {
	var cleared uint
	svc := &stubCartService{
		clearCartFn: func(userID uint) error {
			cleared = userID
			return nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodDelete, "/cart", NewCartHandler(svc).ClearCart)

	w := doRequest(t, r, http.MethodDelete, "/cart", "")
	assertStatus(t, w, http.StatusNoContent)
	if cleared != 7 {
		t.Errorf("expected clear for user 7, got %d", cleared)
	}
}

func TestAddToCartUnauthorized(t *testing.T) {
	svc := &stubCartService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/cart/items", NewCartHandler(svc).AddToCart)

	w := doRequest(t, r, http.MethodPost, "/cart/items", `{"product_id":3,"quantity":2}`)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestAddToCartServiceError(t *testing.T) {
	svc := &stubCartService{
		addToCartFn: func(userID uint, req *dto.AddToCartRequest) (*dto.CartResponse, error) {
			return nil, appErr(http.StatusConflict, "different_merchant", "conflict")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/cart/items", NewCartHandler(svc).AddToCart)

	w := doRequest(t, r, http.MethodPost, "/cart/items", `{"product_id":3,"quantity":2}`)
	assertStatus(t, w, http.StatusConflict)

	code, _ := decodeError(t, w)
	if code != "different_merchant" {
		t.Errorf("unexpected error code: %q", code)
	}
}

func TestUpdateCartItemUnauthorized(t *testing.T) {
	svc := &stubCartService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPut, "/cart/items/:itemID", NewCartHandler(svc).UpdateCartItem)

	w := doRequest(t, r, http.MethodPut, "/cart/items/9", `{"quantity":3}`)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestUpdateCartItemServiceError(t *testing.T) {
	svc := &stubCartService{
		updateCartItemFn: func(userID, itemID uint, qty int) (*dto.CartResponse, error) {
			return nil, appErr(http.StatusNotFound, "cart_item_not_found", "not found")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPut, "/cart/items/:itemID", NewCartHandler(svc).UpdateCartItem)

	w := doRequest(t, r, http.MethodPut, "/cart/items/9", `{"quantity":3}`)
	assertStatus(t, w, http.StatusNotFound)
}

func TestRemoveFromCartUnauthorized(t *testing.T) {
	svc := &stubCartService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodDelete, "/cart/items/:itemID", NewCartHandler(svc).RemoveFromCart)

	w := doRequest(t, r, http.MethodDelete, "/cart/items/9", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestRemoveFromCartServiceError(t *testing.T) {
	svc := &stubCartService{
		removeFromCartFn: func(userID, itemID uint) (*dto.CartResponse, error) {
			return nil, appErr(http.StatusNotFound, "cart_item_not_found", "not found")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodDelete, "/cart/items/:itemID", NewCartHandler(svc).RemoveFromCart)

	w := doRequest(t, r, http.MethodDelete, "/cart/items/9", "")
	assertStatus(t, w, http.StatusNotFound)
}

func TestClearCartUnauthorized(t *testing.T) {
	svc := &stubCartService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodDelete, "/cart", NewCartHandler(svc).ClearCart)

	w := doRequest(t, r, http.MethodDelete, "/cart", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestClearCartServiceError(t *testing.T) {
	svc := &stubCartService{
		clearCartFn: func(userID uint) error {
			return appErr(http.StatusInternalServerError, "clear_cart_failed", "failed")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodDelete, "/cart", NewCartHandler(svc).ClearCart)

	w := doRequest(t, r, http.MethodDelete, "/cart", "")
	assertStatus(t, w, http.StatusInternalServerError)
}
