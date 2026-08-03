package middleware

import (
	"errors"
	"net/http"

	apperrors "gocart/internal/errors"
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
			c.Error(apperrors.New(
				http.StatusUnauthorized,
				"unauthorized",
				"unauthorized",
				nil,
			))
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uint)
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

		merchant, err := merchantRepo.GetByUserID(userID)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Error(apperrors.New(
					http.StatusForbidden,
					"merchant_required",
					"merchant account required",
					nil,
				))
				c.Abort()
				return
			}

			c.Error(apperrors.New(
				http.StatusInternalServerError,
				"verify_merchant_failed",
				"failed to verify merchant",
				err,
			))
			c.Abort()
			return
		}

		c.Set("merchantID", merchant.ID)
		c.Set("merchant", merchant)

		c.Next()
	}
}
