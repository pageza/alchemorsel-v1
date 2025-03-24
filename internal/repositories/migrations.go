package repositories

import (
	"context"
	"fmt"
	"os"

	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the database connection instance
var DB *gorm.DB

// InitializeDB initializes the database connection.
// For in-memory SQLite, forcing only one open connection ensures the DB persists.
func InitializeDB(dsn string) error {
	// If the repositories DB is already initialized, return immediately.
	if DB != nil {
		return nil
	}
	// If the global db connection is already initialized, use it.
	if db.DB != nil {
		DB = db.DB
		return nil
	}

	driver := os.Getenv("DB_DRIVER")
	var err error
	if driver == "" || driver == "sqlite" {
		// Adjust DSN for in‑memory SQLite if needed.
		if dsn == "file::memory:?cache=shared" {
			if os.Getenv("INTEGRATION_TEST") == "true" {
				dsn = "./test.db"
			} else {
				dsn = "file:memdb1?mode=memory&cache=shared"
			}
		}
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
	} else if driver == "postgres" {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported DB_DRIVER: %s", driver)
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	// Limit to one connection so that the in‑memory database (or file‑based DB, in tests) is shared correctly.
	sqlDB.SetMaxOpenConns(1)
	// Ensure the single connection remains available.
	sqlDB.SetMaxIdleConns(1)
	// Setting lifetime to 0 disables expiration of the connection.
	sqlDB.SetConnMaxLifetime(0)
	return nil
}

// AutoMigrate runs database migrations for our models.
func AutoMigrate() error {
	// Use the underlying sql.DB to execute the extension creation outside any transaction.
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	if _, err := sqlDB.ExecContext(context.Background(), "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" WITH SCHEMA public;"); err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	return DB.AutoMigrate(
		&models.User{},
		&models.Recipe{},
		&models.Tag{},
		&models.Appliance{},
	)
}

// ClearUsers removes all data from the users table.
// This helper is used for test and integration setup.
func ClearUsers() error {
	return DB.Exec("DELETE FROM users").Error
}
