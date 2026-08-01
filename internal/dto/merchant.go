package dto

import (
	"gocart/internal/models"
	"time"
)

type MerchantRegisterRequest struct {
	BusinessName string `json:"business_name" binding:"required"`
	Description  string `json:"description"`
	Phone        string `json:"phone"`
	LogoURL      string `json:"logo_url"`
}

type UpdateMerchantRequest struct {
	BusinessName *string `json:"business_name" binding:"omitempty,min=3,max=255"`
	Description  *string `json:"description"`
	Phone        *string `json:"phone"`
	LogoURL      *string `json:"logo_url"`
}

type MerchantResponse struct {
	ID           uint      `json:"id"`
	BusinessName string    `json:"business_name"`
	Description  string    `json:"description"`
	Phone        string    `json:"phone"`
	LogoURL      string    `json:"logo_url"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MerchantDashboardResponse struct {
	TotalProducts    int64                         `json:"total_products"`
	TotalOrders      int64                         `json:"total_orders"`
	PendingOrders    int64                         `json:"pending_orders"`
	CompletedOrders  int64                         `json:"completed_orders"`
	TotalRevenue     float64                       `json:"total_revenue"`
	LowStockProducts int64                         `json:"low_stock_products"`
	RecentOrders     []MerchantRecentOrderResponse `json:"recent_orders"`
}

type MerchantRecentOrderResponse struct {
	ID        uint               `json:"id"`
	Customer  string             `json:"customer"`
	Status    models.OrderStatus `json:"status"`
	Total     float64            `json:"total"`
	ItemCount int                `json:"item_count"`
	CreatedAt time.Time          `json:"created_at"`
}
