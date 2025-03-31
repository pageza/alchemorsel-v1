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
		RequestsPerSecond: 100.0, // Allow more requests in test mode
		Burst:             100,
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

// isTestMode checks if we're in test mode
func isTestMode() bool {
	return gin.Mode() == gin.TestMode ||
		os.Getenv("INTEGRATION_TEST") == "true" ||
		os.Getenv("TEST_RATE_LIMIT_STRICT") == "true"
}

// calculateRetryAfter calculates the retry after duration
func calculateRetryAfter(config RateLimitConfig) time.Duration {
	if config.RequestsPerSecond <= 0 {
		return time.Minute // More reasonable default for testing
	}
	// Use float division to handle rates less than 1
	return time.Duration(float64(time.Second) / config.RequestsPerSecond)
}

// RateLimiter limits the rate of requests per IP and path
func RateLimiter() gin.HandlerFunc {
	config := DefaultConfig()
	if isTestMode() {
		config = TestConfig()
	}

	lmt := tollbooth.NewLimiter(config.RequestsPerSecond, &limiter.ExpirableOptions{
		DefaultExpirationTTL: config.ExpirationTTL,
	})
	lmt.SetBurst(config.Burst)

	return func(c *gin.Context) {
		// Get client IP for rate limiting
		clientIP := c.ClientIP()
		path := c.Request.URL.Path

		// Create a unique key for this client and path
		key := clientIP + ":" + path
		limiter := getLimiter(key, config)

		if !limiter.Allow() {
			retryAfter := calculateRetryAfter(config)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter.String(),
			})
			return
		}

		// Check tollbooth limiter as well
		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if err != nil {
			retryAfter := calculateRetryAfter(config)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter.String(),
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

	if isTestMode() {
		config = TestConfig()
	}

	return func(c *gin.Context) {
		if os.Getenv("DISABLE_RATE_LIMITER") == "true" {
			c.Next()
			return
		}
		clientIP := c.ClientIP()
		limiter := getLimiter("login:"+clientIP, config)

		if !limiter.Allow() {
			retryAfter := calculateRetryAfter(config)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many login attempts. Please try again later.",
				"retry_after": retryAfter.String(),
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

	if isTestMode() {
		config = TestConfig()
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiter := getLimiter("forgot_password:"+clientIP, config)

		if !limiter.Allow() {
			retryAfter := calculateRetryAfter(config)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many password reset attempts. Please try again later.",
				"retry_after": retryAfter.String(),
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

	// Initialize limiters based on test mode
	if isTestMode() {
		config := TestConfig()
		limiters["login:test"] = rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst)
		limiters["forgot_password:test"] = rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst)
	}
}

func init() {
	ResetLimiters()
}
