package services

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"

	stdErrors "errors"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	appErrors "github.com/pageza/alchemorsel-v1/internal/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user.
func CreateUser(user *models.User) error {
	// Always override any provided ID with a new UUID.
	user.ID = uuid.NewString()

	if user.Password == "" {
		return stdErrors.New("password required")
	}

	// Sanitize input: trim spaces and lowercase the email.
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Name = strings.TrimSpace(user.Name)

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
	user, err := repositories.GetUser(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user with id %s not found", id)
	}
	// Omit password for security.
	user.Password = ""
	return user, nil
}

// UpdateUser updates an existing user.
func UpdateUser(id string, user *models.User) error {
	// Retrieve the existing user.
	existingUser, err := repositories.GetUser(id)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return fmt.Errorf("user not found")
	}

	// No need to check for deactivation explicitly; soft-deleted records are excluded.

	// Update allowed fields if provided.
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}

	return repositories.UpdateUser(id, existingUser)
}

// DeleteUser deletes a user by ID.
func DeleteUser(id string) error {
	// TODO: Implement logic to delete a user.
	return stdErrors.New("not implemented")
}

// LoginUser authenticates a user and returns a JWT token.
func LoginUser(req *models.LoginRequest) (string, error) {
	// Retrieve user by email.
	user, err := repositories.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		return "", stdErrors.New("invalid credentials")
	}

	// With soft deletion in place, deleted users are not returned by default.

	// Compare the provided password with the stored hashed password.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", stdErrors.New("invalid credentials")
	}

	// Generate JWT token using "id" as claim key to match the middleware.
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", stdErrors.New("JWT secret is not set")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GetAllUsers returns a slice of all users.
func GetAllUsers() ([]*models.User, error) {
	users, err := repositories.GetAllUsers()
	if err != nil {
		return nil, err
	}
	// Optionally, you can sanitize or filter user data here.
	for _, u := range users {
		u.Password = ""
	}
	return users, nil
}

// DeactivateUser marks a user as inactive (soft delete).
func DeactivateUser(id string) error {
	return repositories.DeactivateUser(id)
}

// PatchUser performs a partial update on a user allowing only permitted fields (e.g., name and email).
func PatchUser(id string, patchData map[string]interface{}) error {
	// Retrieve the existing user.
	user, err := GetUser(id)
	if err != nil {
		return err
	}
	if user == nil {
		return appErrors.ErrUserNotFound
	}

	// Allow updates only on specific fields.
	if name, ok := patchData["name"]; ok {
		if str, ok := name.(string); ok {
			user.Name = strings.TrimSpace(str)
		}
	}
	if email, ok := patchData["email"]; ok {
		if str, ok := email.(string); ok {
			user.Email = strings.ToLower(strings.TrimSpace(str))
		}
	}

	// Reuse the existing UpdateUser function to persist changes.
	return UpdateUser(id, user)
}
