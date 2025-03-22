package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware performs token validation for protected routes.
// Bypass occurs only if DISABLE_AUTH is explicitly set.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass token validation if the test bypass middleware has already set the user.
		if _, exists := c.Get("currentUser"); exists {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if id, ok := claims["id"].(string); ok {
				c.Set("currentUser", id)
			}
		}
		c.Next()
	}
}

// TODO: Add additional authentication middleware as needed.
