package unit

import (
	"testing"

	"recipeservice/internal/models"
	"recipeservice/internal/services"
)

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
