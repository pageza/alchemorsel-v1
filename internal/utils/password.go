package utils

import (
	"errors"
	"unicode"
)

// ValidatePassword checks that the password has at least 8 characters,
// includes one uppercase letter, one lowercase letter, one digit, and one special character.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	var hasNumber, hasUpper, hasLower, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	if !hasNumber {
		return errors.New("password must contain at least one digit")
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	return nil
}
