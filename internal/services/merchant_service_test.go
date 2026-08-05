package services

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

func newTestMerchantService(
	scope *fakeScope,
	merchantRepo *stubMerchantRepo,
	orderRepo *stubOrderRepo,
	productRepo *stubProductRepo,
) *MerchantService {
	if scope == nil {
		scope = &fakeScope{merchant: merchantRepo, order: orderRepo, product: productRepo}
	}
	return NewMerchantService(&fakeTxManager{scope: scope}, merchantRepo, orderRepo, productRepo)
}

func TestRegisterMerchantSuccess(t *testing.T) {
	var created *models.Merchant
	scope := &fakeScope{
		auth: &stubAuthRepo{
			getByIDFn: func(id uint) (*models.User, error) {
				return &models.User{ID: 1, Email: "chris@example.com"}, nil
			},
		},
		merchant: &stubMerchantRepo{
			createFn: func(m *models.Merchant) error { created = m; return nil },
		},
	}

	svc := newTestMerchantService(scope, &stubMerchantRepo{}, &stubOrderRepo{}, &stubProductRepo{})

	resp, err := svc.RegisterMerchant(1, &dto.MerchantRegisterRequest{
		BusinessName: "Acme Inc",
		Description:  "Sells things",
		Phone:        "555-0100",
		LogoURL:      "/logo.png",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created == nil || created.UserID != 1 || created.BusinessName != "Acme Inc" {
		t.Errorf("merchant not created correctly: %+v", created)
	}
	if created.IsVerified {
		t.Errorf("expected unverified merchant, got %+v", created)
	}
	if resp == nil || resp.BusinessName != "Acme Inc" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestRegisterMerchantUserNotFound(t *testing.T) {
	scope := &fakeScope{
		auth:     &stubAuthRepo{},
		merchant: &stubMerchantRepo{},
	}
	svc := newTestMerchantService(scope, &stubMerchantRepo{}, &stubOrderRepo{}, &stubProductRepo{})

	_, err := svc.RegisterMerchant(99, &dto.MerchantRegisterRequest{BusinessName: "Acme Inc"})
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeUserNotFound)
}

func TestRegisterMerchantDuplicate(t *testing.T) {
	scope := &fakeScope{
		auth: &stubAuthRepo{
			getByIDFn: func(id uint) (*models.User, error) {
				return &models.User{ID: 1}, nil
			},
		},
		merchant: &stubMerchantRepo{
			createFn: func(m *models.Merchant) error { return repositories.ErrDuplicate },
		},
	}
	svc := newTestMerchantService(scope, &stubMerchantRepo{}, &stubOrderRepo{}, &stubProductRepo{})

	_, err := svc.RegisterMerchant(1, &dto.MerchantRegisterRequest{BusinessName: "Acme Inc"})
	assertAppError(t, err, http.StatusConflict, apperrors.CodeMerchantExists)
}

func TestGetProfileNotFound(t *testing.T) {
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, &stubOrderRepo{}, &stubProductRepo{})

	_, err := svc.GetProfile(99)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeMerchantNotFound)
}

func TestGetProfileSuccess(t *testing.T) {
	merchantRepo := &stubMerchantRepo{
		getByIDFn: func(id uint) (*models.Merchant, error) {
			return &models.Merchant{ID: 5, BusinessName: "Acme Inc"}, nil
		},
	}
	svc := newTestMerchantService(nil, merchantRepo, &stubOrderRepo{}, &stubProductRepo{})

	resp, err := svc.GetProfile(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != 5 || resp.BusinessName != "Acme Inc" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestUpdateProfileSuccess(t *testing.T) {
	merchant := &models.Merchant{ID: 5, BusinessName: "Old Name"}
	merchantRepo := &stubMerchantRepo{
		getByIDFn: func(id uint) (*models.Merchant, error) {
			return merchant, nil
		},
		updateFn: func(m *models.Merchant) error { return nil },
	}
	svc := newTestMerchantService(nil, merchantRepo, &stubOrderRepo{}, &stubProductRepo{})

	name := "New Name"
	desc := "New description"
	resp, err := svc.UpdateProfile(5, &dto.UpdateMerchantRequest{BusinessName: &name, Description: &desc})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merchant.BusinessName != "New Name" || merchant.Description != "New description" {
		t.Errorf("merchant not updated: %+v", merchant)
	}
	if resp.BusinessName != "New Name" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestUpdateProfileNotFound(t *testing.T) {
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, &stubOrderRepo{}, &stubProductRepo{})

	_, err := svc.UpdateProfile(99, &dto.UpdateMerchantRequest{})
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeMerchantNotFound)
}

func TestGetOrders(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getByMerchantFn: func(merchantID uint, p *dto.PaginationQuery) ([]models.Order, int64, error) {
			return []models.Order{{ID: 1, Status: models.OrderStatusConfirmed}}, 5, nil
		},
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, &stubProductRepo{})

	resp, err := svc.GetOrders(1, &dto.PaginationQuery{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 5 || resp.PageSize != 2 || resp.TotalPage != 3 {
		t.Errorf("unexpected pagination: %+v", resp)
	}
	if data, ok := resp.Data.([]dto.MerchantOrderResponse); !ok || len(data) != 1 {
		t.Errorf("expected 1 order, got %T: %+v", resp.Data, resp.Data)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, &stubOrderRepo{}, &stubProductRepo{})

	_, err := svc.GetOrder(1, 99)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeOrderNotFound)
}

func TestGetOrderSuccess(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getMerchantByIDFn: func(merchantID uint, orderID uint) (*models.Order, error) {
			return &models.Order{ID: 1, Status: models.OrderStatusConfirmed}, nil
		},
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, &stubProductRepo{})

	resp, err := svc.GetOrder(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != 1 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestUpdateOrderStatusShippedFromConfirmed(t *testing.T) {
	var updated models.OrderStatus
	orderRepo := &stubOrderRepo{
		getMerchantByIDFn: func(merchantID uint, orderID uint) (*models.Order, error) {
			return &models.Order{ID: 1, Status: models.OrderStatusConfirmed}, nil
		},
		updateStatusFn: func(orderID uint, status models.OrderStatus) error {
			updated = status
			return nil
		},
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, &stubProductRepo{})

	err := svc.UpdateOrderStatus(1, 1, &dto.UpdateOrderStatusRequest{Status: models.OrderStatusShipped})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != models.OrderStatusShipped {
		t.Errorf("expected shipped, got %q", updated)
	}
}

func TestUpdateOrderStatusInvalidTransition(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getMerchantByIDFn: func(merchantID uint, orderID uint) (*models.Order, error) {
			return &models.Order{ID: 1, Status: models.OrderStatusPending}, nil
		},
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, &stubProductRepo{})

	err := svc.UpdateOrderStatus(1, 1, &dto.UpdateOrderStatusRequest{Status: models.OrderStatusShipped})
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeInvalidOrderStatus)
}

func TestUpdateOrderStatusUnknownStatus(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getMerchantByIDFn: func(merchantID uint, orderID uint) (*models.Order, error) {
			return &models.Order{ID: 1, Status: models.OrderStatusConfirmed}, nil
		},
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, &stubProductRepo{})

	err := svc.UpdateOrderStatus(1, 1, &dto.UpdateOrderStatusRequest{Status: models.OrderStatusCancelled})
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeInvalidOrderStatus)
}

func TestGetDashboard(t *testing.T) {
	orderRepo := &stubOrderRepo{
		countByMerchantFn: func(merchantID uint) (int64, error) { return 10, nil },
		countByStatusFn: func(merchantID uint, status models.OrderStatus) (int64, error) {
			if status == models.OrderStatusConfirmed {
				return 3, nil
			}
			return 5, nil
		},
		sumRevenueFn: func(merchantID uint) (int64, error) { return 25000, nil },
		getRecentFn: func(merchantID uint, limit int) ([]models.Order, error) {
			return []models.Order{
				{ID: 1, User: models.User{Email: "buyer@example.com"}, Status: models.OrderStatusConfirmed, Total: 1000, Items: []models.OrderItem{{}, {}}},
			}, nil
		},
	}
	productRepo := &stubProductRepo{
		countByMerchantFn: func(merchantID uint) (int64, error) { return 7, nil },
		countLowStockFn:   func(merchantID uint, threshold int) (int64, error) { return 2, nil },
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, productRepo)

	resp, err := svc.GetDashboard(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TotalProducts != 7 || resp.TotalOrders != 10 || resp.AwaitingShipment != 3 || resp.CompletedOrders != 5 {
		t.Errorf("unexpected counts: %+v", resp)
	}
	if resp.LowStockProducts != 2 {
		t.Errorf("expected low stock 2, got %d", resp.LowStockProducts)
	}
	if resp.TotalRevenue != 250 {
		t.Errorf("expected revenue 250, got %v", resp.TotalRevenue)
	}
	if len(resp.RecentOrders) != 1 || resp.RecentOrders[0].Customer != "buyer@example.com" || resp.RecentOrders[0].ItemCount != 2 {
		t.Errorf("unexpected recent orders: %+v", resp.RecentOrders)
	}
}

func TestGetDashboardRepoError(t *testing.T) {
	productRepo := &stubProductRepo{
		countByMerchantFn: func(merchantID uint) (int64, error) { return 0, errBoom },
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, &stubOrderRepo{}, productRepo)

	_, err := svc.GetDashboard(1)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchDashboard)
}

func TestUpdateOrderStatusDeliveredFromShipped(t *testing.T) {
	var updated models.OrderStatus
	orderRepo := &stubOrderRepo{
		getMerchantByIDFn: func(merchantID uint, orderID uint) (*models.Order, error) {
			return &models.Order{ID: 1, Status: models.OrderStatusShipped}, nil
		},
		updateStatusFn: func(orderID uint, status models.OrderStatus) error {
			updated = status
			return nil
		},
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, &stubProductRepo{})

	err := svc.UpdateOrderStatus(1, 1, &dto.UpdateOrderStatusRequest{Status: models.OrderStatusDelivered})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != models.OrderStatusDelivered {
		t.Errorf("expected delivered, got %q", updated)
	}
}

func TestUpdateOrderStatusDeliveredFromPending(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getMerchantByIDFn: func(merchantID uint, orderID uint) (*models.Order, error) {
			return &models.Order{ID: 1, Status: models.OrderStatusPending}, nil
		},
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, &stubProductRepo{})

	err := svc.UpdateOrderStatus(1, 1, &dto.UpdateOrderStatusRequest{Status: models.OrderStatusDelivered})
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeInvalidOrderStatus)
}

func TestUpdateOrderStatusFetchFails(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getMerchantByIDFn: func(merchantID uint, orderID uint) (*models.Order, error) {
			return nil, errBoom
		},
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, orderRepo, &stubProductRepo{})

	err := svc.UpdateOrderStatus(1, 1, &dto.UpdateOrderStatusRequest{Status: models.OrderStatusShipped})
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchOrder)
}

func TestGetDashboardLowStockError(t *testing.T) {
	productRepo := &stubProductRepo{
		countByMerchantFn: func(merchantID uint) (int64, error) { return 1, nil },
		countLowStockFn:   func(merchantID uint, threshold int) (int64, error) { return 0, errBoom },
	}
	svc := newTestMerchantService(nil, &stubMerchantRepo{}, &stubOrderRepo{}, productRepo)

	_, err := svc.GetDashboard(1)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchDashboard)
}
