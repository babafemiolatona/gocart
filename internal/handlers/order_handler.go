package handlers

import (
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/query"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService OrderService
}

func NewOrderHandler(orderService OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// Checkout godoc
//
//	@Summary		Checkout cart
//	@Description	Create an order from the authenticated user's cart
//	@Tags			Orders
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CheckoutRequest	true	"Checkout request"
//	@Success		201		{object}	dto.CheckoutResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		404		{object}	errors.ErrorResponse
//	@Failure		409		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/orders/checkout [post]
func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.CheckoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError(err))
		return
	}

	order, err := h.orderService.ProcessCheckout(userID, req.ShippingAddress, req.IdempotencyKey)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetMyOrders godoc
//
//	@Summary		List my orders
//	@Description	Get all orders belonging to the authenticated user
//	@Tags			Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}		dto.OrderResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/orders [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	orders, err := h.orderService.GetUserOrders(userID, query.ParsePagination(c))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrder godoc
//
//	@Summary		Get order
//	@Description	Get an order by ID
//	@Tags			Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Order ID"
//	@Success		200	{object}	dto.OrderResponse
//	@Failure		400	{object}	errors.ErrorResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
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

	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	order, err := h.orderService.GetOrder(userID, uint(orderID))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, order)
}

// CancelOrder godoc
//
//	@Summary		Cancel order
//	@Description	Cancel an existing order
//	@Tags			Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Order ID"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	errors.ErrorResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		409	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/orders/{id}/cancel [put]
func (h *OrderHandler) CancelOrder(c *gin.Context) {

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

	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = h.orderService.CancelOrder(userID, uint(orderID))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}
