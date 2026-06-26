package models

import "time"

type OrderItem struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	OrderID   uint `gorm:"not null" json:"order_id"`
	ProductID uint `gorm:"not null" json:"product_id"`
	// Product     Product   `gorm:"foreignKey:ProductID" json:"product"`
	ProductName string    `gorm:"not null" json:"product_name"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	Price       float64   `gorm:"not null" json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
