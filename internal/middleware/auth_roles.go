package middleware

import (
	"net/http"

	apperrors "gocart/internal/errors"
	"gocart/internal/models"

	"github.com/gin-gonic/gin"
)

// RequireRole is a middleware that checks if the user has the required role
func RequireRole(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleValue, exists := c.Get("userRole")
		if !exists {
			c.Error(apperrors.New(
				http.StatusUnauthorized,
				"unauthorized",
				"unauthorized",
				nil,
			))
			c.Abort()
			return
		}

		userRole, ok := userRoleValue.(models.Role)

		if !ok {
			c.Error(apperrors.New(
				http.StatusUnauthorized,
				"unauthorized",
				"unauthorized",
				nil,
			))
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.Error(apperrors.New(
			http.StatusForbidden,
			"forbidden",
			"forbidden - insufficient permissions",
			nil,
		))
		c.Abort()
	}
}
