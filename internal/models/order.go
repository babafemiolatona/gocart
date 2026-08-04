package models

import "time"

type OrderStatus string

const (
	OrderStatusPending        OrderStatus = "pending"
	OrderStatusConfirmed      OrderStatus = "confirmed"
	OrderStatusShipped        OrderStatus = "shipped"
	OrderStatusDelivered      OrderStatus = "delivered"
	OrderStatusCancelled      OrderStatus = "cancelled"
	OrderStatusPendingPayment OrderStatus = "pending_payment"
)

type Order struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	UserID          uint        `gorm:"not null" json:"user_id"`
	User            User        `gorm:"foreignKey:UserID" json:"user"`
	Status          OrderStatus `gorm:"not null" json:"status"`
	Total           int64       `gorm:"not null" json:"total"`
	ShippingAddress string      `gorm:"not null" json:"shipping_address"`
	IdempotencyKey  string      `gorm:"size:100;uniqueIndex:idx_order_idem,where:idempotency_key <> ''" json:"-"`
	Items           []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}
