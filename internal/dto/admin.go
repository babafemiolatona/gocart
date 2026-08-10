package dto

import "gocart/internal/models"

type AdminDashboardResponse struct {
	TotalUsers     int64                        `json:"total_users"`
	TotalMerchants int64                        `json:"total_merchants"`
	TotalProducts  int64                        `json:"total_products"`
	TotalOrders    int64                        `json:"total_orders"`
	TotalRevenue   float64                      `json:"total_revenue"`
	OrdersByStatus map[models.OrderStatus]int64 `json:"orders_by_status"`
}
