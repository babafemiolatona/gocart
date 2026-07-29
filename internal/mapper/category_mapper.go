package mapper

import (
	"gocart/internal/dto"
	"gocart/internal/models"
)

func ToCategoryResponse(category *models.Category) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		Slug:        category.Slug,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}

func ToCategoryResponses(categories []models.Category) []dto.CategoryResponse {
	responses := make([]dto.CategoryResponse, len(categories))

	for i, category := range categories {
		responses[i] = *ToCategoryResponse(&category)
	}

	return responses
}
