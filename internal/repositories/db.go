package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pageza/alchemorsel-v1/internal/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// dbInstance is the global database instance
var dbInstance *gorm.DB

// PoolStats represents database connection pool statistics
type PoolStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
}

// GetPoolStats returns the current connection pool statistics
func GetPoolStats() (*PoolStats, error) {
	if dbInstance == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	stats := &PoolStats{
		MaxOpenConnections: sqlDB.Stats().MaxOpenConnections,
		OpenConnections:    sqlDB.Stats().OpenConnections,
		InUse:              sqlDB.Stats().InUse,
		Idle:               sqlDB.Stats().Idle,
		WaitCount:          sqlDB.Stats().WaitCount,
		WaitDuration:       sqlDB.Stats().WaitDuration,
	}

	return stats, nil
}

// MonitorPool starts monitoring the connection pool and logs statistics
func MonitorPool(ctx context.Context, interval time.Duration) {
	log := logrus.WithField("component", "db_pool_monitor")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping pool monitoring")
			return
		case <-ticker.C:
			stats, err := GetPoolStats()
			if err != nil {
				log.WithError(err).Error("Failed to get pool stats")
				continue
			}

			log.WithFields(logrus.Fields{
				"max_open_connections": stats.MaxOpenConnections,
				"open_connections":     stats.OpenConnections,
				"in_use":               stats.InUse,
				"idle":                 stats.Idle,
				"wait_count":           stats.WaitCount,
				"wait_duration":        stats.WaitDuration,
			}).Info("Connection pool statistics")
		}
	}
}

// InitDB initializes the database connection with retry logic and connection pooling
func InitDB(cfg *config.Config) error {
	log := logrus.WithField("component", "database")

	// Create DSN
	dsn := cfg.GetDSN()

	// Configure GORM logger
	gormLogger := logger.New(
		logrus.New(),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		// Set connection pool settings
		PrepareStmt: true,
	}

	// Create exponential backoff
	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.MaxElapsedTime = 30 * time.Second
	backoffConfig.InitialInterval = 1 * time.Second

	// Retry database connection
	operation := func() error {
		db, err := gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			log.WithError(err).Error("Failed to connect to database")
			return err
		}

		// Test the connection
		sqlDB, err := db.DB()
		if err != nil {
			log.WithError(err).Error("Failed to get database instance")
			return err
		}

		// Set connection pool settings
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

		// Test the connection
		if err := sqlDB.Ping(); err != nil {
			log.WithError(err).Error("Failed to ping database")
			return err
		}

		dbInstance = db
		log.Info("Successfully connected to database")
		return nil
	}

	// Execute with backoff
	if err := backoff.Retry(operation, backoffConfig); err != nil {
		return fmt.Errorf("failed to connect to database after retries: %w", err)
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
