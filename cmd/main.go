package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Initialize the router
	router := gin.Default()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Initialize database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize DeepSeek client
	deepseekClient := repositories.NewDeepSeekClient("your-api-key", "your-api-url")

	// Initialize recipe cache
	recipeCache := repositories.NewRecipeCache(rdb)

	// Initialize handlers
	recipeHandler := handlers.NewRecipeHandler(database, recipeCache, deepseekClient)

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
