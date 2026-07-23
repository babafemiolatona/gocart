package mapper

import (
	"gocart/internal/dto"
	"gocart/internal/models"
)

func ToOrderItemResponse(item models.OrderItem) dto.OrderItemResponse {
	return dto.OrderItemResponse{
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		Quantity:    item.Quantity,
		Price:       item.Price,
	}
}

func ToOrderItemResponses(items []models.OrderItem) []dto.OrderItemResponse {
	responses := make([]dto.OrderItemResponse, len(items))

	for i, item := range items {
		responses[i] = ToOrderItemResponse(item)
	}

	return responses
}

func ToOrderDetailsResponse(order *models.Order) *dto.OrderDetailsResponse {
	return &dto.OrderDetailsResponse{
		ID:              order.ID,
		Status:          string(order.Status),
		Total:           order.Total,
		ShippingAddress: order.ShippingAddress,
		Items:           ToOrderItemResponses(order.Items),
		CreatedAt:       order.CreatedAt,
	}
}

func ToOrderCheckoutResponse(order *models.Order) dto.OrderCheckoutResponse {
	return dto.OrderCheckoutResponse{
		ID:              order.ID,
		Status:          string(order.Status),
		Total:           order.Total,
		ShippingAddress: order.ShippingAddress,
		Items:           ToOrderItemResponses(order.Items),
	}
}
