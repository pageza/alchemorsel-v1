package repositories

import (
	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the database connection instance
var DB *gorm.DB

// InitializeDB initializes the database connection
func InitializeDB(dsn string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	return err
}

// AutoMigrate runs database migrations for our models.
func AutoMigrate() error {
	return DB.AutoMigrate(&models.User{})
}

// ClearUsers deletes all records from the users table.
func ClearUsers() error {
	return DB.Exec("DELETE FROM users").Error
}
