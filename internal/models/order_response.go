package models

import "time"

type OrderResponse struct {
	ID              uint      `json:"id"`
	Status          string    `json:"status"`
	Total           float64   `json:"total"`
	ShippingAddress string    `json:"shipping_address"`
	CreatedAt       time.Time `json:"created_at"`
}
