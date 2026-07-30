package handlers

import (
	"net/http"

	apperrors "gocart/internal/errors"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// ProcessPayment godoc
//
//	@Summary		Process payment
//	@Description	Process a payment using its reference
//	@Tags			Payments
//	@Security		BearerAuth
//	@Produce		json
//	@Param			reference	path		string	true	"Payment Reference"
//	@Success		200			{object}	dto.PaymentResponse
//	@Failure		400			{object}	errors.ErrorResponse
//	@Failure		401			{object}	errors.ErrorResponse
//	@Failure		404			{object}	errors.ErrorResponse
//	@Failure		409			{object}	errors.ErrorResponse
//	@Failure		500			{object}	errors.ErrorResponse
//	@Router			/api/v1/payments/{reference}/process [post]
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_payment_reference",
			"payment reference is required",
			nil,
		))
		return
	}

	payment, err := h.paymentService.ProcessPayment(reference)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetPayment godoc
//
//	@Summary		Get payment
//	@Description	Get payment details by reference
//	@Tags			Payments
//	@Security		BearerAuth
//	@Produce		json
//	@Param			reference	path		string	true	"Payment Reference"
//	@Success		200			{object}	dto.PaymentResponse
//	@Failure		400			{object}	errors.ErrorResponse
//	@Failure		401			{object}	errors.ErrorResponse
//	@Failure		404			{object}	errors.ErrorResponse
//	@Failure		500			{object}	errors.ErrorResponse
//	@Router			/api/v1/payments/{reference} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_payment_reference",
			"payment reference is required",
			nil,
		))
		return
	}

	payment, err := h.paymentService.GetPayment(reference)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, payment)
}
