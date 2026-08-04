package mapper

import (
	"gocart/internal/dto"
	"gocart/internal/models"
)

func ToPaymentResponse(payment *models.Payment) *dto.PaymentResponse {
	return &dto.PaymentResponse{
		ID:        payment.ID,
		OrderID:   payment.OrderID,
		Reference: payment.Reference,
		Amount:    MinorUnitsToUnit(payment.Amount),
		Status:    string(payment.Status),
		Provider:  payment.Provider,
		CreatedAt: payment.CreatedAt,
		UpdatedAt: payment.UpdatedAt,
	}
}
