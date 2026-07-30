package handlers

import (
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
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

// GetCart godoc
//
//	@Summary		Get current user's cart
//	@Description	Retrieve the authenticated user's shopping cart
//	@Tags			Cart
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.CartResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, err := getUserID(c)

	if err != nil {
		c.Error(err)
		return
	}

	cart, err := h.cartService.GetCartResponse(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

// AddToCart godoc
//
//	@Summary		Add item to cart
//	@Description	Add a product to the authenticated user's cart
//	@Tags			Cart
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.AddToCartRequest	true	"Add to cart request"
//	@Success		200		{object}	dto.CartResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		404		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/cart/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, err := getUserID(c)

	if err != nil {
		c.Error(err)
		return
	}

	var req dto.AddToCartRequest
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

// UpdateCartItem godoc
//
//	@Summary		Update cart item
//	@Description	Update the quantity of an item in the authenticated user's cart
//	@Tags			Cart
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			itemID	path		int								true	"Cart Item ID"
//	@Param			request	body		dto.UpdateCartItemRequest		true	"Update cart item request"
//	@Success		200		{object}	dto.CartResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		404		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/cart/items/{itemID} [put]
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	userID, err := getUserID(c)

	if err != nil {
		c.Error(err)
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

	var req dto.UpdateCartItemRequest
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

// RemoveFromCart godoc
//
//	@Summary		Remove item from cart
//	@Description	Remove an item from the authenticated user's cart
//	@Tags			Cart
//	@Security		BearerAuth
//	@Produce		json
//	@Param			itemID	path		int	true	"Cart Item ID"
//	@Success		200		{object}	dto.CartResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		404		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/cart/items/{itemID} [delete]
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
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

// ClearCart godoc
//
//	@Summary		Clear cart
//	@Description	Remove all items from the authenticated user's cart
//	@Tags			Cart
//	@Security		BearerAuth
//	@Success		204
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.cartService.ClearCart(userID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
