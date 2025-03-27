package main

import (
	"log"
	"strings"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/config"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/migrations"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/routes"
	"gorm.io/gorm"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	// Load configuration from .env file
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize the database connection with retry logic
	var database *gorm.DB
	var err error
	maxAttempts := 10
	for i := 1; i <= maxAttempts; i++ {
		config := db.NewConfig()
		database, err = db.InitDB(config)
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
	if database.Migrator().HasConstraint(&models.User{}, "uni_users_email") {
		log.Println("Legacy constraint 'uni_users_email' exists, dropping it...")
		if err := database.Migrator().DropConstraint(&models.User{}, "uni_users_email"); err != nil {
			log.Printf("Error dropping legacy constraint: %v", err)
		} else {
			log.Println("Legacy constraint dropped successfully.")
		}
	} else {
		log.Println("Legacy constraint 'uni_users_email' does not exist; no drop needed.")
	}

	// Run migrations
	if err := migrations.RunMigrations(database); err != nil {
		if strings.Contains(err.Error(), "uni_users_email") {
			log.Printf("Ignoring legacy drop error: %v", err)
		} else {
			log.Fatalf("Error running migrations: %v", err)
		}
	}

	// Setup and start the Gin router with database dependency
	router := routes.SetupRouter(database)
	log.Println("Starting server on :8080")
	err = router.Run(":8080")
	log.Printf("router.Run returned with error: %v", err)
	log.Println("Server exiting")
}
