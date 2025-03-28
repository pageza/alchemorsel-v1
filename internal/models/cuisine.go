package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cuisine represents a cuisine type (e.g., Italian, Chinese).
type Cuisine struct {
	ID   string `json:"id" gorm:"type:uuid;primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}

// BeforeCreate hook to set a UUID before creating a Cuisine record if ID is not set
func (c *Cuisine) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
