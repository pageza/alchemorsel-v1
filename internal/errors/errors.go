package errors

import "errors"

// Common custom error variables.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDeactivated = errors.New("account is deactivated")
)
