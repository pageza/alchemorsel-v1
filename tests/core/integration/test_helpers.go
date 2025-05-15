package integration

import (
	"log"
	"os"

	"github.com/pageza/alchemorsel-v1/internal/logging"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
)

func createTestLogger() *logging.Logger {
	config := logging.LogConfig{
		LogLevel:        "debug",
		LogFormat:       "text",
		EnableConsole:   true,
		EnableFile:      false,
		RequestIDHeader: "X-Request-ID",
	}
	logger, err := logging.NewLogger(config)
	if err != nil {
		panic(err)
	}
	return logger
}

func createTestRedisClient() *repositories.RedisClient {
	// For tests, try to connect to Redis, but don't fail if it's not available
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient, err := repositories.NewRedisClient(redisAddr)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v. Using nil client for tests.", err)
		return nil
	}
	return redisClient
}
