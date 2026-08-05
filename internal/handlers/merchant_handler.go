package handlers

import (
	"net/http"
	"strconv"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/query"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

type MerchantHandler struct {
	merchantService *services.MerchantService
}

func NewMerchantHandler(
	merchantService *services.MerchantService,
) *MerchantHandler {
	return &MerchantHandler{
		merchantService: merchantService,
	}
}

// RegisterMerchant godoc
//
//	@Summary		Register as a merchant
//	@Description	Register the authenticated user as a merchant
//	@Tags			Merchants
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.MerchantRegisterRequest	true	"Merchant register request"
//	@Success		201		{object}	dto.MerchantResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		409		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/register [post]
func (h *MerchantHandler) RegisterMerchant(c *gin.Context) {

	var req dto.MerchantRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError(err))
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	merchant, err := h.merchantService.RegisterMerchant(
		userID,
		&req,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, merchant)
}

// GetMe godoc
//
//	@Summary		Get current merchant
//	@Description	Get the authenticated merchant profile
//	@Tags			Merchants
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.MerchantResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		403	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/me [get]
func (h *MerchantHandler) GetMe(c *gin.Context) {

	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	merchant, err := h.merchantService.GetProfile(merchantID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, merchant)
}

// UpdateMe godoc
//
//	@Summary		Update current merchant
//	@Description	Update the authenticated merchant profile
//	@Tags			Merchants
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UpdateMerchantRequest	true	"Merchant update request"
//	@Success		200	{object}	dto.MerchantResponse
//	@Failure		400	{object}	errors.ErrorResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		403	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/me [put]
func (h *MerchantHandler) UpdateMe(c *gin.Context) {

	var req dto.UpdateMerchantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError(err))
		return
	}

	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	merchant, err := h.merchantService.UpdateProfile(
		merchantID,
		&req,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, merchant)
}

// GetOrders godoc
//
//	@Summary		Get merchant orders
//	@Description	Get all orders for the authenticated merchant
//	@Tags			Merchant Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}		dto.MerchantOrderResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		403	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/orders [get]
func (h *MerchantHandler) GetOrders(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	orders, err := h.merchantService.GetOrders(merchantID, query.ParsePagination(c))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrder godoc
//
//	@Summary		Get merchant order
//	@Description	Get a single order belonging to the authenticated merchant
//	@Tags			Merchant Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Order ID"
//	@Success		200	{object}	dto.MerchantOrderResponse
//	@Failure		400	{object}	errors.ErrorResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		403	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/orders/{id} [get]
func (h *MerchantHandler) GetOrder(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_order_id",
			"invalid order id",
			err,
		))
		return
	}

	order, err := h.merchantService.GetOrder(
		merchantID,
		uint(orderID),
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, order)
}

// UpdateOrderStatus godoc
//
//	@Summary		Update merchant order status
//	@Description	Update the status of an order belonging to the authenticated merchant
//	@Tags			Merchant Orders
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int								true	"Order ID"
//	@Param			request	body		dto.UpdateOrderStatusRequest		true	"Order status"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		403		{object}	errors.ErrorResponse
//	@Failure		404		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/orders/{id}/status [patch]
func (h *MerchantHandler) UpdateOrderStatus(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_order_id",
			"invalid order id",
			err,
		))
		return
	}

	var req dto.UpdateOrderStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError(err))
		return
	}

	err = h.merchantService.UpdateOrderStatus(
		merchantID,
		uint(orderID),
		&req,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "order status updated successfully",
	})
}

// GetDashboard godoc
//
//	@Summary		Get merchant dashboard
//	@Description	Get dashboard metrics for the authenticated merchant
//	@Tags			Merchants
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.MerchantDashboardResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		403	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/dashboard [get]
func (h *MerchantHandler) GetDashboard(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	dashboard, err := h.merchantService.GetDashboard(merchantID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
