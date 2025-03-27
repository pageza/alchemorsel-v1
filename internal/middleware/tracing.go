package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/monitoring"
	"go.uber.org/zap"
)

const (
	RequestIDKey = "request_id"
	StartTimeKey = "start_time"
)

// Tracing middleware adds request tracing and logging
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)

		// Set start time
		startTime := time.Now()
		c.Set(StartTimeKey, startTime)

		// Create logger with request context
		logger := zap.L().With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		// Log request
		logger.Info("Incoming request",
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()),
		)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log response
		logger.Info("Request completed",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)

		// Record metrics
		monitoring.ObserveHTTPRequest(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)

		// Add response headers
		c.Header("X-Request-ID", requestID)
		c.Header("X-Response-Time", duration.String())
	}
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}

// GetStartTime retrieves the request start time from the context
func GetStartTime(c *gin.Context) time.Time {
	if startTime, exists := c.Get(StartTimeKey); exists {
		return startTime.(time.Time)
	}
	return time.Now()
}
