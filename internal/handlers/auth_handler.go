package handlers

import (
	"net/http"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Create a new customer account
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest			true	"Register request"
//	@Success		201		{object}	dto.UserResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		409		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError(err))
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login godoc
//
//	@Summary		Authenticate user
//	@Description	Login with email or username and password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest			true	"Login request"
//	@Success		200		{object}	dto.AuthResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError(err))
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChangePassword godoc
//
//	@Summary		Change password
//	@Description	Change the authenticated user's password after verifying the current one
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ChangePasswordRequest	true	"Change password request"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/users/me/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError(err))
		return
	}

	if err := h.authService.ChangePassword(userID, &req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
