package middleware

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
	ExpirationTTL     time.Duration
}

// DefaultConfig returns default rate limit configuration
func DefaultConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 5.0,
		Burst:             10,
		ExpirationTTL:     time.Hour,
	}
}

// TestConfig returns configuration suitable for testing
func TestConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 1.0,
		Burst:             1,
		ExpirationTTL:     time.Minute,
	}
}

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.RWMutex
)

// getLimiter returns or creates a rate limiter for the given key
func getLimiter(key string, config RateLimitConfig) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if limiter, exists := limiters[key]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst)
	limiters[key] = limiter
	return limiter
}

// RateLimiter limits the rate of requests per IP and path
func RateLimiter() gin.HandlerFunc {
	config := DefaultConfig()
	if gin.Mode() == gin.TestMode || os.Getenv("INTEGRATION_TEST") == "true" {
		config = TestConfig()
	}

	lmt := tollbooth.NewLimiter(config.RequestsPerSecond, &limiter.ExpirableOptions{
		DefaultExpirationTTL: config.ExpirationTTL,
	})

	return func(c *gin.Context) {
		// Get client IP for rate limiting
		clientIP := c.ClientIP()
		path := c.Request.URL.Path

		// Create a unique key for this client and path
		key := clientIP + ":" + path
		limiter := getLimiter(key, config)

		if !limiter.Allow() {
			retryAfter := time.Second
			if config.RequestsPerSecond > 0 {
				retryAfter = time.Second / time.Duration(config.RequestsPerSecond)
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": time.Until(time.Now().Add(retryAfter)).String(),
			})
			return
		}

		// Check tollbooth limiter as well
		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if err != nil {
			retryAfter := time.Second
			if config.RequestsPerSecond > 0 {
				retryAfter = time.Second / time.Duration(config.RequestsPerSecond)
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": time.Until(time.Now().Add(retryAfter)).String(),
			})
			return
		}

		c.Next()
	}
}

// LoginRateLimiter limits the rate of login attempts per IP
func LoginRateLimiter() gin.HandlerFunc {
	config := RateLimitConfig{
		RequestsPerSecond: 0.1, // 1 request per 10 seconds
		Burst:             1,
		ExpirationTTL:     time.Hour,
	}

	if gin.Mode() == gin.TestMode || os.Getenv("INTEGRATION_TEST") == "true" {
		config = TestConfig()
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiter := getLimiter("login:"+clientIP, config)

		if !limiter.Allow() {
			retryAfter := time.Second
			if config.RequestsPerSecond > 0 {
				retryAfter = time.Second / time.Duration(config.RequestsPerSecond)
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many login attempts. Please try again later.",
				"retry_after": time.Until(time.Now().Add(retryAfter)).String(),
			})
			return
		}

		c.Next()
	}
}

// ForgotPasswordRateLimiter limits the rate of forgot/reset password endpoints per IP
func ForgotPasswordRateLimiter() gin.HandlerFunc {
	config := RateLimitConfig{
		RequestsPerSecond: 0.05, // 1 request per 20 seconds
		Burst:             1,
		ExpirationTTL:     time.Hour,
	}

	if gin.Mode() == gin.TestMode || os.Getenv("INTEGRATION_TEST") == "true" {
		config = TestConfig()
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiter := getLimiter("forgot_password:"+clientIP, config)

		if !limiter.Allow() {
			retryAfter := time.Second
			if config.RequestsPerSecond > 0 {
				retryAfter = time.Second / time.Duration(config.RequestsPerSecond)
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many password reset attempts. Please try again later.",
				"retry_after": time.Until(time.Now().Add(retryAfter)).String(),
			})
			return
		}

		c.Next()
	}
}

// ResetLimiters resets all rate limiters (useful in tests)
func ResetLimiters() {
	mu.Lock()
	defer mu.Unlock()

	limiters = make(map[string]*rate.Limiter)

	// If TEST_RATE_LIMIT_STRICT or INTEGRATION_TEST is set, use strict limits
	if os.Getenv("TEST_RATE_LIMIT_STRICT") == "true" || os.Getenv("INTEGRATION_TEST") == "true" {
		// Create strict limiters for test mode
		limiters["login:test"] = rate.NewLimiter(0, 1)
		limiters["forgot_password:test"] = rate.NewLimiter(0, 1)
	} else if gin.Mode() == gin.TestMode {
		// Create test mode limiters
		limiters["login:test"] = rate.NewLimiter(0.1, 1)
		limiters["forgot_password:test"] = rate.NewLimiter(0.1, 1)
	}
}

func init() {
	ResetLimiters()
}
