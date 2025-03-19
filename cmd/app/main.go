package main

import (
	"log"
	"recipeservice/internal/config"
	"recipeservice/internal/routes"
)

func main() {
	// Load configuration from .env file
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Setup and start the Gin router
	router := routes.SetupRouter()
	// TODO: Configure any additional routes or middleware if needed

	// Start server (port can be read from config)
	if err := router.Run(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
