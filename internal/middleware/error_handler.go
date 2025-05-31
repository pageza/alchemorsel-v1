package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/errors"
	"go.uber.org/zap"
)



// ErrorHandler middleware handles errors consistently across the application
func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Log the error
			logger.Error("HTTP error",
				zap.Error(err),
				zap.Int("status_code", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("remote_addr", c.ClientIP()),
			)

		// Create error response
		response := dtos.ErrorResponse{
			Code:    getErrorCode(c.Writer.Status()),
			Message: err.Error(),
		}

		// Send response
		c.JSON(c.Writer.Status(), response)
		}
	}
}

// getErrorCode returns the appropriate error code for the given HTTP status
func getErrorCode(status int) string {
	switch status {
	case 400:
		return errors.ErrValidation
	case 401:
		return errors.ErrUnauthorized
	case 403:
		return errors.ErrForbidden
	case 404:
		return errors.ErrNotFound
	case 409:
		return errors.ErrConflict
	case 500:
		return errors.ErrInternal
	case 504:
		return errors.ErrTimeout
	default:
		return errors.ErrInternal
	}
}
