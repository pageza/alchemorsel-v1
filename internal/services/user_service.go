package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"

	"github.com/golang-jwt/jwt/v4"
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

// LoginUser authenticates a user and returns a JWT token.
func LoginUser(req *models.LoginRequest) (string, error) {
	// Retrieve user by email.
	user, err := repositories.GetUserByEmail(req.Email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Compare the provided password with the stored hashed password.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token.
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", errors.New("JWT secret is not set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
