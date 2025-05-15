package repositories

import (
	"fmt"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	// Drop existing tables using Migrator
	if err := db.Migrator().DropTable(&models.Recipe{}, &models.User{}); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Create tables using Migrator
	if err := db.Migrator().CreateTable(&models.User{}); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create recipes table with explicit column definitions
	if err := db.Exec(`
		CREATE TABLE recipes (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			servings INTEGER,
			prep_time_minutes INTEGER,
			cook_time_minutes INTEGER,
			total_time_minutes INTEGER,
			ingredients JSONB,
			instructions JSONB,
			nutrition JSONB,
			tags TEXT[],
			difficulty VARCHAR(50),
			embedding vector(1536),
			created_at TIMESTAMP WITH TIME ZONE,
			updated_at TIMESTAMP WITH TIME ZONE,
			user_id UUID NOT NULL,
			CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create recipes table: %w", err)
	}

	// Create vector similarity search index using GORM
	if os.Getenv("DB_DRIVER") == "postgres" {
		var indexExists bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'recipe_embedding_idx')").Scan(&indexExists).Error
		if err != nil {
			return fmt.Errorf("failed to check if index exists: %w", err)
		}

		if !indexExists {
			err = db.Exec("CREATE INDEX recipe_embedding_idx ON recipes USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)").Error
			if err != nil {
				return fmt.Errorf("failed to create vector similarity index: %w", err)
			}
		}
	}

	return nil
}
