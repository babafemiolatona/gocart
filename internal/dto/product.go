package dto

import "time"

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

type ProductResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Price       float64                `json:"price"`
	Stock       int                    `json:"stock"`
	CategoryID  uint                   `json:"category_id"`
	MerchantID  uint                   `json:"merchant_id"`
	Slug        string                 `json:"slug"`
	Sku         string                 `json:"sku"`
	Images      []ProductImageResponse `json:"images"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ProductImageResponse struct {
	ID        uint      `json:"id"`
	ImageURL  string    `json:"image_url"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}
