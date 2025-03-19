package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig loads environment variables from a .env file.
func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}
	// TODO: Load and validate required environment variables as needed
	return nil
}

// GetEnv retrieves the value of the environment variable named by the key or returns defaultValue if not set.
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
