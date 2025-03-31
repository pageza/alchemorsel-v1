package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user of the application.
type User struct {
	ID                       string         `json:"id,omitempty" gorm:"type:uuid;primaryKey"`
	Name                     string         `json:"name" binding:"required"`
	Email                    string         `json:"email" binding:"required,email" gorm:"uniqueIndex:users_email_key"`
	Password                 string         `json:"password" binding:"required" gorm:"column:password_hash"`
	IsAdmin                  bool           `json:"is_admin,omitempty" gorm:"default:false"`
	EmailVerified            bool           `json:"email_verified,omitempty" gorm:"default:false"`
	EmailVerificationToken   string         `json:"email_verification_token,omitempty" gorm:"index"`
	EmailVerificationExpires *time.Time     `json:"email_verification_expires,omitempty"`
	ResetPasswordToken       string         `json:"reset_password_token,omitempty" gorm:"index"`
	ResetPasswordExpires     *time.Time     `json:"reset_password_expires,omitempty"`
	LastLoginAt              *time.Time     `json:"last_login_at,omitempty"`
	LastActiveAt             *time.Time     `json:"last_active_at,omitempty"`
	DeletedAt                gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	CreatedAt                time.Time      `json:"created_at,omitempty"`
	UpdatedAt                time.Time      `json:"updated_at,omitempty"`
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
