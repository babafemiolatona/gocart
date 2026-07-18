package handlers

import (
	"errors"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func getUserId(c *gin.Context) (uint, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, errors.New("missing user id")
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, errors.New("invalid user id")
	}

	return id, nil
}

func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized",
			err,
		))
	}

	var req models.CheckoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
		return
	}

	order, err := h.orderService.ProcessCheckout(userID, req.ShippingAddress)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized",
			err,
		))
		return
	}

	orders, err := h.orderService.GetUserOrders(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, orders)
}

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

	order, err := h.orderService.GetOrder(uint(orderID))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_order_id",
			"invalid order id",
			err,
		))
	}

	err = h.orderService.CancelOrder(uint(orderID))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}
