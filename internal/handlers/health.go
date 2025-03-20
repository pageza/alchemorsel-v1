package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck provides a simple endpoint that returns 200.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
