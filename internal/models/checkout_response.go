package models

type CheckoutResponse struct {
	Order   *Order   `json:"order"`
	Payment *Payment `json:"payment"`
}
