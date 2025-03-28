package db

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global database instance
var DB *gorm.DB

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConfig creates a new database configuration from environment variables
func NewConfig() *Config {
	return &Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  "disable",
	}
}

// InitDB initializes the database connection
func InitDB(config *Config) (*gorm.DB, error) {
	logger := zap.L()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	gormLogger := NewGormLogger(logger)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		logger.Error("failed to connect to database",
			zap.Error(err),
			zap.String("host", config.Host),
			zap.String("port", config.Port),
			zap.String("dbname", config.DBName))
		return nil, err
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to get database instance",
			zap.Error(err))
		return nil, err
	}

	err = sqlDB.Ping()
	if err != nil {
		logger.Error("failed to ping database",
			zap.Error(err))
		return nil, err
	}

	logger.Info("successfully connected to database",
		zap.String("host", config.Host),
		zap.String("port", config.Port),
		zap.String("dbname", config.DBName))

	// Set the global DB instance
	DB = db
	return db, nil
}
