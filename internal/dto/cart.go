package dto

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type CartResponse struct {
	ID        uint               `json:"id"`
	Total     float64            `json:"total"`
	ItemCount int                `json:"item_count"`
	Items     []CartItemResponse `json:"items"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type CartItemResponse struct {
	ID       uint                `json:"id"`
	Product  CartProductResponse `json:"product"`
	Quantity int                 `json:"quantity"`
	Price    float64             `json:"price"`
	Subtotal float64             `json:"subtotal"`
}

type CartProductResponse struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	ImageURL string  `json:"image_url,omitempty"`
}
