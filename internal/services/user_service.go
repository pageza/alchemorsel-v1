package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/utils"

	stdErrors "errors"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	appErrors "github.com/pageza/alchemorsel-v1/internal/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user.
func CreateUser(ctx context.Context, user *models.User) error {
	// Always override any provided ID with a new UUID.
	user.ID = uuid.NewString()

	if user.Password == "" {
		return stdErrors.New("password required")
	}

	// Sanitize input: trim spaces and lowercase the email.
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Name = strings.TrimSpace(user.Name)

	// Validate password strength.
	if err := utils.ValidatePassword(user.Password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Configurable bcrypt cost.
	cost := bcrypt.DefaultCost
	if costStr := os.Getenv("BCRYPT_COST"); costStr != "" {
		if parsedCost, err := strconv.Atoi(costStr); err == nil {
			cost = parsedCost
		}
	}

	// Hash the user's password using bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), cost)
	if err != nil {
		zap.L().Error("failed to hash password", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Generate email verification token using secure token generation.
	token, err := utils.GenerateSecureToken(16)
	if err != nil {
		zap.L().Error("failed to generate secure token", zap.Error(err))
		return fmt.Errorf("failed to generate secure token: %w", err)
	}
	user.EmailVerificationToken = token
	expiration := time.Now().Add(24 * time.Hour)
	user.EmailVerificationExpires = &expiration

	// Audit log for user creation.
	zap.L().Info("Creating new user", zap.String("email", user.Email))

	return repositories.CreateUser(ctx, user)
}

// GetUser retrieves a user by ID.
func GetUser(ctx context.Context, id string) (*models.User, error) {
	user, err := repositories.GetUser(ctx, id)
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
func UpdateUser(ctx context.Context, id string, user *models.User) error {
	// Retrieve the existing user.
	existingUser, err := repositories.GetUser(ctx, id)
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
		existingUser.Email = strings.ToLower(strings.TrimSpace(user.Email))
	}

	return repositories.UpdateUser(ctx, id, existingUser)
}

// DeleteUser deletes a user by ID.
func DeleteUser(ctx context.Context, id string) error {
	// Soft delete the user by marking them as inactive.
	return repositories.DeactivateUser(ctx, id)
}

// LoginUser authenticates a user and returns a JWT token.
func LoginUser(ctx context.Context, req *models.LoginRequest) (string, error) {
	user, err := repositories.GetUserByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return "", appErrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", appErrors.ErrInvalidCredentials
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := repositories.UpdateUser(ctx, user.ID, user); err != nil {
		zap.L().Error("failed to update last login timestamp", zap.Error(err))
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", stdErrors.New("JWT secret is not set")
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})
	tokenString, err := tokenClaims.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GetAllUsers returns a slice of all users.
func GetAllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := repositories.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	// Sanitize sensitive fields.
	for _, u := range users {
		u.Password = ""
	}
	return users, nil
}

// DeactivateUser marks a user as inactive (soft delete).
func DeactivateUser(ctx context.Context, id string) error {
	return repositories.DeactivateUser(ctx, id)
}

// PatchUser performs a partial update on a user allowing only permitted fields (e.g., name and email).
func PatchUser(ctx context.Context, id string, patchData map[string]interface{}) error {
	user, err := GetUser(ctx, id)
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
	return UpdateUser(ctx, id, user)
}

// GetUserByEmailVerificationToken retrieves a user by token and checks expiration.
func GetUserByEmailVerificationToken(ctx context.Context, token string) (*models.User, error) {
	user, err := repositories.GetUserByEmailVerificationToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	if user.EmailVerificationExpires != nil && user.EmailVerificationExpires.Before(time.Now()) {
		return nil, nil
	}
	return user, nil
}

// ForgotPassword generates a reset token for the user.
func ForgotPassword(ctx context.Context, email string) error {
	user, err := repositories.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return stdErrors.New("user not found")
	}
	// Generate reset token using secure token generation.
	resetToken, err := utils.GenerateSecureToken(16)
	if err != nil {
		zap.L().Error("failed to generate secure reset token", zap.Error(err))
		return fmt.Errorf("failed to generate secure reset token: %w", err)
	}
	expiry := time.Now().Add(1 * time.Hour)
	if expStr := os.Getenv("PASSWORD_RESET_TOKEN_EXPIRY"); expStr != "" {
		if d, err := time.ParseDuration(expStr); err == nil {
			expiry = time.Now().Add(d)
		}
	}
	user.ResetPasswordToken = resetToken
	user.ResetPasswordExpires = &expiry
	zap.L().Info("Password reset token generated", zap.String("userID", user.ID))
	return repositories.UpdateUser(ctx, user.ID, user)
}

// ResetPassword updates the user's password using the provided token.
func ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := repositories.GetUserByResetPasswordToken(ctx, token)
	if err != nil || user == nil {
		return stdErrors.New("invalid or expired reset token")
	}
	if user.ResetPasswordExpires == nil || user.ResetPasswordExpires.Before(time.Now()) {
		return stdErrors.New("reset token expired")
	}
	if err := utils.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}
	cost := bcrypt.DefaultCost
	if costStr := os.Getenv("BCRYPT_COST"); costStr != "" {
		if parsedCost, err := strconv.Atoi(costStr); err == nil {
			cost = parsedCost
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), cost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	user.Password = string(hashedPassword)
	user.ResetPasswordToken = ""
	user.ResetPasswordExpires = nil
	zap.L().Info("Password reset successfully", zap.String("userID", user.ID))
	return repositories.UpdateUser(ctx, user.ID, user)
}
