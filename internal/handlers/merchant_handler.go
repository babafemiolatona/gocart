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
//	@Router			/api/v1/merchant/register [post]
func (h *MerchantHandler) RegisterMerchant(c *gin.Context) {

	var req dto.MerchantRegisterRequest

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
