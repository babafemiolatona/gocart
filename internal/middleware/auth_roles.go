package middleware

import (
	"gocart/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole is a middleware that checks if the user has the required role
func RequireRole(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleValue, exists := c.Get("userRole")
		if !exists {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userRole, ok := userRoleValue.(models.Role)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden - insufficient permissions"})
		c.Abort()
	}
}
