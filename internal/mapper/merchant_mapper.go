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

func ToMerchantOrderItemResponse(item models.OrderItem) dto.MerchantOrderItemResponse {
	return dto.MerchantOrderItemResponse{
		ID:          item.ID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		Quantity:    item.Quantity,
		Price:       item.Price,
	}
}

func ToMerchantOrderResponse(order *models.Order) *dto.MerchantOrderResponse {
	items := make([]dto.MerchantOrderItemResponse, 0, len(order.Items))

	for _, item := range order.Items {
		items = append(items, ToMerchantOrderItemResponse(item))
	}

	return &dto.MerchantOrderResponse{
		ID:              order.ID,
		Status:          string(order.Status),
		Total:           order.Total,
		ShippingAddress: order.ShippingAddress,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		Items:           items,
	}
}

func ToMerchantOrderResponses(orders []models.Order) []dto.MerchantOrderResponse {
	responses := make([]dto.MerchantOrderResponse, 0, len(orders))

	for _, order := range orders {
		responses = append(responses, *ToMerchantOrderResponse(&order))
	}

	return responses
}

func ToMerchantRecentOrderResponse(order models.Order) dto.MerchantRecentOrderResponse {

	return dto.MerchantRecentOrderResponse{
		ID:        order.ID,
		Customer:  order.User.Email,
		Status:    order.Status,
		Total:     order.Total,
		ItemCount: len(order.Items),
		CreatedAt: order.CreatedAt,
	}
}

func ToMerchantRecentOrderResponses(orders []models.Order) []dto.MerchantRecentOrderResponse {

	responses := make(
		[]dto.MerchantRecentOrderResponse,
		0,
		len(orders),
	)

	for _, order := range orders {
		responses = append(
			responses,
			ToMerchantRecentOrderResponse(order),
		)
	}

	return responses
}
