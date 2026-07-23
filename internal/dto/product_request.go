package dto

type CreateProductRequest struct {
	Name        string  `json:"name" form:"name" binding:"required"`
	Description string  `json:"description" form:"description"`
	Price       float64 `json:"price" form:"price" binding:"required"`
	Stock       int     `json:"stock" form:"stock" binding:"required"`
	CategoryID  uint    `json:"category_id" form:"category_id" binding:"required"`
	Sku         string  `json:"sku" form:"sku" binding:"required"`
	Slug        string  `json:"slug" form:"slug" binding:"required"`
}
