package dto

type CartResponse struct {
	ID        uint               `json:"id"`
	Total     float64            `json:"total"`
	ItemCount int                `json:"item_count"`
	Items     []CartItemResponse `json:"items"`
}
