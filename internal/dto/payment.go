package dto

import "time"

type PaymentResponse struct {
	ID        uint      `json:"id"`
	OrderID   uint      `json:"order_id"`
	Reference string    `json:"reference"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
