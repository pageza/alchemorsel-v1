package unit

import (
	"context"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
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

	// Initialize the repositories DB so repository functions use the same connection.
	if err := repositories.InitializeDB("file::memory:?cache=shared"); err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	user := &models.User{
		Name:  "Test User",
		Email: "test@example.com",
		// Updated password: at least 8 characters with one digit, one uppercase, one lowercase, and one special character.
		Password: "Test1234!",
	}
	err := services.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("User creation failed: %v", err)
	}
	// Verify that the plain text password is replaced with a hashed one.
	if user.Password == "Test1234!" {
		t.Errorf("Expected password to be hashed, but it remains in plain text")
	}
}
