package middleware

import (
	"net/http"
	"os"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Tollbooth-based rate limiter for general endpoints.
func RateLimiter() gin.HandlerFunc {
	// Set the allowed request count.
	allowedRequests := 5.0
	if gin.Mode() == gin.TestMode || os.Getenv("INTEGRATION_TEST") == "true" {
		allowedRequests = 1.0
	}
	lmt := tollbooth.NewLimiter(allowedRequests, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	return func(c *gin.Context) {
		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

var loginLimiter = rate.NewLimiter(1, 3)          // Adjust rate and burst as needed.
var forgotPasswordLimiter = rate.NewLimiter(1, 3) // Adjust rate and burst as needed.

// LoginRateLimiter limits the rate of login attempts.
func LoginRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !loginLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Please try again later."})
			return
		}
		c.Next()
	}
}

// ForgotPasswordRateLimiter limits the rate of forgot/reset password endpoints.
func ForgotPasswordRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !forgotPasswordLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again later."})
			return
		}
		c.Next()
	}
}
