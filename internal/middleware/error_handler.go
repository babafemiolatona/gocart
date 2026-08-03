package middleware

import (
	"errors"
	"net/http"

	apperrors "gocart/internal/errors"
	"gocart/internal/logger"

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

func logError(c *gin.Context, err error) {
	var appErr *apperrors.AppError

	if errors.As(err, &appErr) {
		logger.Log.Error().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", appErr.Status).
			Str("code", appErr.Code).
			Str("message", appErr.Message).
			Err(appErr.Err).
			Msg("request failed")
		return
	}

	logger.Log.Error().
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Err(err).
		Msg("request failed")
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		for _, entry := range c.Errors {
			logError(c, entry.Err)
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
