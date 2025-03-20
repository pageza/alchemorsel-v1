package routes

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/middleware"

	"strings"
)

// SetupRouter sets up the Gin routes for the API.
func SetupRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Always add security headers unless explicitly disabled.
	if os.Getenv("DISABLE_SECURITY_HEADERS") != "true" {
		router.Use(middleware.SecurityHeaders())
	}

	// Only add the rate limiter if DISABLE_RATE_LIMITER is not set to "true".
	if os.Getenv("DISABLE_RATE_LIMITER") != "true" {
		router.Use(middleware.RateLimiter())
	}

	// Conditionally add test authentication bypass middleware when in test mode.
	if gin.Mode() == gin.TestMode {
		router.Use(func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				// For testing purposes, assume the token itself is the user ID.
				token := strings.TrimPrefix(authHeader, "Bearer ")
				c.Set("currentUser", token)
			}
			c.Next()
		})
	}

	// Grouping versioned API routes
	v1 := router.Group("/v1")
	{
		// Public user endpoints for registration and login.
		v1.POST("/users", handlers.CreateUser)
		v1.POST("/users/login", handlers.LoginUser)
		v1.GET("/users/verify-email/:token", handlers.VerifyEmail)
		v1.POST("/users/forgot-password", handlers.ForgotPassword)
		v1.POST("/users/reset-password", handlers.ResetPassword)
		// (Optional) In the future, we might add public endpoints for user lookup with proper measures.

		// Recipe endpoints
		v1.GET("/recipes", handlers.ListRecipes)
		v1.GET("/recipes/:id", handlers.GetRecipe)
		v1.POST("/recipes", handlers.CreateRecipe)
		v1.PUT("/recipes/:id", handlers.UpdateRecipe)
		v1.DELETE("/recipes/:id", handlers.DeleteRecipe)

		// Recipe resolution endpoint
		v1.POST("/recipes/resolve", handlers.ResolveRecipe)

		// Group for endpoints that require authentication.
		secured := v1.Group("")
		secured.Use(middleware.AuthMiddleware())
		{
			// Regular endpoint: current user's profile.
			secured.GET("/users/me", handlers.GetCurrentUser)

			// Endpoint for updating current user's information.
			secured.PUT("/users/me", handlers.UpdateCurrentUser)

			// NEW: Add PATCH endpoint to update current user partially.
			secured.PATCH("/users/me", handlers.PatchCurrentUser)

			// Endpoint for deactivating (soft deleting) current user.
			secured.DELETE("/users/me", handlers.DeleteCurrentUser)

			// Admin-only endpoint: list users (or search/filter users).
			secured.GET("/admin/users", handlers.GetAllUsers)
		}

		// Health-check endpoint to support TestHealthCheck.
		v1.GET("/health", handlers.HealthCheck)
	}

	return router
}
