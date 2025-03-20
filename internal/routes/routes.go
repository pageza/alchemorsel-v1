package routes

import (
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the Gin routes for the API.
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// NEW: Add security headers middleware globally.
	router.Use(middleware.SecurityHeaders())

	// Grouping versioned API routes
	v1 := router.Group("/v1")
	{
		// Public user endpoints for registration and login.
		v1.POST("/users", handlers.CreateUser)
		v1.POST("/users/login", middleware.LoginRateLimiter(), handlers.LoginUser)
		v1.GET("/users/verify-email/:token", handlers.VerifyEmail)
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

			// Endpoint for deactivating (soft deleting) current user.
			secured.DELETE("/users/me", handlers.DeleteCurrentUser)

			// Admin-only endpoint: list users (or search/filter users).
			secured.GET("/admin/users", handlers.GetAllUsers)
		}
	}

	return router
}
