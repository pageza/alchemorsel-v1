package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserServiceInterface defines the methods for user-related business logic.
type UserServiceInterface interface {
	Authenticate(ctx context.Context, email string, password string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
	PatchUser(ctx context.Context, id string, updates map[string]interface{}) error
	GetAllUsers(ctx context.Context) ([]*models.User, error)
}

// UserService is the implementation of UserServiceInterface.
type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Helper methods
func (s *UserService) validateUser(user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	if user.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(user.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

// Service methods
func (s *UserService) Authenticate(ctx context.Context, email string, password string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	if err := s.validateUser(user); err != nil {
		return err
	}

	// Check if user exists
	existingUser, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Assign a UUID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, user *models.User) error {
	if err := s.validateUser(user); err != nil {
		return err
	}
	return s.repo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

// ForgotPassword initiates the password reset process
func (s *UserService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		// Don't reveal if email exists
		return nil
	}

	// Generate reset token
	token := generateResetToken()
	user.ResetPasswordToken = token
	expiry := time.Now().Add(24 * time.Hour)
	user.ResetPasswordExpires = &expiry

	return s.repo.UpdateUser(ctx, user)
}

// ResetPassword completes the password reset process
func (s *UserService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	user, err := s.repo.GetUserByResetPasswordToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	if time.Now().After(*user.ResetPasswordExpires) {
		return fmt.Errorf("reset token has expired")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.ResetPasswordToken = ""
	user.ResetPasswordExpires = nil

	return s.repo.UpdateUser(ctx, user)
}

// PatchUser updates specific fields of a user
func (s *UserService) PatchUser(ctx context.Context, id string, updates map[string]interface{}) error {
	zap.S().Debugw("PatchUser: received patch payload", "id", id, "updates", updates)

	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	zap.S().Debugw("PatchUser: original user retrieved", "user", user)

	// Update fields
	for field, value := range updates {
		zap.S().Debugw("PatchUser: updating field", "field", field, "value", value)
		switch field {
		case "email":
			if email, ok := value.(string); ok {
				user.Email = email
				zap.S().Debugw("PatchUser: updated email", "email", email)
			}
		case "name":
			if name, ok := value.(string); ok {
				user.Name = name
				zap.S().Debugw("PatchUser: updated name", "name", name)
			}
		case "password":
			if password, ok := value.(string); ok {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					zap.S().Errorw("PatchUser: error hashing password", "error", err)
					return err
				}
				user.Password = string(hashedPassword)
				zap.S().Debug("PatchUser: updated password")
			}
		default:
			zap.S().Warnw("PatchUser: unrecognized field, skipping update", "field", field)
		}
	}
	zap.S().Debugw("PatchUser: updated user fields", "user", user)
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		zap.S().Errorw("PatchUser: repository failed to update user", "userID", user.ID, "error", err)
		return err
	}
	zap.S().Debugw("PatchUser: repository updated user successfully", "userID", user.ID)
	return nil
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAllUsers(ctx)
}

// Helper function to generate reset token
func generateResetToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
