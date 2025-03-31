package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
)

// RespondError sends a standardized error response.
func RespondError(c *gin.Context, code int, message string) {
	c.JSON(code, dtos.ErrorResponse{
		Code:    "ERROR",
		Message: message,
	})
}

// RespondSuccess sends a standardized success response.
func RespondSuccess(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}
