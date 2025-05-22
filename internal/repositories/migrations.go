package repositories

import (
	"fmt"
	"os"

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

// ClearUsers removes all data from the users table.
// This helper is used for test and integration setup.
func ClearUsers() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return DB.Exec("DELETE FROM users").Error
}

func RunMigrations(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create required PostgreSQL extensions using GORM
	if os.Getenv("DB_DRIVER") == "postgres" {
		// Create extensions using GORM's Exec
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" WITH SCHEMA public").Error; err != nil {
			return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
		}
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
			return fmt.Errorf("failed to create vector extension: %w", err)
		}
	}

	// Use the Recipe model from models package as the single source of truth
	if err := db.AutoMigrate(&models.Recipe{}); err != nil {
		return fmt.Errorf("failed to migrate recipes table: %w", err)
	}

	// Create vector similarity search index using GORM
	if os.Getenv("DB_DRIVER") == "postgres" {
		// Create the index using GORM's Exec
		if err := db.Exec("CREATE INDEX IF NOT EXISTS recipe_embedding_idx ON recipes USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)").Error; err != nil {
			return fmt.Errorf("failed to create vector similarity index: %w", err)
		}
	}

	return nil
}
