package dto

type CheckoutResponse struct {
	Order   OrderCheckoutResponse   `json:"order"`
	Payment PaymentCheckoutResponse `json:"payment"`
}
