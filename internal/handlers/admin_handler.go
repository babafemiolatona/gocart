package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService AdminService
}

func NewAdminHandler(
	adminService AdminService,
) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// GetDashboard godoc
//
//	@Summary		Get admin dashboard
//	@Description	Get platform-wide metrics for admins
//	@Tags			Admin
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.AdminDashboardResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		403	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/admin/dashboard [get]
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	dashboard, err := h.adminService.GetDashboard()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
