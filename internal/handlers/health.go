package handlers

import (
	"gocart/internal/models"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(200, models.HealthResponse{
		Status:  "OK",
		Message: "Server is running",
		Version: "1.0.0",
	})
}
