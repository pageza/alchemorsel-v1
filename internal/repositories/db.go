package repositories

import (
	"context"
	"fmt"
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

// InitDB initializes the database connection with proper connection pooling
func InitDB(config *config.Config) error {
	var err error
	dsn := config.GetDSN()

	dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(25)                  // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum amount of time a connection may be reused
	sqlDB.SetConnMaxIdleTime(time.Minute * 30) // Maximum amount of time a connection may be idle

	// Start connection pool monitoring
	go MonitorPool(context.Background(), time.Minute)

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
