package models

import "time"

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;index" json:"name"`
	Description string    `json:"description"`
	Price       float64   `gorm:"not null" json:"price"`
	Stock       int       `gorm:"not null;default:0" json:"stock"`
	CategoryID  uint      `gorm:"not null" json:"category_id"`
	Category    Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	ImageURL    string    `json:"image_url"`
	Slug        string    `gorm:"uniqueIndex" json:"slug"`
	Sku         string    `gorm:"uniqueIndex" json:"sku"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock" binding:"required"`
	CategoryID  uint    `json:"category_id" binding:"required"`
	Sku         string  `json:"sku" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	CategoryID  *uint    `json:"category_id"`
	Sku         *string  `json:"sku"`
	Slug        *string  `json:"slug"`
}

type ProductFilters struct {
	CategoryID  uint
	MinPrice    float64
	MaxPrice    float64
	InStock     *bool
	SearchQuery string
}
