package models

import "time"

type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;index" json:"name"`
	Description string         `json:"description"`
	Price       int64          `gorm:"not null" json:"price"`
	Stock       int            `gorm:"not null;default:0" json:"stock"`
	CategoryID  uint           `gorm:"not null" json:"category_id"`
	MerchantID  uint           `gorm:"not null" json:"merchant_id"`
	Merchant    Merchant       `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
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
