package handlers

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
)

func TestRegisterMerchantSuccess(t *testing.T) {
	var gotUser uint
	svc := &stubMerchantService{
		registerFn: func(userID uint, req *dto.MerchantRegisterRequest) (*dto.MerchantResponse, error) {
			gotUser = userID
			return &dto.MerchantResponse{ID: 2, BusinessName: req.BusinessName}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/merchants/register", NewMerchantHandler(svc).RegisterMerchant)

	w := doRequest(t, r, http.MethodPost, "/merchants/register", `{"business_name":"Acme Inc"}`)
	assertStatus(t, w, http.StatusCreated)

	var resp dto.MerchantResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.ID != 2 || resp.BusinessName != "Acme Inc" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if gotUser != 7 {
		t.Errorf("expected user 7, got %d", gotUser)
	}
}

func TestRegisterMerchantValidationError(t *testing.T) {
	svc := &stubMerchantService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/merchants/register", NewMerchantHandler(svc).RegisterMerchant)

	w := doRequest(t, r, http.MethodPost, "/merchants/register", `{}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetMeMerchant(t *testing.T) {
	svc := &stubMerchantService{
		getProfileFn: func(merchantID uint) (*dto.MerchantResponse, error) {
			return &dto.MerchantResponse{ID: merchantID, BusinessName: "Acme Inc"}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/me", NewMerchantHandler(svc).GetMe)

	w := doRequest(t, r, http.MethodGet, "/merchants/me", "")
	assertStatus(t, w, http.StatusOK)
}

func TestGetMeMerchantUnauthorized(t *testing.T) {
	svc := &stubMerchantService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/merchants/me", NewMerchantHandler(svc).GetMe)

	w := doRequest(t, r, http.MethodGet, "/merchants/me", "")
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "merchant_required" {
		t.Errorf("expected merchant_required, got %q", code)
	}
}

func TestUpdateMeSuccess(t *testing.T) {
	svc := &stubMerchantService{
		updateProfileFn: func(merchantID uint, req *dto.UpdateMerchantRequest) (*dto.MerchantResponse, error) {
			return &dto.MerchantResponse{ID: merchantID}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPut, "/merchants/me", NewMerchantHandler(svc).UpdateMe)

	w := doRequest(t, r, http.MethodPut, "/merchants/me", `{"business_name":"New Co"}`)
	assertStatus(t, w, http.StatusOK)
}

func TestGetOrdersMerchant(t *testing.T) {
	svc := &stubMerchantService{
		getOrdersFn: func(merchantID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error) {
			if merchantID != 2 {
				t.Errorf("expected merchant 2, got %d", merchantID)
			}
			return &dto.PaginatedResponse{Total: 1, Page: 1, PageSize: 10, TotalPage: 1}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/orders", NewMerchantHandler(svc).GetOrders)

	w := doRequest(t, r, http.MethodGet, "/merchants/orders", "")
	assertStatus(t, w, http.StatusOK)
}

func TestGetOrderMerchant(t *testing.T) {
	svc := &stubMerchantService{
		getOrderFn: func(merchantID, orderID uint) (*dto.MerchantOrderResponse, error) {
			if merchantID != 2 || orderID != 5 {
				t.Errorf("unexpected args: merchant=%d order=%d", merchantID, orderID)
			}
			return &dto.MerchantOrderResponse{ID: 5}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/orders/:id", NewMerchantHandler(svc).GetOrder)

	w := doRequest(t, r, http.MethodGet, "/merchants/orders/5", "")
	assertStatus(t, w, http.StatusOK)
}

func TestGetOrderMerchantInvalidID(t *testing.T) {
	svc := &stubMerchantService{}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/orders/:id", NewMerchantHandler(svc).GetOrder)

	w := doRequest(t, r, http.MethodGet, "/merchants/orders/abc", "")
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "invalid_order_id" {
		t.Errorf("expected invalid_order_id, got %q", code)
	}
}

func TestUpdateOrderStatusSuccess(t *testing.T) {
	var gotStatus dto.UpdateOrderStatusRequest
	svc := &stubMerchantService{
		updateStatusFn: func(merchantID, orderID uint, req *dto.UpdateOrderStatusRequest) error {
			gotStatus = *req
			return nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPatch, "/merchants/orders/:id/status", NewMerchantHandler(svc).UpdateOrderStatus)

	w := doRequest(t, r, http.MethodPatch, "/merchants/orders/5/status", `{"status":"shipped"}`)
	assertStatus(t, w, http.StatusOK)
	if gotStatus.Status != "shipped" {
		t.Errorf("unexpected status request: %+v", gotStatus)
	}
}

func TestUpdateOrderStatusInvalidStatus(t *testing.T) {
	svc := &stubMerchantService{}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPatch, "/merchants/orders/:id/status", NewMerchantHandler(svc).UpdateOrderStatus)

	w := doRequest(t, r, http.MethodPatch, "/merchants/orders/5/status", `{"status":""}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetDashboardSuccess(t *testing.T) {
	svc := &stubMerchantService{
		getDashboardFn: func(merchantID uint) (*dto.MerchantDashboardResponse, error) {
			return &dto.MerchantDashboardResponse{TotalProducts: 4, TotalOrders: 10}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/dashboard", NewMerchantHandler(svc).GetDashboard)

	w := doRequest(t, r, http.MethodGet, "/merchants/dashboard", "")
	assertStatus(t, w, http.StatusOK)

	var resp dto.MerchantDashboardResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.TotalProducts != 4 || resp.TotalOrders != 10 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestRegisterMerchantUnauthorized(t *testing.T) {
	svc := &stubMerchantService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/merchants/register", NewMerchantHandler(svc).RegisterMerchant)

	w := doRequest(t, r, http.MethodPost, "/merchants/register", `{"business_name":"Acme Inc"}`)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestRegisterMerchantServiceError(t *testing.T) {
	svc := &stubMerchantService{
		registerFn: func(userID uint, req *dto.MerchantRegisterRequest) (*dto.MerchantResponse, error) {
			return nil, appErr(http.StatusConflict, "merchant_exists", "exists")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/merchants/register", NewMerchantHandler(svc).RegisterMerchant)

	w := doRequest(t, r, http.MethodPost, "/merchants/register", `{"business_name":"Acme Inc"}`)
	assertStatus(t, w, http.StatusConflict)
}

func TestGetMeMerchantNotFound(t *testing.T) {
	svc := &stubMerchantService{
		getProfileFn: func(merchantID uint) (*dto.MerchantResponse, error) {
			return nil, appErr(http.StatusNotFound, "merchant_not_found", "not found")
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/me", NewMerchantHandler(svc).GetMe)

	w := doRequest(t, r, http.MethodGet, "/merchants/me", "")
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdateMeValidationError(t *testing.T) {
	svc := &stubMerchantService{
		updateProfileFn: func(merchantID uint, req *dto.UpdateMerchantRequest) (*dto.MerchantResponse, error) {
			t.Fatal("service should not be called on validation failure")
			return nil, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPut, "/merchants/me", NewMerchantHandler(svc).UpdateMe)

	w := doRequest(t, r, http.MethodPut, "/merchants/me", `{"business_name":123}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpdateMeUnauthorized(t *testing.T) {
	svc := &stubMerchantService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPut, "/merchants/me", NewMerchantHandler(svc).UpdateMe)

	w := doRequest(t, r, http.MethodPut, "/merchants/me", `{"business_name":"New Co"}`)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetOrdersMerchantUnauthorized(t *testing.T) {
	svc := &stubMerchantService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/merchants/orders", NewMerchantHandler(svc).GetOrders)

	w := doRequest(t, r, http.MethodGet, "/merchants/orders", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetOrdersMerchantServiceError(t *testing.T) {
	svc := &stubMerchantService{
		getOrdersFn: func(merchantID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error) {
			return nil, appErr(http.StatusInternalServerError, "fetch_orders_failed", "failed")
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/orders", NewMerchantHandler(svc).GetOrders)

	w := doRequest(t, r, http.MethodGet, "/merchants/orders", "")
	assertStatus(t, w, http.StatusInternalServerError)
}

func TestGetOrderMerchantNotFound(t *testing.T) {
	svc := &stubMerchantService{
		getOrderFn: func(merchantID, orderID uint) (*dto.MerchantOrderResponse, error) {
			return nil, appErr(http.StatusNotFound, "order_not_found", "not found")
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/orders/:id", NewMerchantHandler(svc).GetOrder)

	w := doRequest(t, r, http.MethodGet, "/merchants/orders/99", "")
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdateOrderStatusUnauthorized(t *testing.T) {
	svc := &stubMerchantService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPatch, "/merchants/orders/:id/status", NewMerchantHandler(svc).UpdateOrderStatus)

	w := doRequest(t, r, http.MethodPatch, "/merchants/orders/5/status", `{"status":"shipped"}`)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestUpdateOrderStatusInvalidID(t *testing.T) {
	svc := &stubMerchantService{}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPatch, "/merchants/orders/:id/status", NewMerchantHandler(svc).UpdateOrderStatus)

	w := doRequest(t, r, http.MethodPatch, "/merchants/orders/abc/status", `{"status":"shipped"}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpdateOrderStatusServiceError(t *testing.T) {
	svc := &stubMerchantService{
		updateStatusFn: func(merchantID, orderID uint, req *dto.UpdateOrderStatusRequest) error {
			return appErr(http.StatusBadRequest, "invalid_order_status", "invalid")
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPatch, "/merchants/orders/:id/status", NewMerchantHandler(svc).UpdateOrderStatus)

	w := doRequest(t, r, http.MethodPatch, "/merchants/orders/5/status", `{"status":"shipped"}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetDashboardUnauthorized(t *testing.T) {
	svc := &stubMerchantService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/merchants/dashboard", NewMerchantHandler(svc).GetDashboard)

	w := doRequest(t, r, http.MethodGet, "/merchants/dashboard", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetDashboardServiceError(t *testing.T) {
	svc := &stubMerchantService{
		getDashboardFn: func(merchantID uint) (*dto.MerchantDashboardResponse, error) {
			return nil, appErr(http.StatusInternalServerError, "dashboard_failed", "failed")
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/dashboard", NewMerchantHandler(svc).GetDashboard)

	w := doRequest(t, r, http.MethodGet, "/merchants/dashboard", "")
	assertStatus(t, w, http.StatusInternalServerError)
}
