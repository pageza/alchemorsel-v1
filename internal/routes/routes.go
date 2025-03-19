package routes

import (
	"github.com/pageza/alchemorsel-v1/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the Gin routes for the API.
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Grouping versioned API routes
	v1 := router.Group("/v1")
	{
		// User endpoints
		v1.POST("/users", handlers.CreateUser)
		v1.POST("/users/login", handlers.LoginUser)
		v1.GET("/users/:id", handlers.GetUser)
		v1.PUT("/users/:id", handlers.UpdateUser)
		v1.DELETE("/users/:id", handlers.DeleteUser)

		// Recipe endpoints
		v1.GET("/recipes", handlers.ListRecipes)
		v1.GET("/recipes/:id", handlers.GetRecipe)
		v1.POST("/recipes", handlers.CreateRecipe)
		v1.PUT("/recipes/:id", handlers.UpdateRecipe)
		v1.DELETE("/recipes/:id", handlers.DeleteRecipe)

		// Recipe resolution endpoint
		v1.POST("/recipes/resolve", handlers.ResolveRecipe)
	}

	return router
}
