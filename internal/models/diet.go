package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Diet represents a diet category (e.g., vegan, keto).
type Diet struct {
	ID   string `json:"id" gorm:"type:uuid;primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}

// BeforeCreate hook to set a UUID before creating a Diet record if ID is not set
func (d *Diet) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
