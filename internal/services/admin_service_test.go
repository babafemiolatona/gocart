package services

import (
	"net/http"
	"testing"

	apperrors "gocart/internal/errors"
	"gocart/internal/models"
)

func newTestAdminService(
	userRepo *stubUserRepo,
	merchantRepo *stubMerchantRepo,
	productRepo *stubProductRepo,
	orderRepo *stubOrderRepo,
) *AdminService {
	return NewAdminService(userRepo, merchantRepo, productRepo, orderRepo)
}

func TestAdminGetDashboardSuccess(t *testing.T) {
	userRepo := &stubUserRepo{
		countAllFn: func() (int64, error) { return 12, nil },
	}
	merchantRepo := &stubMerchantRepo{}
	productRepo := &stubProductRepo{}
	orderRepo := &stubOrderRepo{
		countAllFn: func() (int64, error) { return 40, nil },
		sumRevenueAllFn: func() (int64, error) {
			return 25000, nil
		},
		countsByStatusFn: func() (map[models.OrderStatus]int64, error) {
			return map[models.OrderStatus]int64{
				models.OrderStatusDelivered: 10,
				models.OrderStatusPending:   30,
			}, nil
		},
	}

	svc := newTestAdminService(userRepo, merchantRepo, productRepo, orderRepo)

	resp, err := svc.GetDashboard()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TotalUsers != 12 || resp.TotalMerchants != 0 || resp.TotalProducts != 0 || resp.TotalOrders != 40 {
		t.Errorf("unexpected counts: %+v", resp)
	}
	if resp.TotalRevenue != 250 {
		t.Errorf("expected revenue 250, got %v", resp.TotalRevenue)
	}
	if resp.OrdersByStatus[models.OrderStatusDelivered] != 10 || resp.OrdersByStatus[models.OrderStatusPending] != 30 {
		t.Errorf("unexpected orders by status: %+v", resp.OrdersByStatus)
	}
}

func TestAdminGetDashboardUserRepoError(t *testing.T) {
	userRepo := &stubUserRepo{
		countAllFn: func() (int64, error) { return 0, errBoom },
	}
	svc := newTestAdminService(userRepo, &stubMerchantRepo{}, &stubProductRepo{}, &stubOrderRepo{})

	_, err := svc.GetDashboard()
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchAdminDashboard)
}

func TestAdminGetDashboardMerchantRepoError(t *testing.T) {
	userRepo := &stubUserRepo{
		countAllFn: func() (int64, error) { return 1, nil },
	}
	merchantRepo := &stubMerchantRepo{
		countAllFn: func() (int64, error) { return 0, errBoom },
	}
	orderRepo := &stubOrderRepo{}

	svc := newTestAdminService(userRepo, merchantRepo, &stubProductRepo{}, orderRepo)

	_, err := svc.GetDashboard()
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchAdminDashboard)
}

func TestAdminGetDashboardOrderRepoError(t *testing.T) {
	userRepo := &stubUserRepo{
		countAllFn: func() (int64, error) { return 1, nil },
	}
	orderRepo := &stubOrderRepo{
		countAllFn: func() (int64, error) { return 0, errBoom },
	}

	svc := newTestAdminService(userRepo, &stubMerchantRepo{}, &stubProductRepo{}, orderRepo)

	_, err := svc.GetDashboard()
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchAdminDashboard)
}
