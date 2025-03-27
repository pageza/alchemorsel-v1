package repositories

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbInstance *gorm.DB
	log        = zap.L().Named("db")
)

// PoolStats represents database connection pool statistics
type PoolStats struct {
	MaxOpenConnections int           // Maximum number of open connections
	OpenConnections    int           // Current number of open connections
	InUseConnections   int           // Number of connections currently in use
	IdleConnections    int           // Number of idle connections
	WaitCount          int64         // Total number of connections waited for
	WaitDuration       time.Duration // Total time blocked waiting for a new connection
}

// GetPoolStats returns the current connection pool statistics
func GetPoolStats() (*PoolStats, error) {
	if dbInstance == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	stats := &PoolStats{
		MaxOpenConnections: sqlDB.Stats().MaxOpenConnections,
		OpenConnections:    sqlDB.Stats().OpenConnections,
		InUseConnections:   sqlDB.Stats().InUse,
		IdleConnections:    sqlDB.Stats().Idle,
		WaitCount:          sqlDB.Stats().WaitCount,
		WaitDuration:       sqlDB.Stats().WaitDuration,
	}

	return stats, nil
}

// MonitorPool starts monitoring the connection pool and logs statistics at specified intervals
func MonitorPool(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, err := GetPoolStats()
			if err != nil {
				log.Error("Failed to get pool stats", zap.Error(err))
				continue
			}

			// Log pool statistics
			log.Info("Database connection pool stats",
				zap.Int("max_open_connections", stats.MaxOpenConnections),
				zap.Int("open_connections", stats.OpenConnections),
				zap.Int("in_use_connections", stats.InUseConnections),
				zap.Int("idle_connections", stats.IdleConnections),
				zap.Int64("wait_count", stats.WaitCount),
				zap.Duration("wait_duration", stats.WaitDuration),
			)

			// Alert if pool is near capacity
			if float64(stats.OpenConnections) > float64(stats.MaxOpenConnections)*0.8 {
				log.Warn("Database connection pool near capacity",
					zap.Int("open_connections", stats.OpenConnections),
					zap.Int("max_open_connections", stats.MaxOpenConnections),
				)
			}

			// Alert if wait time is high
			if stats.WaitDuration > time.Second*5 {
				log.Warn("High database connection wait time",
					zap.Duration("wait_duration", stats.WaitDuration),
					zap.Int64("wait_count", stats.WaitCount),
				)
			}
		}
	}
}

// RetryConfig holds configuration for connection retries
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	MaxElapsedTime  time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      5,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		MaxElapsedTime:  30 * time.Second,
	}
}

// CircuitBreaker implements a simple circuit breaker pattern
type CircuitBreaker struct {
	failures    int32
	lastFailure time.Time
	threshold   int32
	timeout     time.Duration
	mu          sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int32, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
	}
}

// RecordFailure records a failure and returns true if circuit is open
func (cb *CircuitBreaker) RecordFailure() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.AddInt32(&cb.failures, 1)
	cb.lastFailure = time.Now()

	return atomic.LoadInt32(&cb.failures) >= cb.threshold
}

// RecordSuccess resets the circuit breaker
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.StoreInt32(&cb.failures, 0)
}

// IsOpen returns true if the circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if atomic.LoadInt32(&cb.failures) >= cb.threshold {
		if time.Since(cb.lastFailure) > cb.timeout {
			// Reset after timeout
			atomic.StoreInt32(&cb.failures, 0)
			return false
		}
		return true
	}
	return false
}

// RetryWithBackoff attempts to execute a function with exponential backoff
func RetryWithBackoff(operation func() error, config *RetryConfig) error {
	var lastErr error
	interval := config.InitialInterval
	startTime := time.Now()

	for i := 0; i < config.MaxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			RecordRetryAttempt()

			// Check if we've exceeded max elapsed time
			if time.Since(startTime) > config.MaxElapsedTime {
				return fmt.Errorf("max elapsed time exceeded: %w", lastErr)
			}

			// Calculate next interval with exponential backoff
			interval = time.Duration(float64(interval) * config.Multiplier)
			if interval > config.MaxInterval {
				interval = config.MaxInterval
			}

			log.Warn("Database operation failed, retrying...",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", config.MaxRetries),
				zap.Duration("interval", interval),
				zap.Error(err),
			)

			time.Sleep(interval)
			continue
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// InitDB initializes the database connection with proper connection pooling and retry logic
func InitDB(config *config.Config) error {
	var err error
	dsn := config.GetDSN()

	// Create circuit breaker
	circuitBreaker := NewCircuitBreaker(5, 30*time.Second)

	// Initialize database with retry logic
	err = RetryWithBackoff(func() error {
		if circuitBreaker.IsOpen() {
			UpdateCircuitBreakerMetrics(true, atomic.LoadInt32(&circuitBreaker.failures))
			return fmt.Errorf("circuit breaker is open")
		}

		dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			circuitBreaker.RecordFailure()
			UpdateCircuitBreakerMetrics(false, atomic.LoadInt32(&circuitBreaker.failures))
			RecordConnectionError()
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		sqlDB, err := dbInstance.DB()
		if err != nil {
			circuitBreaker.RecordFailure()
			UpdateCircuitBreakerMetrics(false, atomic.LoadInt32(&circuitBreaker.failures))
			RecordConnectionError()
			return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
		}

		// Test the connection
		if err := sqlDB.Ping(); err != nil {
			circuitBreaker.RecordFailure()
			UpdateCircuitBreakerMetrics(false, atomic.LoadInt32(&circuitBreaker.failures))
			RecordConnectionError()
			return fmt.Errorf("failed to ping database: %w", err)
		}

		circuitBreaker.RecordSuccess()
		UpdateCircuitBreakerMetrics(false, 0)

		// Set connection pool parameters
		sqlDB.SetMaxIdleConns(25)                  // Maximum number of idle connections
		sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
		sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum amount of time a connection may be reused
		sqlDB.SetConnMaxIdleTime(time.Minute * 30) // Maximum amount of time a connection may be idle

		// Start connection pool monitoring
		go MonitorPool(context.Background(), time.Minute)

		// Start metrics collection
		go StartMetricsCollection(sqlDB, time.Minute)

		// Start backup scheduler if using PostgreSQL
		if config.Database.Driver == "postgres" {
			backupConfig := &BackupConfig{
				BackupDir:          "/var/backups/db",  // Default backup directory
				RetentionPeriod:    7 * 24 * time.Hour, // 7 days
				BackupInterval:     24 * time.Hour,     // 1 day
				CompressionEnabled: true,
			}
			go StartBackupScheduler(context.Background(), backupConfig)
		}

		return nil
	}, DefaultRetryConfig())

	if err != nil {
		return fmt.Errorf("failed to initialize database after retries: %w", err)
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return dbInstance
}

// CloseDB closes the database connection
func CloseDB() error {
	if dbInstance != nil {
		sqlDB, err := dbInstance.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// WithTransaction executes a function within a database transaction
func WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return dbInstance.WithContext(ctx).Transaction(fn)
}

// HealthCheck checks if the database is healthy
func HealthCheck(ctx context.Context) error {
	if dbInstance == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}
