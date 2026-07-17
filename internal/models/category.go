package models

import "time"

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

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
