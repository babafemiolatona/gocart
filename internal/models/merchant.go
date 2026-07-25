package models

import "time"

type Merchant struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"-"`
	BusinessName string    `gorm:"size:255;not null" json:"business_name"`
	Description  string    `gorm:"type:text" json:"description"`
	Phone        string    `gorm:"size:20" json:"phone"`
	LogoURL      string    `json:"logo_url"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
