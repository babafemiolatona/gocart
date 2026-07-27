package dto

type PaymentCheckoutResponse struct {
	Reference string  `json:"reference"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}
