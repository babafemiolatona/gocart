package dto

import "time"

type CheckoutResponse struct {
	Order   OrderCheckoutResponse   `json:"order"`
	Payment PaymentCheckoutResponse `json:"payment"`
}

type OrderCheckoutResponse struct {
	ID              uint                `json:"id"`
	Status          string              `json:"status"`
	Total           float64             `json:"total"`
	ShippingAddress string              `json:"shipping_address"`
	Items           []OrderItemResponse `json:"items"`
}

type OrderDetailsResponse struct {
	ID              uint                `json:"id"`
	Status          string              `json:"status"`
	Total           float64             `json:"total"`
	ShippingAddress string              `json:"shipping_address"`
	Items           []OrderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
}

type OrderItemResponse struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type OrderResponse struct {
	ID              uint      `json:"id"`
	Status          string    `json:"status"`
	Total           float64   `json:"total"`
	ShippingAddress string    `json:"shipping_address"`
	CreatedAt       time.Time `json:"created_at"`
}
