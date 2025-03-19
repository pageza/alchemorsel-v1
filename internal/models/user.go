package models

// User represents a user of the application.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"` // TODO: Store a hashed password instead of plain text.
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
