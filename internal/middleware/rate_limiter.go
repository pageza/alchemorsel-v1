package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var loginLimiter = rate.NewLimiter(1, 3) // Adjust rate and burst accordingly

// cursor--NEW: Define a limiter for forgot/reset password endpoints.
var forgotPasswordLimiter = rate.NewLimiter(1, 3)

// LoginRateLimiter limits the rate of login attempts.
func LoginRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply the limiter on login endpoint.
		if !loginLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Please try again later."})
			return
		}
		c.Next()
	}
}

// cursor--NEW: ForgotPasswordRateLimiter limits the rate of forgot/reset password endpoints.
func ForgotPasswordRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !forgotPasswordLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again later."})
			return
		}
		c.Next()
	}
}

// cursor--ADD: Create stub for rate limiter middleware.
// RateLimiter is a stub middleware for rate limiting.
// In production, replace this with a fully featured rate limiter.
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// No rate limiting logic is applied; simply proceed.
		c.Next()
	}
}
