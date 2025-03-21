package routes

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/middleware"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"

	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
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
				rawToken := strings.TrimPrefix(authHeader, "Bearer ")
				secret := os.Getenv("JWT_SECRET")
				parsedToken, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
					return []byte(secret), nil
				})
				if err == nil && parsedToken.Valid {
					if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
						if id, ok := claims["id"].(string); ok {
							c.Set("currentUser", id)
						}
					}
				}
			}
			c.Next()
		})
	}

	// Grouping versioned API routes
	v1 := router.Group("/v1")
	{
		// Public user endpoints for registration and login.
		v1.POST("/users", handlers.CreateUser)
		v1.POST("/users/login", middleware.LoginRateLimiter(), handlers.LoginUser)
		v1.GET("/users/verify-email/:token", handlers.VerifyEmail)
		v1.POST("/users/forgot-password", handlers.ForgotPassword)
		v1.POST("/users/reset-password", handlers.ResetPassword)
		v1.GET("/users/:id", handlers.GetUser)
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

	// Always initialize the database connection.
	dsn := os.Getenv("DB_SOURCE")
	if dsn == "" {
		logrus.Fatal("DB_SOURCE environment variable not set")
	}
	if err := repositories.InitializeDB(dsn); err != nil {
		logrus.WithError(err).Fatal("failed to initialize database")
	}
	if err := repositories.AutoMigrate(); err != nil {
		logrus.WithError(err).Fatal("failed to migrate database")
	}

	// For test or integration environments, clear users and insert a dummy user.
	if gin.Mode() == gin.TestMode || os.Getenv("INTEGRATION_TEST") == "true" {
		if err := repositories.ClearUsers(); err != nil {
			logrus.WithError(err).Fatal("failed to clear users table")
		}
		dummyUser := models.User{
			ID:       "1",
			Name:     "Dummy User",
			Email:    "dummy@example.com",
			Password: "dummy", // Replace with a hashed password in production if needed.
		}
		if err := repositories.DB.FirstOrCreate(&dummyUser, models.User{ID: "1"}).Error; err != nil {
			logrus.WithError(err).Fatal("failed to create dummy user")
		}
	}

	return router
}
