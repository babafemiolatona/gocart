package models

import "time"

type OrderDetailsResponse struct {
	ID              uint        `json:"id"`
	Status          string      `json:"status"`
	Total           float64     `json:"total"`
	ShippingAddress string      `json:"shipping_address"`
	Items           []OrderItem `json:"items"`
	CreatedAt       time.Time   `json:"created_at"`
}
