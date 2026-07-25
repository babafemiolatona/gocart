package dto

import "time"

type MerchantRegistrationRequest struct {
	BusinessName string `json:"business_name" binding:"required"`
	Description  string `json:"description"`
	Phone        string `json:"phone"`
	LogoURL      string `json:"logo_url"`
}

type UpdateMerchantRequest struct {
	BusinessName string `json:"business_name" binding:"required,min=3,max=255"`
	Description  string `json:"description"`
	Phone        string `json:"phone"`
	LogoURL      string `json:"logo_url"`
}

type MerchantResponse struct {
	ID           uint      `json:"id"`
	BusinessName string    `json:"business_name"`
	Description  string    `json:"description"`
	Phone        string    `json:"phone"`
	LogoURL      string    `json:"logo_url"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
}
