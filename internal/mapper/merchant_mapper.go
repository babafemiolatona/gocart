package mapper

import (
	"gocart/internal/dto"
	"gocart/internal/models"
)

func ToMerchantResponse(merchant *models.Merchant) *dto.MerchantResponse {
	return &dto.MerchantResponse{
		ID:           merchant.ID,
		BusinessName: merchant.BusinessName,
		Description:  merchant.Description,
		Phone:        merchant.Phone,
		LogoURL:      merchant.LogoURL,
		IsVerified:   merchant.IsVerified,
		CreatedAt:    merchant.CreatedAt,
		UpdatedAt:    merchant.UpdatedAt,
	}
}
