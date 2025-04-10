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

// CreateUser inserts a new user into the database.
func CreateUser(ctx context.Context, user *models.User) error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Always generate a new ID to prevent external IDs from causing inconsistencies.
	user.ID = uuid.NewString()
	user.Email = normalizeEmail(user.Email)

	// Insert the user record.
	if err := DB.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID from the database.
// Returns nil if the user is not found.
func GetUser(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := DB.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser modifies an existing user in the database.
func UpdateUser(ctx context.Context, id string, user *models.User) error {
	return DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(user).Error
}

// DeleteUser removes a user by ID from the database. (Not used for soft deletion)
func DeleteUser(ctx context.Context, id string) error {
	return DB.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

// DeactivateUser performs a soft deletion by marking the user as inactive.
func DeactivateUser(ctx context.Context, id string) error {
	// Soft delete the user by marking them as inactive.
	return DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// GetUserByEmail retrieves a user by email from the database.
func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var user models.User
	normalizedEmail := normalizeEmail(email)
	err := DB.WithContext(ctx).Where("email = ?", normalizedEmail).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetAllUsers retrieves all users from the database.
func GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := DB.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserByEmailVerificationToken retrieves a user with a matching email verification token.
func GetUserByEmailVerificationToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	if err := DB.WithContext(ctx).Where("email_verification_token = ?", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByResetPasswordToken retrieves a user by the reset token.
func GetUserByResetPasswordToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	if err := DB.WithContext(ctx).Where("reset_password_token = ?", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
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
	// TRACE: entering the repository update function
	zap.S().Debug("TRACE: Entering UpdateUser repository function.")
	zap.S().Debugw("DefaultUserRepository: Attempting to update user", "user", user)
	// TRACE: calling the DB Save method
	zap.S().Debug("TRACE: Calling database Save method.")
	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		zap.S().Errorw("DefaultUserRepository: Failed to update user", "userID", user.ID, "error", err)
	} else {
		zap.S().Debugw("DefaultUserRepository: Successfully updated user", "userID", user.ID)
	}
	// TRACE: exiting the repository update function
	zap.S().Debug("TRACE: Exiting UpdateUser repository function.")
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
