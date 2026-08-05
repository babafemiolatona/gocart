package handlers

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
)

func TestCheckoutSuccess(t *testing.T) {
	var gotAddress, gotKey string
	svc := &stubOrderService{
		checkoutFn: func(userID uint, shippingAddress, idempotencyKey string) (*dto.CheckoutResponse, error) {
			gotAddress, gotKey = shippingAddress, idempotencyKey
			return &dto.CheckoutResponse{Order: &dto.OrderCheckoutResponse{ID: 1, Status: "pending", Total: 19.99}}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/orders/checkout", NewOrderHandler(svc).Checkout)

	w := doRequest(t, r, http.MethodPost, "/orders/checkout", `{"shipping_address":"1 Main St","idempotency_key":"key-123"}`)
	assertStatus(t, w, http.StatusCreated)

	var resp dto.CheckoutResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.Order == nil || resp.Order.ID != 1 || resp.Order.Total != 19.99 {
		t.Errorf("unexpected response: %+v", resp)
	}
	if gotAddress != "1 Main St" || gotKey != "key-123" {
		t.Errorf("unexpected service args: address=%q key=%q", gotAddress, gotKey)
	}
}

func TestCheckoutValidationError(t *testing.T) {
	svc := &stubOrderService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/orders/checkout", NewOrderHandler(svc).Checkout)

	w := doRequest(t, r, http.MethodPost, "/orders/checkout", `{}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetMyOrdersSuccess(t *testing.T) {
	svc := &stubOrderService{
		getUserOrders: func(userID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error) {
			if p == nil || p.Page != 1 {
				t.Errorf("unexpected pagination: %+v", p)
			}
			return &dto.PaginatedResponse{Total: 1, Page: 1, PageSize: 10, TotalPage: 1}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/orders", NewOrderHandler(svc).GetMyOrders)

	w := doRequest(t, r, http.MethodGet, "/orders?page=1", "")
	assertStatus(t, w, http.StatusOK)
}

func TestGetOrderSuccess(t *testing.T) {
	svc := &stubOrderService{
		getOrderFn: func(userID, orderID uint) (*dto.OrderDetailsResponse, error) {
			if userID != 7 || orderID != 5 {
				t.Errorf("unexpected args: userID=%d orderID=%d", userID, orderID)
			}
			return &dto.OrderDetailsResponse{ID: 5, Status: "pending"}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/orders/:id", NewOrderHandler(svc).GetOrder)

	w := doRequest(t, r, http.MethodGet, "/orders/5", "")
	assertStatus(t, w, http.StatusOK)
}

func TestGetOrderInvalidID(t *testing.T) {
	svc := &stubOrderService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/orders/:id", NewOrderHandler(svc).GetOrder)

	w := doRequest(t, r, http.MethodGet, "/orders/abc", "")
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "invalid_order_id" {
		t.Errorf("expected invalid_order_id, got %q", code)
	}
}

func TestCancelOrderSuccess(t *testing.T) {
	var cancelled uint
	svc := &stubOrderService{
		cancelFn: func(userID, orderID uint) error {
			cancelled = orderID
			return nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPut, "/orders/:id/cancel", NewOrderHandler(svc).CancelOrder)

	w := doRequest(t, r, http.MethodPut, "/orders/5/cancel", "")
	assertStatus(t, w, http.StatusOK)
	if cancelled != 5 {
		t.Errorf("expected to cancel order 5, got %d", cancelled)
	}
}

func TestCancelOrderServiceError(t *testing.T) {
	svc := &stubOrderService{
		cancelFn: func(userID, orderID uint) error {
			return appErr(http.StatusConflict, "invalid_order_status", "cannot cancel")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPut, "/orders/:id/cancel", NewOrderHandler(svc).CancelOrder)

	w := doRequest(t, r, http.MethodPut, "/orders/5/cancel", "")
	assertStatus(t, w, http.StatusConflict)

	code, _ := decodeError(t, w)
	if code != "invalid_order_status" {
		t.Errorf("unexpected error code: %q", code)
	}
}

func TestCheckoutUnauthorized(t *testing.T) {
	svc := &stubOrderService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/orders/checkout", NewOrderHandler(svc).Checkout)

	w := doRequest(t, r, http.MethodPost, "/orders/checkout", `{"shipping_address":"1 Main St"}`)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestCheckoutServiceError(t *testing.T) {
	svc := &stubOrderService{
		checkoutFn: func(userID uint, shippingAddress, idempotencyKey string) (*dto.CheckoutResponse, error) {
			return nil, appErr(http.StatusConflict, "checkout_conflict", "conflict")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/orders/checkout", NewOrderHandler(svc).Checkout)

	w := doRequest(t, r, http.MethodPost, "/orders/checkout", `{"shipping_address":"1 Main St"}`)
	assertStatus(t, w, http.StatusConflict)
}

func TestGetMyOrdersUnauthorized(t *testing.T) {
	svc := &stubOrderService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/orders", NewOrderHandler(svc).GetMyOrders)

	w := doRequest(t, r, http.MethodGet, "/orders", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetMyOrdersServiceError(t *testing.T) {
	svc := &stubOrderService{
		getUserOrders: func(userID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error) {
			return nil, appErr(http.StatusInternalServerError, "fetch_orders_failed", "failed")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/orders", NewOrderHandler(svc).GetMyOrders)

	w := doRequest(t, r, http.MethodGet, "/orders", "")
	assertStatus(t, w, http.StatusInternalServerError)
}

func TestGetOrderUnauthorized(t *testing.T) {
	svc := &stubOrderService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/orders/:id", NewOrderHandler(svc).GetOrder)

	w := doRequest(t, r, http.MethodGet, "/orders/5", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetOrderNotFound(t *testing.T) {
	svc := &stubOrderService{
		getOrderFn: func(userID, orderID uint) (*dto.OrderDetailsResponse, error) {
			return nil, appErr(http.StatusNotFound, "order_not_found", "not found")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/orders/:id", NewOrderHandler(svc).GetOrder)

	w := doRequest(t, r, http.MethodGet, "/orders/99", "")
	assertStatus(t, w, http.StatusNotFound)
}

func TestCancelOrderInvalidID(t *testing.T) {
	svc := &stubOrderService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPut, "/orders/:id/cancel", NewOrderHandler(svc).CancelOrder)

	w := doRequest(t, r, http.MethodPut, "/orders/abc/cancel", "")
	assertStatus(t, w, http.StatusBadRequest)
}
