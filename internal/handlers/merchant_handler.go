package handlers

import (
	"net/http"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
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

func (h *MerchantHandler) RegisterMerchant(c *gin.Context) {

	var req dto.MerchantRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
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

func (h *MerchantHandler) GetMe(c *gin.Context) {

	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	merchant, err := h.merchantService.GetProfile(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, merchant)
}

func (h *MerchantHandler) UpdateMe(c *gin.Context) {

	var req dto.UpdateMerchantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	merchant, err := h.merchantService.UpdateProfile(
		userID,
		&req,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, merchant)
}
