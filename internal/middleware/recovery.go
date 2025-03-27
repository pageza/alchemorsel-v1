package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/errors"
	"go.uber.org/zap"
)

// Recovery middleware recovers from panics and logs the error
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("remote_addr", c.ClientIP()),
				)

				// Create error response
				response := struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					Code:    errors.ErrInternal,
					Message: "An unexpected error occurred",
				}

				// Send response
				c.JSON(500, response)
			}
		}()

		c.Next()
	}
}
