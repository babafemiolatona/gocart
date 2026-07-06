package models

import "time"

type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;index" json:"name"`
	Description string         `json:"description"`
	Price       float64        `gorm:"not null" json:"price"`
	Stock       int            `gorm:"not null;default:0" json:"stock"`
	CategoryID  uint           `gorm:"not null" json:"category_id"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images      []ProductImage `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;" json:"images,omitempty"`
	Slug        string         `gorm:"uniqueIndex" json:"slug"`
	Sku         string         `gorm:"uniqueIndex" json:"sku"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ProductImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"-"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	ImageURL  string    `gorm:"not null" json:"image_url"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" form:"name" binding:"required"`
	Description string  `json:"description" form:"description"`
	Price       float64 `json:"price" form:"price" binding:"required"`
	Stock       int     `json:"stock" form:"stock" binding:"required"`
	CategoryID  uint    `json:"category_id" form:"category_id" binding:"required"`
	Sku         string  `json:"sku" form:"sku" binding:"required"`
	Slug        string  `json:"slug" form:"slug" binding:"required"`
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
