package dto

import "time"

type CategoryRequest struct {
	Name        string `json:"name" binding:"required,min=3"`
	Description string `json:"description"`
	Slug        string `json:"slug" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=3"`
	Description *string `json:"description"`
	Slug        *string `json:"slug"`
}

type CategoryResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
