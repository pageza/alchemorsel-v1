package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogConfig holds logging configuration
type LogConfig struct {
	LogDir            string
	MaxSize           int // MB
	MaxBackups        int
	MaxAge            int // days
	Compress          bool
	LogLevel          string
	RequestIDHeader   string
	LogFormat         string // json or text
	EnableConsole     bool
	EnableFile        bool
	EnableElastic     bool
	ElasticURL        string
	ElasticIndex      string
	EnableCompression bool
}

// Logger handles all logging operations
type Logger struct {
	config     LogConfig
	logger     *zap.Logger
	rotator    *lumberjack.Logger
	elastic    *ElasticLogger
	compressor *LogCompressor
	mu         sync.RWMutex
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	RequestID   string                 `json:"request_id"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields"`
	Stacktrace  string                 `json:"stacktrace,omitempty"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
}

// NewLogger creates a new logger instance
func NewLogger(config LogConfig) (*Logger, error) {
	l := &Logger{
		config: config,
	}

	// Initialize log rotation
	if config.EnableFile {
		rotator := &lumberjack.Logger{
			Filename:   filepath.Join(config.LogDir, "app.log"),
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
		l.rotator = rotator
	}

	// Initialize log compressor
	if config.EnableCompression {
		l.compressor = NewLogCompressor(config.LogDir)
	}

	// Initialize elastic logger if enabled
	if config.EnableElastic {
		elastic, err := NewElasticLogger(config.ElasticURL, config.ElasticIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize elastic logger: %w", err)
		}
		l.elastic = elastic
	}

	// Configure zap logger
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if config.LogFormat == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var cores []zapcore.Core

	// Add console core if enabled
	if config.EnableConsole {
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			zap.NewAtomicLevelAt(zapcore.DebugLevel),
		)
		cores = append(cores, consoleCore)
	}

	// Add file core if enabled
	if config.EnableFile {
		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(l.rotator),
			zap.NewAtomicLevelAt(zapcore.DebugLevel),
		)
		cores = append(cores, fileCore)
	}

	// Create multi-core logger
	core := zapcore.NewTee(cores...)
	l.logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return l, nil
}

// RequestIDMiddleware adds request ID to context and logs
func (l *Logger) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(l.config.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		// Store start time for duration calculation
		startTime := time.Now()

		// Log request start
		l.Info("Request started",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		c.Next()

		// Log request end with duration
		l.Info("Request completed",
			zap.String("request_id", requestID),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(startTime)),
		)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// WithContext returns a logger with context fields
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return l.logger.With(zap.String("request_id", requestID))
	}
	return l.logger
}

// RotateLogs rotates the current log file
func (l *Logger) RotateLogs() error {
	if l.rotator != nil {
		return l.rotator.Rotate()
	}
	return nil
}

// CompressLogs compresses old log files
func (l *Logger) CompressLogs() error {
	if l.compressor != nil {
		return l.compressor.CompressOldLogs()
	}
	return nil
}

// SearchLogs searches through log files
func (l *Logger) SearchLogs(query string, startTime, endTime time.Time) ([]LogEntry, error) {
	var results []LogEntry
	logFiles, err := filepath.Glob(filepath.Join(l.config.LogDir, "*.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to find log files: %w", err)
	}

	for _, file := range logFiles {
		entries, err := l.searchFile(file, query, startTime, endTime)
		if err != nil {
			return nil, err
		}
		results = append(results, entries...)
	}

	return results, nil
}

// searchFile searches a single log file
func (l *Logger) searchFile(filePath, query string, startTime, endTime time.Time) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []LogEntry
	decoder := json.NewDecoder(file)

	for {
		var entry LogEntry
		if err := decoder.Decode(&entry); err == io.EOF {
			break
		} else if err != nil {
			continue
		}

		if entry.Timestamp.After(startTime) && entry.Timestamp.Before(endTime) {
			if matchesQuery(entry, query) {
				results = append(results, entry)
			}
		}
	}

	return results, nil
}

// matchesQuery checks if a log entry matches the search query
func matchesQuery(entry LogEntry, query string) bool {
	// Search in message
	if contains(entry.Message, query) {
		return true
	}

	// Search in fields
	for _, value := range entry.Fields {
		if str, ok := value.(string); ok && contains(str, query) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// getLogLevel converts string level to zapcore.Level
func getLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
