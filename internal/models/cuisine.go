package models

// Cuisine represents a cuisine type (e.g., Italian, Chinese).
type Cuisine struct {
	ID   string `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}
