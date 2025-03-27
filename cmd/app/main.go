package main

import (
	"log"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/config"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/routes"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			// Optionally, re-panic if you want to ensure complete termination.
			// panic(r)
		}
	}()

	// Load configuration from .env file
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize the database connection with retry logic
	var err error
	maxAttempts := 10
	for i := 1; i <= maxAttempts; i++ {
		err = db.Init()
		if err == nil {
			break
		}
		log.Printf("Attempt %d: error initializing database: %v", i, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Error initializing database after %d attempts: %v", maxAttempts, err)
	}

	// Check and drop legacy constraint 'uni_users_email' if it exists
	if db.DB.Migrator().HasConstraint(&models.User{}, "uni_users_email") {
		log.Println("Legacy constraint 'uni_users_email' exists, dropping it...")
		if err := db.DB.Migrator().DropConstraint(&models.User{}, "uni_users_email"); err != nil {
			log.Printf("Error dropping legacy constraint: %v", err)
		} else {
			log.Println("Legacy constraint dropped successfully.")
		}
	} else {
		log.Println("Legacy constraint 'uni_users_email' does not exist; no drop needed.")
	}

	// Migrations disabled. Please run SQL migration scripts manually.

	// Setup and start the Gin router
	router := routes.SetupRouter()
	// TODO: Configure any additional routes or middleware if needed

	// Start server on explicit port :8080 and log the attempt
	log.Println("Starting server on :8080")
	err = router.Run(":8080")
	log.Printf("router.Run returned with error: %v", err)
	log.Println("Server exiting")
}
