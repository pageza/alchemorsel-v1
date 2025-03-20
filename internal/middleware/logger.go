package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// Generate a unique correlation ID for the request.
		correlationID := uuid.NewString()
		c.Set("correlationID", correlationID)
		c.Next()
		duration := time.Since(start)
		zap.L().Info("HTTP request",
			zap.String("correlation_id", correlationID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	}
}
