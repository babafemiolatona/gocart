package models

import "time"

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type Payment struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	OrderID   uint          `gorm:"not null;index" json:"order_id"`
	Order     Order         `gorm:"foreignKey:OrderID" json:"-"`
	Reference string        `gorm:"size:100;uniqueIndex;not null" json:"reference"`
	Amount    float64       `gorm:"not null" json:"amount"`
	Status    PaymentStatus `gorm:"size:20;default:'pending'" json:"status"`
	Provider  string        `gorm:"size:50;default:'mock'" json:"provider"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
