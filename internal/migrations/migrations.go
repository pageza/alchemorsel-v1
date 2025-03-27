package migrations

import (
	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	// Add your models here to be migrated
	return db.AutoMigrate(
		&models.User{},
		&models.Recipe{},
	)
}
