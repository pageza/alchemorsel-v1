package repositories

import (
	"os"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the database connection instance
var DB *gorm.DB

// InitializeDB initializes the database connection.
// For in-memory SQLite, forcing only one open connection ensures the DB persists.
func InitializeDB(dsn string) error {
	// If using the in-memory DSN, select an appropriate DSN.
	if dsn == "file::memory:?cache=shared" {
		if os.Getenv("INTEGRATION_TEST") == "true" {
			// Use a file-based DB for integration tests to ensure persistence across connections.
			dsn = "./test.db"
		} else {
			// Use shared in-memory if not in integration test mode.
			dsn = "file:memdb1?mode=memory&cache=shared"
		}
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
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
	return DB.AutoMigrate(&models.User{})
}
