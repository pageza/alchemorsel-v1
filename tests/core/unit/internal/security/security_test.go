package security_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pageza/alchemorsel-v1/internal/security"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTest(_ *testing.T) (*security.SecurityManager, *gin.Engine) {
	config := security.SecurityConfig{
		JWTSecret:          "test-secret",
		JWTExpirationHours: 24,
		APIKeyRotationDays: 30,
		RateLimitRequests:  100,
		RateLimitWindow:    time.Minute,
		AllowedOrigins:     []string{"http://localhost:3000"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:     []string{"Authorization", "Content-Type"},
		ExposedHeaders:     []string{"Content-Length"},
		AllowCredentials:   true,
		MaxAgeSeconds:      3600,
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	sm := security.NewSecurityManager(config, redisClient)
	gin.SetMode(gin.TestMode)
	router := gin.New()

	return sm, router
}

func TestInputValidation(t *testing.T) {
	sm, _ := setupTest(t)

	type TestStruct struct {
		Name  string `validate:"required,min=3"`
		Email string `validate:"required,email"`
	}

	tests := []struct {
		name     string
		input    TestStruct
		wantErr  bool
		errCount int
	}{
		{
			name: "valid input",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid input",
			input: TestStruct{
				Name:  "Jo", // too short
				Email: "invalid-email",
			},
			wantErr:  true,
			errCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := sm.ValidateInput(tt.input)
			if tt.wantErr {
				assert.Len(t, errors, tt.errCount)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	sm, router := setupTest(t)
	router.Use(sm.CORSMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	tests := []struct {
		name           string
		origin         string
		wantAllow      bool
		wantStatusCode int
	}{
		{
			name:           "allowed origin",
			origin:         "http://localhost:3000",
			wantAllow:      true,
			wantStatusCode: 200,
		},
		{
			name:           "disallowed origin",
			origin:         "http://malicious.com",
			wantAllow:      false,
			wantStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantAllow {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestAPIKeyManagement(t *testing.T) {
	sm, _ := setupTest(t)

	t.Run("generate and validate API key", func(t *testing.T) {
		userID := "test-user"
		scopes := []string{"read", "write"}

		key, err := sm.GenerateAPIKey(userID, scopes)
		assert.NoError(t, err)
		assert.NotEmpty(t, key)

		info, err := sm.ValidateAPIKey(key)
		assert.NoError(t, err)
		assert.Equal(t, userID, info.UserID)
		assert.Equal(t, scopes, info.Scopes)
	})

	t.Run("invalid API key", func(t *testing.T) {
		_, err := sm.ValidateAPIKey("invalid-key")
		assert.Error(t, err)
	})
}

func TestAuthentication(t *testing.T) {
	sm, _ := setupTest(t)

	t.Run("generate and validate token", func(t *testing.T) {
		user := &security.User{
			ID:       "test-user",
			Username: "testuser",
			Roles:    []string{"user"},
		}

		token, err := sm.GenerateToken(user)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		validatedToken, err := sm.ValidateToken(token)
		assert.NoError(t, err)
		assert.True(t, validatedToken.Valid)

		claims, ok := validatedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, user.ID, claims["user_id"])
		assert.Equal(t, user.Username, claims["username"])

		// Convert interface{} to []string for comparison
		roles, ok := claims["roles"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, roles, 1)
		assert.Equal(t, "user", roles[0].(string))
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := sm.ValidateToken("invalid-token")
		assert.Error(t, err)
	})
}

func TestAuthorization(t *testing.T) {
	sm, router := setupTest(t)
	router.Use(sm.RequireRole("admin"))

	router.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("authorized access", func(t *testing.T) {
		user := &security.User{
			ID:       "test-user",
			Username: "testuser",
			Roles:    []string{"admin"},
		}

		token, _ := sm.GenerateToken(user)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		user := &security.User{
			ID:       "test-user",
			Username: "testuser",
			Roles:    []string{"user"},
		}

		token, _ := sm.GenerateToken(user)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
	})
}

func TestOutputEncoding(t *testing.T) {
	sm, _ := setupTest(t)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "encode HTML",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "no encoding needed",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.EncodeOutput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	sm, router := setupTest(t)
	router.Use(sm.SecurityHeadersMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	headers := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Strict-Transport-Security",
		"Content-Security-Policy",
		"Referrer-Policy",
		"Permissions-Policy",
	}

	for _, header := range headers {
		assert.NotEmpty(t, w.Header().Get(header))
	}
}

func TestRateLimiting(t *testing.T) {
	sm, router := setupTest(t)
	router.Use(sm.RateLimitMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Test rate limiting with Redis disabled (should allow all requests)
	t.Run("rate limiting with Redis disabled", func(t *testing.T) {
		sm.DisableRedis()
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1"
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}
	})

	// Test rate limiting with Redis enabled
	t.Run("rate limiting with Redis enabled", func(t *testing.T) {
		// Create a Redis client for testing
		redisClient := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			DB:   1, // Use a different DB for testing
		})

		// Test Redis connection
		ctx := context.Background()
		_, err := redisClient.Ping(ctx).Result()
		if err != nil {
			t.Skip("Redis not available, skipping rate limit test")
		}

		sm.SetRedisClient(redisClient)
		defer redisClient.Close()

		// Clear any existing rate limit data
		redisClient.FlushDB(ctx)

		// Set a lower rate limit for testing
		sm.SetRateLimitConfig(5, time.Second*10)

		// Make requests up to the limit
		rateLimitConfig := sm.GetRateLimitConfig()
		for i := 0; i < rateLimitConfig.Requests; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1"
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code, "Request %d should succeed", i+1)
		}

		// Wait a moment to ensure rate limit is enforced
		time.Sleep(time.Millisecond * 100)

		// Next request should be rate limited
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request should be rate limited")

		// Verify the response body
		var response map[string]string
		err = json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "rate limit exceeded", response["error"])
	})
}
