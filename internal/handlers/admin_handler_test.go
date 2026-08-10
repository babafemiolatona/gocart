package handlers

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
)

type stubAdminService struct {
	getDashboardFn func() (*dto.AdminDashboardResponse, error)
}

func (s *stubAdminService) GetDashboard() (*dto.AdminDashboardResponse, error) {
	if s.getDashboardFn != nil {
		return s.getDashboardFn()
	}
	return nil, nil
}

func TestAdminHandler_GetDashboard(t *testing.T) {
	r := newTestRouter(0, nil)

	svc := &stubAdminService{
		getDashboardFn: func() (*dto.AdminDashboardResponse, error) {
			return &dto.AdminDashboardResponse{
				TotalUsers:     10,
				TotalMerchants: 2,
				TotalProducts:  15,
				TotalOrders:    40,
				TotalRevenue:   250,
				OrdersByStatus: map[models.OrderStatus]int64{
					models.OrderStatusDelivered: 10,
				},
			}, nil
		},
	}
	registerHandler(r, http.MethodGet, "/dashboard", NewAdminHandler(svc).GetDashboard)

	w := doRequest(t, r, http.MethodGet, "/dashboard", "")
	assertStatus(t, w, http.StatusOK)

	var resp dto.AdminDashboardResponse
	decodeJSON(t, w.Body.Bytes(), &resp)

	if resp.TotalUsers != 10 || resp.TotalOrders != 40 || resp.TotalRevenue != 250 {
		t.Errorf("unexpected dashboard: %+v", resp)
	}
	if resp.OrdersByStatus[models.OrderStatusDelivered] != 10 {
		t.Errorf("unexpected orders by status: %+v", resp.OrdersByStatus)
	}
}

func TestAdminHandler_GetDashboardServiceError(t *testing.T) {
	r := newTestRouter(0, nil)

	svc := &stubAdminService{
		getDashboardFn: func() (*dto.AdminDashboardResponse, error) {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeFetchAdminDashboard,
				"failed to fetch admin dashboard",
				nil,
			)
		},
	}
	registerHandler(r, http.MethodGet, "/dashboard", NewAdminHandler(svc).GetDashboard)

	w := doRequest(t, r, http.MethodGet, "/dashboard", "")
	assertStatus(t, w, http.StatusInternalServerError)

	code, _ := decodeError(t, w)
	if code != "fetch_admin_dashboard_failed" {
		t.Errorf("expected fetch_admin_dashboard_failed, got %q", code)
	}
}
