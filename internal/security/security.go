package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	JWTSecret          string
	JWTExpirationHours int
	APIKeyRotationDays int
	RateLimitRequests  int
	RateLimitWindow    time.Duration
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials   bool
	MaxAgeSeconds      int
}

// SecurityManager handles all security-related operations
type SecurityManager struct {
	config    SecurityConfig
	validator *validator.Validate
	redis     *redis.Client
	apiKeys   map[string]APIKeyInfo
}

// APIKeyInfo stores information about API keys
type APIKeyInfo struct {
	Key       string
	CreatedAt time.Time
	ExpiresAt time.Time
	UserID    string
	Scopes    []string
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config SecurityConfig, redisClient *redis.Client) *SecurityManager {
	return &SecurityManager{
		config:    config,
		validator: validator.New(),
		redis:     redisClient,
		apiKeys:   make(map[string]APIKeyInfo),
	}
}

// Input Validation
type ValidationError struct {
	Field string
	Tag   string
	Value interface{}
}

func (sm *SecurityManager) ValidateInput(input interface{}) []ValidationError {
	var errors []ValidationError
	err := sm.validator.Struct(input)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field: err.Field(),
				Tag:   err.Tag(),
				Value: err.Value(),
			})
		}
	}
	return errors
}

// CORS Middleware
func (sm *SecurityManager) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			allowed := false
			for _, allowedOrigin := range sm.config.AllowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}
			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
				c.Header("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
				c.Header("Access-Control-Expose-Headers", strings.Join(sm.config.ExposedHeaders, ", "))
				if sm.config.AllowCredentials {
					c.Header("Access-Control-Allow-Credentials", "true")
				}
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", sm.config.MaxAgeSeconds))
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// API Key Management
func (sm *SecurityManager) GenerateAPIKey(userID string, scopes []string) (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	apiKey := base64.URLEncoding.EncodeToString(key)

	info := APIKeyInfo{
		Key:       apiKey,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, sm.config.APIKeyRotationDays),
		UserID:    userID,
		Scopes:    scopes,
	}

	sm.apiKeys[apiKey] = info
	return apiKey, nil
}

func (sm *SecurityManager) ValidateAPIKey(key string) (*APIKeyInfo, error) {
	info, exists := sm.apiKeys[key]
	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}
	if time.Now().After(info.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}
	return &info, nil
}

// Authentication
type User struct {
	ID       string
	Username string
	Password string
	Roles    []string
}

func (sm *SecurityManager) GenerateToken(user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"roles":    user.Roles,
		"exp":      time.Now().Add(time.Hour * time.Duration(sm.config.JWTExpirationHours)).Unix(),
	})

	return token.SignedString([]byte(sm.config.JWTSecret))
}

func (sm *SecurityManager) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(sm.config.JWTSecret), nil
	})
}

// Authorization
func (sm *SecurityManager) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		token, err := sm.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		userRoles, ok := claims["roles"].([]interface{})
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid roles"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			for _, userRole := range userRoles {
				if roleStr, ok := userRole.(string); ok && role == roleStr {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Output Encoding
func (sm *SecurityManager) EncodeOutput(input string) string {
	// First encode ampersands to prevent double encoding
	input = strings.ReplaceAll(input, "&", "&amp;")

	// Then encode other special characters
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#39;")

	return input
}

// Security Headers Middleware
func (sm *SecurityManager) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

// Rate Limiting
func (sm *SecurityManager) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If Redis is not available, allow the request
		if sm.redis == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		ctx := c.Request.Context()

		// Get current count
		count, err := sm.redis.Get(ctx, key).Int()
		if err == redis.Nil {
			// Key doesn't exist, set it with initial count and expiration
			err = sm.redis.SetEx(ctx, key, 1, sm.config.RateLimitWindow).Err()
			if err != nil {
				c.Next() // On Redis error, allow the request
				return
			}
			c.Next()
			return
		}
		if err != nil {
			c.Next() // On Redis error, allow the request
			return
		}

		// Check if rate limit exceeded
		if count >= sm.config.RateLimitRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		// Increment count using INCR
		pipe := sm.redis.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, sm.config.RateLimitWindow) // Reset expiration on increment
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next() // On Redis error, allow the request
			return
		}

		c.Next()
	}
}

// Security Monitoring
type SecurityEvent struct {
	Type      string
	Timestamp time.Time
	IP        string
	UserID    string
	Details   map[string]interface{}
}

func (sm *SecurityManager) LogSecurityEvent(event SecurityEvent) error {
	// Implement security event logging
	// This could send events to a security monitoring service
	return nil
}

// DisableRedis disables Redis for testing purposes
func (sm *SecurityManager) DisableRedis() {
	sm.redis = nil
}

// SetRedisClient sets the Redis client for testing purposes
func (sm *SecurityManager) SetRedisClient(client *redis.Client) {
	sm.redis = client
}

// RateLimitConfig represents the rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// SetRateLimitConfig sets the rate limit configuration
func (sm *SecurityManager) SetRateLimitConfig(requests int, window time.Duration) {
	sm.config.RateLimitRequests = requests
	sm.config.RateLimitWindow = window
}

// GetRateLimitConfig returns the current rate limit configuration
func (sm *SecurityManager) GetRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Requests: sm.config.RateLimitRequests,
		Window:   sm.config.RateLimitWindow,
	}
}
