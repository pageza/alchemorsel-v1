package repositories

import (
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/gorm"
)

// CreateUser inserts a new user into the database.
func CreateUser(user *models.User) error {
	return db.DB.Create(user).Error
}

// GetUser retrieves a user by ID from the database.
// Returns nil if the user is not found.
func GetUser(id string) (*models.User, error) {
	var user models.User
	if err := db.DB.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser modifies an existing user in the database.
func UpdateUser(id string, user *models.User) error {
	return db.DB.Model(&models.User{}).Where("id = ?", id).Updates(user).Error
}

// DeleteUser removes a user by ID from the database. (Not used for soft deletion)
func DeleteUser(id string) error {
	return db.DB.Delete(&models.User{}, "id = ?", id).Error
}

// DeactivateUser performs a soft deletion by marking the user as inactive.
func DeactivateUser(id string) error {
	// Delete uses GORM's soft delete if the model has a DeletedAt field.
	return db.DB.Delete(&models.User{}, "id = ?", id).Error
}

// GetUserByEmail retrieves a user by email from the database.
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers retrieves all users from the database.
func GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	if err := db.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
