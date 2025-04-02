package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// Config holds all configuration for the application
type Config struct {
	Environment Environment
	Database    DatabaseConfig
	Server      ServerConfig
	RateLimit   RateLimitConfig
	JWT         JWTConfig
	Email       EmailConfig
	Logging     LoggingConfig
}

// DatabaseConfig holds database configuration settings
type DatabaseConfig struct {
	Driver            string        `env:"DB_DRIVER" envDefault:"postgres" validate:"required,oneof=postgres sqlite"`
	Host              string        `env:"POSTGRES_HOST" envDefault:"localhost" validate:"required"`
	Port              int           `env:"POSTGRES_PORT" envDefault:"5432" validate:"required,min=1,max=65535"`
	User              string        `env:"POSTGRES_USER" envDefault:"postgres" validate:"required"`
	Password          string        `env:"POSTGRES_PASSWORD" envDefault:"postgres" validate:"required"`
	DBName            string        `env:"POSTGRES_DB" envDefault:"alchemorsel" validate:"required"`
	SSLMode           string        `env:"DB_SSL_MODE" envDefault:"disable" validate:"required,oneof=disable require verify-ca verify-full"`
	BackupDir         string        `env:"DB_BACKUP_DIR" envDefault:"/var/backups/db" validate:"required"`
	BackupRetention   time.Duration `env:"DB_BACKUP_RETENTION" envDefault:"168h" validate:"required"` // 7 days
	BackupInterval    time.Duration `env:"DB_BACKUP_INTERVAL" envDefault:"24h" validate:"required"`   // 1 day
	BackupCompression bool          `env:"DB_BACKUP_COMPRESSION" envDefault:"true"`
}

// ServerConfig holds server configuration settings
type ServerConfig struct {
	Port         int           `env:"SERVER_PORT" envDefault:"8080" validate:"required,min=1,max=65535"`
	Timeout      time.Duration `env:"SERVER_TIMEOUT" envDefault:"30s" validate:"required"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"10s" validate:"required"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"10s" validate:"required"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64       `env:"RATE_LIMIT_REQUESTS" envDefault:"5.0" validate:"required,min=0"`
	Burst             int           `env:"RATE_LIMIT_BURST" envDefault:"10" validate:"required,min=1"`
	ExpirationTTL     time.Duration `env:"RATE_LIMIT_EXPIRATION" envDefault:"1h" validate:"required"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string `env:"JWT_SECRET" envDefault:"your-secret-key" validate:"required,min=32"`
	ExpirationHours int    `env:"JWT_EXPIRATION_HOURS" envDefault:"24" validate:"required,min=1"`
	RefreshHours    int    `env:"JWT_REFRESH_HOURS" envDefault:"168" validate:"required,min=1"` // 7 days
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Host     string `env:"EMAIL_HOST" envDefault:"smtp.gmail.com" validate:"required,hostname"`
	Port     int    `env:"EMAIL_PORT" envDefault:"587" validate:"required,min=1,max=65535"`
	Username string `env:"EMAIL_USERNAME" envDefault:"" validate:"required,email"`
	Password string `env:"EMAIL_PASSWORD" envDefault:"" validate:"required"`
	From     string `env:"EMAIL_FROM" envDefault:"noreply@alchemorsel.com" validate:"required,email"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info" validate:"required,oneof=debug info warn error"`
	Format string `env:"LOG_FORMAT" envDefault:"json" validate:"required,oneof=json text"`
	Output string `env:"LOG_OUTPUT" envDefault:"stdout" validate:"required"`
}

// NewConfig creates a new Config with default values and validates the configuration
func NewConfig() (*Config, error) {
	env := Environment(getEnvOrDefault("APP_ENV", "development"))
	if !isValidEnvironment(env) {
		return nil, fmt.Errorf("invalid environment: %s", env)
	}

	cfg := &Config{
		Environment: env,
	}

	// Load configuration from environment
	if err := cfg.loadFromEnv(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() error {
	// Database configuration
	c.Database.Driver = getEnvOrDefault("DB_DRIVER", "postgres")
	c.Database.Host = getEnvOrDefault("POSTGRES_HOST", "localhost")
	c.Database.Port = getEnvIntOrDefault("POSTGRES_PORT", 5432)
	c.Database.User = getEnvOrDefault("POSTGRES_USER", "postgres")
	c.Database.Password = getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	c.Database.DBName = getEnvOrDefault("POSTGRES_DB", "alchemorsel")
	c.Database.SSLMode = getEnvOrDefault("DB_SSL_MODE", "disable")
	c.Database.BackupDir = getEnvOrDefault("DB_BACKUP_DIR", "/var/backups/db")
	c.Database.BackupRetention = getEnvDurationOrDefault("DB_BACKUP_RETENTION", 7*24*time.Hour)
	c.Database.BackupInterval = getEnvDurationOrDefault("DB_BACKUP_INTERVAL", 24*time.Hour)
	c.Database.BackupCompression = getEnvBoolOrDefault("DB_BACKUP_COMPRESSION", true)

	// Server configuration
	c.Server.Port = getEnvIntOrDefault("SERVER_PORT", 8080)
	c.Server.Timeout = getEnvDurationOrDefault("SERVER_TIMEOUT", 30*time.Second)
	c.Server.ReadTimeout = getEnvDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second)
	c.Server.WriteTimeout = getEnvDurationOrDefault("SERVER_WRITE_TIMEOUT", 10*time.Second)

	// Rate limit configuration
	c.RateLimit.RequestsPerSecond = getEnvFloatOrDefault("RATE_LIMIT_REQUESTS", 5.0)
	c.RateLimit.Burst = getEnvIntOrDefault("RATE_LIMIT_BURST", 10)
	c.RateLimit.ExpirationTTL = getEnvDurationOrDefault("RATE_LIMIT_EXPIRATION", time.Hour)

	// JWT configuration
	c.JWT.Secret = getEnvOrDefault("JWT_SECRET", "your-secret-key")
	c.JWT.ExpirationHours = getEnvIntOrDefault("JWT_EXPIRATION_HOURS", 24)
	c.JWT.RefreshHours = getEnvIntOrDefault("JWT_REFRESH_HOURS", 168)
	log.Printf("Loaded JWT_SECRET length: %d", len(c.JWT.Secret))
	log.Printf("Loaded DB configuration: Host=%s, Port=%d, DBName=%s", c.Database.Host, c.Database.Port, c.Database.DBName)

	// Email configuration
	c.Email.Host = getEnvOrDefault("EMAIL_HOST", "smtp.gmail.com")
	c.Email.Port = getEnvIntOrDefault("EMAIL_PORT", 587)
	c.Email.Username = getEnvOrDefault("EMAIL_USERNAME", "")
	c.Email.Password = getEnvOrDefault("EMAIL_PASSWORD", "")
	c.Email.From = getEnvOrDefault("EMAIL_FROM", "noreply@alchemorsel.com")

	// Logging configuration
	c.Logging.Level = getEnvOrDefault("LOG_LEVEL", "info")
	c.Logging.Format = getEnvOrDefault("LOG_FORMAT", "json")
	c.Logging.Output = getEnvOrDefault("LOG_OUTPUT", "stdout")

	return nil
}

// validate validates the configuration
func (c *Config) validate() error {
	// Validate database configuration
	if c.Database.Driver != "postgres" && c.Database.Driver != "sqlite" {
		return fmt.Errorf("invalid database driver: %s", c.Database.Driver)
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	if c.Database.SSLMode != "disable" && c.Database.SSLMode != "require" &&
		c.Database.SSLMode != "verify-ca" && c.Database.SSLMode != "verify-full" {
		return fmt.Errorf("invalid SSL mode: %s", c.Database.SSLMode)
	}

	// Validate server configuration
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Validate rate limit configuration
	if c.RateLimit.RequestsPerSecond < 0 {
		return fmt.Errorf("invalid rate limit requests per second: %f", c.RateLimit.RequestsPerSecond)
	}
	if c.RateLimit.Burst < 1 {
		return fmt.Errorf("invalid rate limit burst: %d", c.RateLimit.Burst)
	}

	// Validate JWT configuration
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}
	if c.JWT.ExpirationHours < 1 {
		return fmt.Errorf("invalid JWT expiration hours: %d", c.JWT.ExpirationHours)
	}
	if c.JWT.RefreshHours < 1 {
		return fmt.Errorf("invalid JWT refresh hours: %d", c.JWT.RefreshHours)
	}

	// Validate email configuration
	if c.Email.Port < 1 || c.Email.Port > 65535 {
		return fmt.Errorf("invalid email port: %d", c.Email.Port)
	}

	// Validate logging configuration
	if c.Logging.Level != "debug" && c.Logging.Level != "info" &&
		c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}
	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}

	return nil
}

// isValidEnvironment checks if the environment is valid
func isValidEnvironment(env Environment) bool {
	switch env {
	case Development, Staging, Production:
		return true
	default:
		return false
	}
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

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// LoadConfig loads configuration from environment files and environment variables
func LoadConfig() error {
	// Try to load .env.development first
	if err := godotenv.Load(".env.development"); err != nil {
		// If .env.development doesn't exist, try .env
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, relying on environment variables")
		}
	}
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

// GetPostgresDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetSQLiteDSN returns the SQLite connection string
func (c *DatabaseConfig) GetSQLiteDSN() string {
	return "alchemorsel.db"
}
