package integration

import (
	"github.com/pageza/alchemorsel-v1/internal/logging"
)

// createTestLogger creates a logger for testing
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
