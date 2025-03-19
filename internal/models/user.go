package models

// User represents a user of the application.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"` // TODO: Store a hashed password instead of plain text.
}
