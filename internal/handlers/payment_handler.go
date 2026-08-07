package handlers

import (
	"net/http"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService PaymentService
}

func NewPaymentHandler(paymentService PaymentService) *PaymentHandler {
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

	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	payment, err := h.paymentService.ProcessPayment(userID, reference)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, payment)
}

// PaymentWebhook godoc
//
//	@Summary		Handle payment provider webhook
//	@Description	Receives a signed payment-provider callback and finalizes the payment
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Param			X-Webhook-Signature	header		string	true	"HMAC-SHA256 signature of the raw body"
//	@Param			event				body		dto.PaymentWebhookEvent	true	"Webhook event"
//	@Success		200					{object}	dto.PaymentResponse
//	@Failure		400					{object}	errors.ErrorResponse
//	@Failure		401					{object}	errors.ErrorResponse
//	@Failure		404					{object}	errors.ErrorResponse
//	@Failure		500					{object}	errors.ErrorResponse
//	@Router			/api/v1/webhooks/payments [post]
func (h *PaymentHandler) PaymentWebhook(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeInvalidWebhookEvent,
			"failed to read request body",
			err,
		))
		return
	}

	signature := c.GetHeader("X-Webhook-Signature")
	if signature == "" {
		c.Error(apperrors.New(
			http.StatusUnauthorized,
			apperrors.CodeInvalidWebhookSignature,
			"missing webhook signature",
			nil,
		))
		return
	}

	payment, err := h.paymentService.ProcessWebhook(body, signature)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, payment)
}

// SimulatePayment godoc
//
//	@Summary		Simulate a payment webhook
//	@Description	Dev-only: builds and delivers a signed webhook event as if sent by a payment provider
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.SimulatePaymentRequest	true	"Webhook simulation request"
//	@Success		200		{object}	dto.PaymentResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		404		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/dev/simulate-payment [post]
func (h *PaymentHandler) SimulatePayment(c *gin.Context) {
	var req dto.SimulatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeValidationError,
			"invalid simulate payment request",
			err,
		))
		return
	}

	payment, err := h.paymentService.SimulateWebhook(req.Reference, req.Status)
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

	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	payment, err := h.paymentService.GetPayment(userID, reference)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, payment)
}
