package routes

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/logging"
	"github.com/pageza/alchemorsel-v1/internal/middleware"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRouter initializes and returns the Gin router with all routes configured
func SetupRouter(db *gorm.DB, logger *logging.Logger) *gin.Engine {
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
		// Create UUID extension and run migrations
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("Failed to get underlying sql.DB", zap.Error(err))
		} else {
			if _, err := sqlDB.ExecContext(context.Background(), "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" WITH SCHEMA public;"); err != nil {
				logger.Error("Failed to create uuid-ossp extension", zap.Error(err))
			}
			if err := repositories.RunMigrations(db); err != nil {
				logger.Error("Failed to run migrations", zap.Error(err))
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

	// Only add the rate limiter if DISABLE_RATE_LIMITER is not set to "true".
	if os.Getenv("DISABLE_RATE_LIMITER") != "true" {
		router.Use(middleware.RateLimiter())
	}

	logger.Info("Setting up routes...")
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

	logger.Info("Router setup complete")
	return router
}
