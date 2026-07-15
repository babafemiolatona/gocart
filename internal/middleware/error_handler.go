package middleware

import (
	"errors"
	"net/http"

	apperrors "gocart/internal/errors"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *apperrors.AppError

		switch {

		case errors.As(err, &appErr):
			c.JSON(appErr.Status, gin.H{
				"error": gin.H{
					"code":    appErr.Code,
					"message": appErr.Message,
				},
			})

		case errors.Is(err, services.ErrCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "category_not_found",
					"message": "category not found",
				},
			})

		case errors.Is(err, gorm.ErrDuplicatedKey):
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "duplicate_resource",
					"message": "resource already exists",
				},
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal_server_error",
					"message": "internal server error",
				},
			})
		}
	}
}
