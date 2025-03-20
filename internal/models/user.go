package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user of the application.
type User struct {
	ID                       string         `json:"id"`
	Name                     string         `json:"name" binding:"required"`
	Email                    string         `json:"email" binding:"required,email" gorm:"uniqueIndex"`
	Password                 string         `json:"password" binding:"required"`
	IsAdmin                  bool           `json:"is_admin" gorm:"default:false"`
	EmailVerified            bool           `json:"email_verified" gorm:"default:false"`
	EmailVerificationToken   string         `json:"email_verification_token" gorm:"index"`
	EmailVerificationExpires *time.Time     `json:"email_verification_expires"`
	ResetPasswordToken       string         `json:"reset_password_token" gorm:"index"`
	ResetPasswordExpires     *time.Time     `json:"reset_password_expires"`
	LastLoginAt              *time.Time     `json:"last_login_at"`
	LastActiveAt             *time.Time     `json:"last_active_at"`
	DeletedAt                gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
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
