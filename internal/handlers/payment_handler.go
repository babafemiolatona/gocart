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
