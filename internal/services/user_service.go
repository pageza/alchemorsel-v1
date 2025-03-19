package services

import (
	"errors"
	"fmt"
	"recipeservice/internal/models"
	"recipeservice/internal/repositories"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user.
func CreateUser(user *models.User) error {
	if user.Password == "" {
		return errors.New("password required")
	}

	// Hash the user's password using bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("failed to hash password", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Call the repository layer to persist the user.
	return repositories.CreateUser(user)
}

// GetUser retrieves a user by ID.
func GetUser(id string) (*models.User, error) {
	// TODO: Implement logic to retrieve a user.
	return nil, errors.New("not implemented")
}

// UpdateUser updates an existing user.
func UpdateUser(id string, user *models.User) error {
	// TODO: Implement logic to update a user.
	return errors.New("not implemented")
}

// DeleteUser deletes a user by ID.
func DeleteUser(id string) error {
	// TODO: Implement logic to delete a user.
	return errors.New("not implemented")
}
