package dtos

import (
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
)

// UserResponse defines the structure exposed to API clients.
type UserResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	IsAdmin       bool      `json:"is_admin"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewUserResponse converts a models.User to a UserResponse DTO.
func NewUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		IsAdmin:       user.IsAdmin,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
