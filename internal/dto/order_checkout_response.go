package dto

type OrderCheckoutResponse struct {
	ID              uint                `json:"id"`
	Status          string              `json:"status"`
	Total           float64             `json:"total"`
	ShippingAddress string              `json:"shipping_address"`
	Items           []OrderItemResponse `json:"items"`
}
