package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
)

func main() {
	// Initialize the router
	router := gin.Default()

	// Initialize Redis client
	redisClient, err := repositories.NewRedisClient("redis:6379")
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}

	// Initialize database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize handlers
	recipeHandler := handlers.NewRecipeHandler(redisClient, database)

	// Setup routes
	v1 := router.Group("/v1")
	{
		// Recipe endpoints
		v1.POST("/recipes", recipeHandler.GenerateRecipe)
	}

	// Start the server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
