package integration

import (
	"os"

	"github.com/pageza/alchemorsel-v1/internal/logging"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/redis/go-redis/v9"
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
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	redisClient := repositories.NewRedisClient(rdb)
	return redisClient
}
