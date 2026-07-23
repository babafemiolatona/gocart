package dto

import "time"

type OrderDetailsResponse struct {
	ID              uint                `json:"id"`
	Status          string              `json:"status"`
	Total           float64             `json:"total"`
	ShippingAddress string              `json:"shipping_address"`
	Items           []OrderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
}
