package middleware

import (
	"net/http"
	"strconv"
	"strings"

	apperrors "gocart/internal/errors"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(AuthService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(apperrors.New(
				http.StatusUnauthorized,
				"missing_auth_header",
				"missing authorization header",
				nil,
			))
			c.Abort()
			return
		}

		var tokenString string

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		} else {
			tokenString = strings.TrimSpace(authHeader)
		}

		if tokenString == "" {
			c.Error(apperrors.New(
				http.StatusUnauthorized,
				"invalid_auth_header",
				"invalid authorization header",
				nil,
			))
			c.Abort()
			return
		}

		claims, err := AuthService.VerifyToken(tokenString)
		if err != nil {
			c.Error(apperrors.New(
				http.StatusUnauthorized,
				"invalid_token",
				"invalid token",
				nil,
			))
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, 32)
		if err != nil {
			c.Error(apperrors.New(
				http.StatusUnauthorized,
				"invalid_user_id",
				"invalid user ID",
				nil,
			))
			c.Abort()
			return
		}

		c.Set("userID", uint(userID))
		c.Set("userRole", claims.Role)

		c.Next()
	}
}
