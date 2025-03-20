package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var loginLimiter = rate.NewLimiter(1, 3) // Adjust rate and burst accordingly

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
