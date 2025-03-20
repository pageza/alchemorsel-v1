package repositories

import (
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the database connection instance
var DB *gorm.DB

// InitializeDB initializes the database connection.
// For in-memory SQLite, forcing only one open connection ensures the DB persists.
func InitializeDB(dsn string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	// Limit to one connection so that the in‑memory database is shared correctly.
	sqlDB.SetMaxOpenConns(1)
	// Ensure the single connection remains available.
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return nil
}

// AutoMigrate runs database migrations for our models.
func AutoMigrate() error {
	return DB.AutoMigrate(&models.User{})
}

// ClearUsers deletes all records from the users table.
func ClearUsers() error {
	return DB.Exec("DELETE FROM users").Error
}
