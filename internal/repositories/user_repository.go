package repositories

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/gorm"
)

// normalizeEmail converts the email to lowercase and removes any alias (e.g. text after '+').
func normalizeEmail(email string) string {
	email = strings.ToLower(email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	local := parts[0]
	domain := parts[1]
	if idx := strings.Index(local, "+"); idx != -1 {
		local = local[:idx]
	}
	return local + "@" + domain
}



type UserRepository interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	GetUserByResetPasswordToken(ctx context.Context, token string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	FindByEmail(email string) (*models.User, error)
}

type DefaultUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &DefaultUserRepository{db: db}
}

func (r *DefaultUserRepository) GetUser(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *DefaultUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normalizedEmail := normalizeEmail(email)
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", normalizedEmail).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *DefaultUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Always generate a new ID to prevent external IDs from causing inconsistencies.
	user.ID = uuid.NewString()
	user.Email = normalizeEmail(user.Email)

	return r.db.WithContext(ctx).Create(user).Error
}

func (r *DefaultUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	zap.S().Debugw("DefaultUserRepository: Attempting to update user", "user", user)
	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		zap.S().Errorw("DefaultUserRepository: Failed to update user", "userID", user.ID, "error", err)
	} else {
		zap.S().Debugw("DefaultUserRepository: Successfully updated user", "userID", user.ID)
	}
	return err
}

func (r *DefaultUserRepository) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

func (r *DefaultUserRepository) GetUserByResetPasswordToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("reset_password_token = ?", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *DefaultUserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *DefaultUserRepository) FindByEmail(email string) (*models.User, error) {
	// Validate email format
	if !isValidEmail(email) {
		return nil, fmt.Errorf("invalid email format")
	}

	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
