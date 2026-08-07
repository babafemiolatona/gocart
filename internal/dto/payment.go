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

// PaymentWebhookEvent is the payload a payment provider sends to the webhook
// endpoint. Amount is in minor units (cents), matching models.Payment.Amount.
type PaymentWebhookEvent struct {
	Reference string `json:"reference"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	Timestamp string `json:"timestamp"`
}

// SimulatePaymentRequest is the body for the dev-only endpoint that mimics a
// payment provider callback.
type SimulatePaymentRequest struct {
	Reference string `json:"reference"`
	Status    string `json:"status"`
}
