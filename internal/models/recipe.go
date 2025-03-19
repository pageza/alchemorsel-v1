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
	Appliances        []string  `json:"appliances"`
	Embedding         []float64 `json:"embedding"` // TODO: Consider defining a custom type if needed.
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
