package main

import (
	"log"
	"os"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/config"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/logging"
	"github.com/pageza/alchemorsel-v1/internal/migrations"
	"github.com/pageza/alchemorsel-v1/internal/routes"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {


	// Initialize logger with console-only output
	logConfig := logging.LogConfig{
		LogLevel:      "debug",
		LogFormat:     "json",
		EnableConsole: true,
		EnableFile:    false, // Disable file logging
	}
	logger, err := logging.NewLogger(logConfig)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered from panic", zap.Any("panic", r))
		}
	}()

	logger.Info("Starting application...")

	// Load configuration from .env file
	if err := config.LoadConfig(); err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
	}
	logger.Info("Configuration loaded successfully")

	// Build configuration and DSN using the config package
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("Error creating configuration", zap.Error(err))
	}



	dsn := cfg.GetDSN()
	logger.Info("DSN constructed", zap.String("DSN", dsn))

	// Initialize the database connection with retry logic
	var database *gorm.DB
	maxAttempts := 10
	for i := 1; i <= maxAttempts; i++ {
		// Use db.NewConfig() as provided
		dbConfig := db.NewConfig()
		database, err = db.InitDB(dbConfig)
		if err == nil {
			logger.Info("Successfully connected to database")
			break
		}
		logger.Warn("Failed to connect to database",
			zap.Int("attempt", i),
			zap.Int("maxAttempts", maxAttempts),
			zap.Error(err))
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		logger.Fatal("Error initializing database after max attempts",
			zap.Int("maxAttempts", maxAttempts),
			zap.Error(err))
	}

	// Run migrations
	if err := migrations.RunMigrations(database); err != nil {
		logger.Warn("Migration warning (continuing)", zap.Error(err))
	}
	logger.Info("Migrations completed")

	// Setup and start the Gin router with database dependency
	logger.Info("Setting up router...")
	router := routes.SetupRouter(database, logger)
	logger.Info("Router setup complete")

	logger.Info("Starting server", zap.String("address", "0.0.0.0:8080"))
	logger.Debug("Server configuration",
		zap.String("host", "0.0.0.0"),
		zap.String("port", "8080"),
		zap.Any("routes", router.Routes()))

	if err := router.Run("0.0.0.0:8080"); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
	logger.Info("Server exiting")
}
