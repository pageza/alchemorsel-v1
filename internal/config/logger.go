package config

import (
	"os"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SetupLogger configures the application logger based on configuration
func SetupLogger(cfg *Config) (*zap.Logger, error) {
	var config zap.Config

	// Set log level based on configuration
	level, err := zapcore.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Configure based on environment
	if cfg.Logging.Format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Set the log level
	config.Level = zap.NewAtomicLevelAt(level)

	// Configure output
	if cfg.Logging.Output != "stdout" {
		config.OutputPaths = []string{cfg.Logging.Output}
	}

	// Add caller and stack trace for development
	if cfg.Logging.Format != "json" {
		config.Development = true
		config.DisableStacktrace = false
		config.DisableCaller = false
	}

	// Create the logger
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// Replace the global logger
	zap.ReplaceGlobals(logger)

	// Also set up logrus for compatibility
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.Level(level))

	return logger, nil
}

// GetLogger returns the configured logger instance
func GetLogger() *zap.Logger {
	return zap.L()
}
