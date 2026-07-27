package middleware

import (
	"errors"
	"net/http"

	apperrors "gocart/internal/errors"

	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, apperrors.ErrorResponse{
		Error: apperrors.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			writeError(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}

		writeError(
			c,
			http.StatusInternalServerError,
			"internal_server_error",
			"internal server error",
		)
	}
}
