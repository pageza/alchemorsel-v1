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
		recipeRepo := repositories.NewRecipeRepository(db)
		cuisineRepo := repositories.NewCuisineRepository(db)
		dietRepo := repositories.NewDietRepository(db)
		applianceRepo := repositories.NewApplianceRepository(db)
		tagRepo := repositories.NewTagRepository(db)

		// Initialize services
		userService := services.NewUserService(userRepo)
		cuisineService := services.NewCuisineService(cuisineRepo)
		dietService := services.NewDietService(dietRepo)
		applianceService := services.NewApplianceService(applianceRepo)
		tagService := services.NewTagService(tagRepo)
		recipeService := services.NewRecipeService(recipeRepo, cuisineService, dietService, applianceService, tagService)

		// Initialize handlers
		userHandler := handlers.NewUserHandler(userService)
		recipeHandler := handlers.NewRecipeHandler(recipeService)

		// Only add the rate limiter if DISABLE_RATE_LIMITER is not set to "true".
		if os.Getenv("DISABLE_RATE_LIMITER") != "true" {
			// Create a group for endpoints that should not have the global rate limiter
			noRateLimit := router.Group("")
			noRateLimit.Use(middleware.RateLimiter())
			{
				// Add all routes except login to the rate-limited group
				noRateLimit.GET("/v1/health", handlers.HealthCheck)
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
			secured.GET("/recipes", recipeHandler.ListRecipes)
			secured.GET("/recipes/:id", recipeHandler.GetRecipe)
			secured.POST("/recipes", recipeHandler.SaveRecipe)
			secured.PUT("/recipes/:id", recipeHandler.UpdateRecipe)
			secured.DELETE("/recipes/:id", recipeHandler.DeleteRecipe)
			secured.POST("/recipes/resolve", recipeHandler.ResolveRecipe)
			secured.POST("/recipes/:id/rate", recipeHandler.RateRecipe)
			secured.GET("/recipes/:id/ratings", recipeHandler.GetRecipeRatings)
			secured.GET("/recipes/search", recipeHandler.SearchRecipes)
		}
	}

	logger.Info("Router setup complete")
	return router
}
