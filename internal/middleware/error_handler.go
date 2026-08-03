package middleware

import (
	"errors"
	"net/http"
	"os"

	apperrors "gocart/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var log = zerolog.New(os.Stderr).With().Timestamp().Logger()

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
			log.Error().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int("status", appErr.Status).
				Str("code", appErr.Code).
				Str("message", appErr.Message).
				Err(appErr.Err).
				Msg("request failed")

			writeError(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}

		log.Error().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Err(err).
			Msg("internal server error")

		writeError(
			c,
			http.StatusInternalServerError,
			"internal_server_error",
			"internal server error",
		)
	}
}
