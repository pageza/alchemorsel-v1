package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pageza/alchemorsel-v1/internal/middleware"
)

// TestRealAuthentication tests the authentication middleware using a real JWT token.
func TestRealAuthentication(t *testing.T) {
	// Unset bypass flags so that real auth is enforced
	os.Unsetenv("DISABLE_AUTH")
	os.Unsetenv("INTEGRATION_TEST")

	// Set up a new Gin router with the AuthMiddleware
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		user, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no user found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": user})
	})

	// Use the default secret defined in AuthMiddleware if not set externally
	// In our test, we use "test-secret-key" if no JWT_SECRET is provided
	secret := "test-secret-key"
	token := generateTestToken(secret)

	req, err := http.NewRequest(http.MethodGet, "/protected", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error parsing JSON response: %v", err)
	}

	if _, ok := resp["user"]; !ok {
		t.Errorf("expected user object in response, got: %v", resp)
	}
}

// generateTestToken creates a JWT token signed with the provided secret for testing purposes.
func generateTestToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "real-test-user",
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return tokenString
}
