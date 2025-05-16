package routes

import (
	"context"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/logging"
	"github.com/pageza/alchemorsel-v1/internal/middleware"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// readSecretFile reads a secret from the Docker secrets directory
func readSecretFile(secretName string) (string, error) {
	// Try Docker secrets path first
	dockerPath := "/run/secrets/" + secretName
	data, err := os.ReadFile(dockerPath)
	if err != nil {
		// Fall back to local secrets directory
		localPath := "./secrets/" + secretName + ".txt"
		data, err = os.ReadFile(localPath)
		if err != nil {
			return "", err
		}
	}
	// Remove any trailing newlines but preserve the rest of the content exactly
	return strings.TrimRight(string(data), "\n\r"), nil
}

// SetupRouter initializes and returns the Gin router with all routes configured
func SetupRouter(db *gorm.DB, logger *logging.Logger, redisClient *repositories.RedisClient) *gin.Engine {
	logger.Info("Starting router setup...")

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

		logger.Info("Setting up database extensions and migrations...")
		// Create UUID extension
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("Failed to get underlying sql.DB", zap.Error(err))
		} else {
			if _, err := sqlDB.ExecContext(context.Background(), "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" WITH SCHEMA public;"); err != nil {
				logger.Error("Failed to create uuid-ossp extension", zap.Error(err))
			}
		}
	}

	logger.Info("Initializing Gin router...")
	router := gin.Default()
	// Disable trailing slash redirection to prevent 301 redirects on endpoints.
	router.RedirectTrailingSlash = false
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(logger.RequestIDMiddleware())

	// Always add security headers unless explicitly disabled.
	if os.Getenv("DISABLE_SECURITY_HEADERS") != "true" {
		router.Use(middleware.SecurityHeaders())
	}

	logger.Info("Setting up routes...")
	// Grouping versioned API routes
	v1 := router.Group("/v1")
	{
		// Initialize repositories
		userRepo := repositories.NewUserRepository(db)

		// Initialize services
		userService := services.NewUserService(userRepo)

		// Initialize handlers
		userHandler := handlers.NewUserHandler(userService)

		// Initialize DeepSeek client
		deepseekKey, err := readSecretFile("deepseek_api_key")
		if err != nil {
			logger.Error("Failed to read DeepSeek API key", zap.Error(err))
		}
		deepseekURL, err := readSecretFile("deepseek_api_url")
		if err != nil {
			logger.Error("Failed to read DeepSeek API URL", zap.Error(err))
		}
		deepseekClient := repositories.NewDeepSeekClient(deepseekKey, deepseekURL)

		// Initialize recipe cache with Redis client
		recipeCache := repositories.NewRecipeCache(redisClient.GetClient())

		// Initialize recipe handler with all required dependencies
		recipeHandler := handlers.NewRecipeHandler(db, recipeCache, deepseekClient)

		// Only add the rate limiter if DISABLE_RATE_LIMITER is not set to "true".
		if os.Getenv("DISABLE_RATE_LIMITER") != "true" {
			// Create a group for endpoints that should not have the global rate limiter
			noRateLimit := router.Group("")
			noRateLimit.Use(middleware.RateLimiter())
			{
				// Add health check to the rate-limited group
				noRateLimit.GET("/v1/health", func(c *gin.Context) {
					c.JSON(200, gin.H{"status": "ok"})
				})
			}
		}

		// Public user endpoints for registration, login and account management
		v1.POST("/users", middleware.RateLimiter(), userHandler.CreateUser)
		v1.POST("/users/login", middleware.LoginRateLimiter(), userHandler.LoginUser)
		v1.GET("/users/verify-email/:token", userHandler.VerifyEmail)
		v1.POST("/users/forgot-password", userHandler.ForgotPassword)
		v1.POST("/users/reset-password", userHandler.ResetPassword)
		v1.GET("/users/:id", userHandler.GetUser)

		// Group for endpoints that require authentication.
		secured := v1.Group("")
		secured.Use(middleware.AuthMiddleware())
		{
			// User endpoints
			secured.GET("/users/me", userHandler.GetCurrentUser)
			secured.PUT("/users/me", userHandler.UpdateCurrentUser)
			secured.PATCH("/users/me", userHandler.PatchCurrentUser)
			secured.DELETE("/users/me", userHandler.DeleteCurrentUser)
			secured.GET("/admin/users", userHandler.GetAllUsers)

			// Recipe endpoints
			secured.POST("/recipes", recipeHandler.GenerateRecipe)
			secured.POST("/recipes/:id/approve", recipeHandler.ApproveRecipe)
			secured.POST("/recipes/:id/modify", recipeHandler.StartRecipeModification)
			secured.POST("/recipes/:id/modify-with-ai", recipeHandler.ModifyRecipeWithAI)
			secured.POST("/recipes/search", recipeHandler.SearchRecipes)
			secured.PUT("/recipes/:id", recipeHandler.ModifyRecipe)
			secured.GET("/recipes/:id", recipeHandler.GetRecipe)
			secured.GET("/recipes", recipeHandler.ListRecipes)
		}
	}

	logger.Info("Router setup complete")
	return router
}
