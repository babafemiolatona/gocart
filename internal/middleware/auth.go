package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(AuthService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		claims, err := AuthService.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			c.Abort()
			return
		}

		c.Set("userID", uint(userID))
		c.Set("userRole", claims.Role)

		c.Next()
	}
}
