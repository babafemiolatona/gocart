package middleware

import (
	"errors"
	"net/http"

	"gocart/internal/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireMerchant(
	merchantRepo repositories.MerchantRepository,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uint)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		merchant, err := merchantRepo.GetByUserID(userID)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "merchant account required",
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to verify merchant",
			})
			c.Abort()
			return
		}

		c.Set("merchantID", merchant.ID)
		c.Set("merchant", merchant)

		c.Next()
	}
}
