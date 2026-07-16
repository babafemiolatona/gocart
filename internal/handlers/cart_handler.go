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

type CartHandler struct {
	cartService *services.CartService
}

func NewCartHandler(cartService *services.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func getUserID(c *gin.Context) (uint, error) {
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

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, err := getUserID(c)

	if err != nil {
		c.Error(apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized access",
			err,
		))
		return
	}

	cart, err := h.cartService.GetCart(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, err := getUserID(c)

	if err != nil {
		c.Error(apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized access",
			err,
		))
		return
	}

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
		return
	}

	cart, err := h.cartService.AddToCart(userID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	userID, err := getUserID(c)

	if err != nil {
		c.Error(apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized access",
			err,
		))
		return
	}

	id, err := strconv.ParseUint(c.Param("itemID"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_cart_item_id",
			"invalid cart item id",
			err,
		))
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
		return
	}

	cart, err := h.cartService.UpdateCartItem(userID, uint(id), req.Quantity)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized access",
			err,
		))
		return
	}

	id, err := strconv.ParseUint(c.Param("itemID"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_cart_item_id",
			"invalid cart item id",
			err,
		))
		return
	}

	cart, err := h.cartService.RemoveFromCart(userID, uint(id))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized access",
			err,
		))
		return
	}

	if err := h.cartService.ClearCart(userID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
