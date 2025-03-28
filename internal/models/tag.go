package models

// Tag represents an extensible label (or flag) that can be attached to a Recipe.
// It can be used for statuses such as "featured", "quick", "seasonal", etc.
type Tag struct {
	ID   string `json:"id" gorm:"type:uuid;primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}
