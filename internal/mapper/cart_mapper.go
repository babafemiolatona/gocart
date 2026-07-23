package mapper

import (
	"gocart/internal/dto"
	"gocart/internal/models"
)

func ToCartProductResponse(product models.Product) dto.CartProductResponse {
	imageURL := ""

	for _, image := range product.Images {
		if image.IsPrimary {
			imageURL = image.ImageURL
			break
		}
	}

	if imageURL == "" && len(product.Images) > 0 {
		imageURL = product.Images[0].ImageURL
	}

	return dto.CartProductResponse{
		ID:       product.ID,
		Name:     product.Name,
		Price:    product.Price,
		ImageURL: imageURL,
	}
}

func ToCartItemResponse(item models.CartItem) dto.CartItemResponse {
	return dto.CartItemResponse{
		ID:       item.ID,
		Product:  ToCartProductResponse(item.Product),
		Quantity: item.Quantity,
		Price:    item.Price,
		Subtotal: item.Price * float64(item.Quantity),
	}
}

func ToCartItemResponses(items []models.CartItem) []dto.CartItemResponse {
	responses := make([]dto.CartItemResponse, len(items))

	for i, item := range items {
		responses[i] = ToCartItemResponse(item)
	}

	return responses
}

func ToCartResponse(cart *models.Cart) *dto.CartResponse {
	return &dto.CartResponse{
		ID:        cart.ID,
		Total:     cart.Total,
		ItemCount: cart.ItemCount,
		Items:     ToCartItemResponses(cart.Items),
	}
}
