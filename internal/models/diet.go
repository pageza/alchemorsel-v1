package models

// Diet represents a diet category (e.g., vegan, keto).
type Diet struct {
	ID   string `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}
