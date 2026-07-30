package handlers

import (
	"net/http"

	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMe godoc
//
//	@Summary		Get current user
//	@Description	Get the profile of the authenticated user
//	@Tags			Users
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.UserResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, err := getUserID(c)

	if err != nil {
		c.Error(err)
		return
	}

	user, err := h.userService.GetMe(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}
