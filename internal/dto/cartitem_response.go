package dto

type CartItemResponse struct {
	ID       uint                `json:"id"`
	Product  CartProductResponse `json:"product"`
	Quantity int                 `json:"quantity"`
	Price    float64             `json:"price"`
	Subtotal float64             `json:"subtotal"`
}
