package repositories

import (
	"context"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/gorm"
)

// CreateUser inserts a new user into the database.
func CreateUser(ctx context.Context, user *models.User) error {
	return db.DB.WithContext(ctx).Create(user).Error
}

// GetUser retrieves a user by ID from the database.
// Returns nil if the user is not found.
func GetUser(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := db.DB.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser modifies an existing user in the database.
func UpdateUser(ctx context.Context, id string, user *models.User) error {
	return db.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(user).Error
}

// DeleteUser removes a user by ID from the database. (Not used for soft deletion)
func DeleteUser(ctx context.Context, id string) error {
	return db.DB.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

// DeactivateUser performs a soft deletion by marking the user as inactive.
func DeactivateUser(ctx context.Context, id string) error {
	// Soft delete the user by marking them as inactive.
	return db.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// GetUserByEmail retrieves a user by email from the database.
func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := db.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers retrieves all users from the database.
func GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := db.DB.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserByEmailVerificationToken retrieves a user with a matching email verification token.
func GetUserByEmailVerificationToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	if err := db.DB.WithContext(ctx).Where("email_verification_token = ?", token).First(&user).Error; err != nil {
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
	if err := db.DB.WithContext(ctx).Where("reset_password_token = ?", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
