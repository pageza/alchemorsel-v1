package unit

import (
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
)

// TestMain sets up an in-memory SQLite database for unit tests.
func TestMain(m *testing.M) {
	var err error
	// Use SQLite in-memory for unit tests instead of PostgreSQL.
	db.DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		os.Exit(1)
	}

	// AutoMigrate creates the required schema for the User model.
	if err := db.DB.AutoMigrate(&models.User{}); err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestCreateUser(t *testing.T) {
	user := &models.User{
		ID:       "1",
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password", // Plain text password to be hashed.
	}
	err := services.CreateUser(user)
	if err != nil {
		t.Errorf("User creation failed: %v", err)
	}
	// Verify that the password was hashed and does not equal the plain text.
	if user.Password == "password" {
		t.Error("Expected password to be hashed, but it remains in plain text")
	}
}
