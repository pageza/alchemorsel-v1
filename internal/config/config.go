package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database struct {
		Driver   string
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	Server struct {
		Port         int
		Timeout      time.Duration
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}
	RateLimit struct {
		RequestsPerSecond float64
		Burst             int
		ExpirationTTL     time.Duration
	}
	JWT struct {
		Secret          string
		ExpirationHours int
		RefreshHours    int
	}
	Email struct {
		Host     string
		Port     int
		Username string
		Password string
		From     string
	}
	Logging struct {
		Level  string
		Format string
		Output string
	}
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	cfg := &Config{}

	// Database defaults
	cfg.Database.Driver = getEnvOrDefault("DB_DRIVER", "postgres")
	cfg.Database.Host = getEnvOrDefault("DB_HOST", "localhost")
	cfg.Database.Port = getEnvIntOrDefault("DB_PORT", 5432)
	cfg.Database.User = getEnvOrDefault("DB_USER", "postgres")
	cfg.Database.Password = getEnvOrDefault("DB_PASSWORD", "postgres")
	cfg.Database.DBName = getEnvOrDefault("DB_NAME", "alchemorsel")
	cfg.Database.SSLMode = getEnvOrDefault("DB_SSL_MODE", "disable")

	// Server defaults
	cfg.Server.Port = getEnvIntOrDefault("SERVER_PORT", 8080)
	cfg.Server.Timeout = getEnvDurationOrDefault("SERVER_TIMEOUT", 30*time.Second)
	cfg.Server.ReadTimeout = getEnvDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second)
	cfg.Server.WriteTimeout = getEnvDurationOrDefault("SERVER_WRITE_TIMEOUT", 10*time.Second)

	// Rate limit defaults
	cfg.RateLimit.RequestsPerSecond = getEnvFloatOrDefault("RATE_LIMIT_REQUESTS", 5.0)
	cfg.RateLimit.Burst = getEnvIntOrDefault("RATE_LIMIT_BURST", 10)
	cfg.RateLimit.ExpirationTTL = getEnvDurationOrDefault("RATE_LIMIT_EXPIRATION", time.Hour)

	// JWT defaults
	cfg.JWT.Secret = getEnvOrDefault("JWT_SECRET", "your-secret-key")
	cfg.JWT.ExpirationHours = getEnvIntOrDefault("JWT_EXPIRATION_HOURS", 24)
	cfg.JWT.RefreshHours = getEnvIntOrDefault("JWT_REFRESH_HOURS", 168) // 7 days

	// Email defaults
	cfg.Email.Host = getEnvOrDefault("EMAIL_HOST", "smtp.gmail.com")
	cfg.Email.Port = getEnvIntOrDefault("EMAIL_PORT", 587)
	cfg.Email.Username = getEnvOrDefault("EMAIL_USERNAME", "")
	cfg.Email.Password = getEnvOrDefault("EMAIL_PASSWORD", "")
	cfg.Email.From = getEnvOrDefault("EMAIL_FROM", "noreply@alchemorsel.com")

	// Logging defaults
	cfg.Logging.Level = getEnvOrDefault("LOG_LEVEL", "info")
	cfg.Logging.Format = getEnvOrDefault("LOG_FORMAT", "json")
	cfg.Logging.Output = getEnvOrDefault("LOG_OUTPUT", "stdout")

	return cfg
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	switch c.Database.Driver {
	case "postgres":
		return c.getPostgresDSN()
	case "sqlite":
		return c.getSQLiteDSN()
	default:
		return c.getPostgresDSN()
	}
}

// getPostgresDSN returns the PostgreSQL connection string
func (c *Config) getPostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// getSQLiteDSN returns the SQLite connection string
func (c *Config) getSQLiteDSN() string {
	return "file:alchemorsel.db?cache=shared&mode=rwc"
}

// Helper functions for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

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
