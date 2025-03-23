package models

import "time"

// Recipe represents a recipe in the application.
type Recipe struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Ingredients       []string  `json:"ingredients"`
	Steps             []string  `json:"steps"`
	NutritionalInfo   string    `json:"nutritional_info"`   // TODO: Consider using a separate struct or table in future.
	AllergyDisclaimer string    `json:"allergy_disclaimer"` // TODO: Consider using a separate struct or table in future.
	Embedding         []float64 `json:"embedding"`          // TODO: Consider defining a custom type if needed.
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// New many-to-many relationships for cuisines and diets
	Cuisines   []Cuisine   `json:"cuisines" gorm:"many2many:recipe_cuisines;"`
	Diets      []Diet      `json:"diets" gorm:"many2many:recipe_diets;"`
	Appliances []Appliance `json:"appliances" gorm:"many2many:recipe_appliances;"`

	// New many-to-many relationship for extensible tags/flags
	Tags []Tag `json:"tags" gorm:"many2many:recipe_tags;"`
}
