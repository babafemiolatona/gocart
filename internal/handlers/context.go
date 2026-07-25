package handlers

import (
	apperrors "gocart/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getUserID(c *gin.Context) (uint, error) {
	value, ok := c.Get("userID")
	if !ok {
		return 0, apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized access",
			nil,
		)
	}

	userID, ok := value.(uint)
	if !ok {
		return 0, apperrors.New(
			http.StatusUnauthorized,
			"unauthorized",
			"unauthorized access",
			nil,
		)
	}

	return userID, nil
}

func getMerchantID(c *gin.Context) (uint, error) {
	id, ok := c.Get("merchantID")
	if !ok {
		return 0, apperrors.New(
			http.StatusUnauthorized,
			"merchant_required",
			"merchant authentication required",
			nil,
		)
	}

	merchantID, ok := id.(uint)
	if !ok {
		return 0, apperrors.New(
			http.StatusUnauthorized,
			"merchant_required",
			"merchant authentication required",
			nil,
		)
	}

	return merchantID, nil
}
