package mapper

import (
	"gocart/internal/dto"
	"gocart/internal/models"
)

func ToProductResponse(product *models.Product) *dto.ProductResponse {
	if product == nil {
		return nil
	}

	return &dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CategoryID:  product.CategoryID,
		MerchantID:  product.MerchantID,
		Slug:        product.Slug,
		Sku:         product.Sku,
		Images:      ToProductImageResponses(product.Images),
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func ToProductResponses(products []models.Product) []dto.ProductResponse {
	responses := make([]dto.ProductResponse, len(products))

	for i, product := range products {
		responses[i] = *ToProductResponse(&product)
	}

	return responses
}

func ToProductImageResponse(image *models.ProductImage) dto.ProductImageResponse {
	return dto.ProductImageResponse{
		ID:        image.ID,
		ImageURL:  image.ImageURL,
		IsPrimary: image.IsPrimary,
		CreatedAt: image.CreatedAt,
	}
}

func ToProductImageResponses(images []models.ProductImage) []dto.ProductImageResponse {
	responses := make([]dto.ProductImageResponse, len(images))

	for i, image := range images {
		responses[i] = ToProductImageResponse(&image)
	}

	return responses
}
