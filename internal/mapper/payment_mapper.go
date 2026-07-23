package mapper

import (
	"gocart/internal/dto"
	"gocart/internal/models"
)

func ToPaymentCheckoutResponse(payment *models.Payment) dto.PaymentCheckoutResponse {
	return dto.PaymentCheckoutResponse{
		Reference: payment.Reference,
		Amount:    payment.Amount,
		Status:    string(payment.Status),
	}
}
