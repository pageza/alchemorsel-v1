package migrations

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/gorm"
)

// MigrationVersion represents the current database schema version
type MigrationVersion struct {
	Version uint      `gorm:"primaryKey"`
	Dirty   bool      `gorm:"not null"`
	Applied time.Time `gorm:"not null"`
}

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	// First, ensure the migrations table exists
	if err := db.AutoMigrate(&MigrationVersion{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Create a new migrate instance
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"))
	m, err := migrate.New("file:///app/migrations", dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// If dirty, force the version to be clean
	if dirty {
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("failed to force clean version: %w", err)
		}
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Update the migrations table
	newVersion, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get new version: %w", err)
	}

	migrationVersion := MigrationVersion{
		Version: uint(newVersion),
		Dirty:   false,
		Applied: time.Now(),
	}

	if err := db.Create(&migrationVersion).Error; err != nil {
		return fmt.Errorf("failed to record migration version: %w", err)
	}

	// Run GORM auto-migrations for any new models
	return db.AutoMigrate(
		&models.User{},
		&models.Recipe{},
	)
}

// RollbackMigrations rolls back the last migration
func RollbackMigrations(db *gorm.DB) error {
	m, err := migrate.New("file:///app/migrations", db.Dialector.Name())
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(db *gorm.DB) (uint, error) {
	var version MigrationVersion
	if err := db.Order("version desc").First(&version).Error; err != nil {
		return 0, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version.Version, nil
}
