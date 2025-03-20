package response

import (
	"github.com/gin-gonic/gin"
)

// RespondError sends a standardized error response.
func RespondError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

// RespondSuccess sends a standardized success response.
func RespondSuccess(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}
