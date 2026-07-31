package dto

import (
	"gocart/internal/models"
	"time"
)

type MerchantOrderItemResponse struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type MerchantOrderResponse struct {
	ID              uint                        `json:"id"`
	Status          string                      `json:"status"`
	Total           float64                     `json:"total"`
	ShippingAddress string                      `json:"shipping_address"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
	Items           []MerchantOrderItemResponse `json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status" binding:"required"`
}
