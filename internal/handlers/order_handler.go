package handlers

import (
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

func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
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
	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
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
