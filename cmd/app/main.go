package main

import (
	"log"

	"github.com/pageza/alchemorsel-v1/internal/config"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/migrations"
	"github.com/pageza/alchemorsel-v1/internal/routes"
)

func main() {
	// Load configuration from .env file
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize the database connection
	if err := db.Init(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Run database migrations before starting the server
	if err := migrations.RunMigrations(); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// Setup and start the Gin router
	router := routes.SetupRouter()
	// TODO: Configure any additional routes or middleware if needed

	// Start server (port can be read from config)
	if err := router.Run(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
