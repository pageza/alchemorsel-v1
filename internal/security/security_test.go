package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTest(_ *testing.T) (*SecurityManager, *gin.Engine) {
	config := SecurityConfig{
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

	sm := NewSecurityManager(config, redisClient)
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
		user := &User{
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
		assert.Equal(t, user.Roles, claims["roles"])
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
		user := &User{
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
		user := &User{
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
			expected: "&lt;script&gt;alert('xss')&lt;/script&gt;",
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

	// Test rate limit
	for i := 0; i < sm.config.RateLimitRequests+1; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1"
		router.ServeHTTP(w, req)

		if i >= sm.config.RateLimitRequests {
			assert.Equal(t, 429, w.Code)
		} else {
			assert.Equal(t, 200, w.Code)
		}
	}
}
