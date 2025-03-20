package models

import (
	"gorm.io/gorm"
)

// User represents a user of the application.
type User struct {
	ID        string         `json:"id"`
	Name      string         `json:"name" binding:"required"`
	Email     string         `json:"email" binding:"required,email"`
	Password  string         `json:"password" binding:"required"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// LoginRequest represents the JSON payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response containing the JWT token.
type LoginResponse struct {
	Token string `json:"token"`
}
