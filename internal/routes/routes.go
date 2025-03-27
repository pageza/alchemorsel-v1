package routes

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/middleware"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/services"

	"strings"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRouter initializes and returns the Gin router with all routes configured
func SetupRouter(db *gorm.DB) *gin.Engine {
	// For integration tests, ensure we use the Postgres test database.
	if os.Getenv("INTEGRATION_TEST") == "true" && os.Getenv("DB_DRIVER") == "" {
		os.Setenv("DB_DRIVER", "postgres")
	}

	if os.Getenv("DB_DRIVER") == "postgres" {
		// For integration tests, set default Postgres env variables if not already set
		if os.Getenv("INTEGRATION_TEST") == "true" {
			if os.Getenv("POSTGRES_HOST") == "" {
				os.Setenv("POSTGRES_HOST", "localhost")
			}
			if os.Getenv("POSTGRES_PORT") == "" {
				os.Setenv("POSTGRES_PORT", "5432")
			}
			if os.Getenv("POSTGRES_USER") == "" {
				os.Setenv("POSTGRES_USER", "postgres")
			}
			if os.Getenv("POSTGRES_PASSWORD") == "" {
				os.Setenv("POSTGRES_PASSWORD", "postgres")
			}
			if os.Getenv("POSTGRES_DB") == "" {
				os.Setenv("POSTGRES_DB", "testdb")
			}
		}

		// Create UUID extension and run migrations
		sqlDB, err := db.DB()
		if err != nil {
			zap.S().Fatalf("failed to get underlying sql.DB: %v", err)
		}
		if _, err := sqlDB.ExecContext(context.Background(), "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" WITH SCHEMA public;"); err != nil {
			zap.S().Warnf("failed to create uuid-ossp extension: %v", err)
		}
		if err := repositories.RunMigrations(db); err != nil {
			zap.S().Fatalf("failed to run migrations: %v", err)
		}
	}

	router := gin.Default()
	// Disable trailing slash redirection to prevent 301 redirects on endpoints.
	router.RedirectTrailingSlash = false
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
		// Initialize repositories
		userRepo := repositories.NewUserRepository(db)
		recipeRepo := repositories.NewRecipeRepository(db)

		// Initialize services
		userService := services.NewUserService(userRepo)
		recipeService := services.NewRecipeService(recipeRepo)

		// Initialize handlers
		userHandler := handlers.NewUserHandler(userService)
		recipeHandler := handlers.NewRecipeHandler(recipeService)

		// Public user endpoints for registration, login and account management
		v1.POST("/users", userHandler.CreateUser)
		v1.POST("/users/login", middleware.LoginRateLimiter(), userHandler.LoginUser)
		v1.GET("/users/verify-email/:token", userHandler.VerifyEmail)
		v1.POST("/users/forgot-password", userHandler.ForgotPassword)
		v1.POST("/users/reset-password", userHandler.ResetPassword)
		v1.GET("/users/:id", userHandler.GetUser)

		// Recipe endpoints
		v1.GET("/recipes", recipeHandler.ListRecipes)
		v1.GET("/recipes/:id", recipeHandler.GetRecipe)
		v1.POST("/recipes", recipeHandler.SaveRecipe)
		v1.PUT("/recipes/:id", recipeHandler.UpdateRecipe)
		v1.DELETE("/recipes/:id", recipeHandler.DeleteRecipe)

		// Recipe resolution endpoint
		v1.POST("/recipes/resolve", recipeHandler.ResolveRecipe)

		// Group for endpoints that require authentication.
		secured := v1.Group("")
		secured.Use(middleware.AuthMiddleware())
		{
			secured.GET("/users/me", userHandler.GetCurrentUser)
			secured.PUT("/users/me", userHandler.UpdateCurrentUser)
			secured.PATCH("/users/me", userHandler.PatchCurrentUser)
			secured.DELETE("/users/me", userHandler.DeleteCurrentUser)
			secured.GET("/admin/users", userHandler.GetAllUsers)
		}

		// Health-check endpoint to support TestHealthCheck.
		v1.GET("/health", handlers.HealthCheck)
	}

	// For test or integration environments, clear users and insert a dummy user.
	if gin.Mode() == gin.TestMode || os.Getenv("INTEGRATION_TEST") == "true" {
		if err := db.Exec("DELETE FROM users").Error; err != nil {
			zap.S().Fatal("failed to clear users table")
		}
		dummyUser := models.User{
			ID:       "1",
			Name:     "Dummy User",
			Email:    "dummy@example.com",
			Password: "dummy", // Replace with a hashed password in production if needed.
		}
		if err := db.FirstOrCreate(&dummyUser, models.User{ID: "1"}).Error; err != nil {
			zap.S().Fatal("failed to create dummy user")
		}
	}

	return router
}
