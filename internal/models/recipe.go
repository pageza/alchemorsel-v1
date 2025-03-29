package models

import (
	"time"

	"github.com/google/uuid"

	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Define a custom type for a slice of float64
type Float64Slice []float64

// Value implements the driver.Valuer interface for Float64Slice.
func (f Float64Slice) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan implements the sql.Scanner interface for Float64Slice.
func (f *Float64Slice) Scan(value interface{}) error {
	if value == nil {
		*f = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal Float64Slice value")
	}
	return json.Unmarshal(bytes, f)
}

// Ingredient represents a single ingredient in a recipe.
type Ingredient struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

// Step represents a single step in a recipe.
type Step struct {
	Order       int    `json:"order"`
	Description string `json:"description"`
}

// Recipe represents a recipe in the application.
type Recipe struct {
	ID                string         `json:"id" gorm:"primaryKey"`
	Title             string         `json:"title" gorm:"not null"`
	Description       string         `json:"description"`
	Ingredients       datatypes.JSON `json:"ingredients" gorm:"type:json"`
	Steps             datatypes.JSON `json:"steps" gorm:"type:json"`
	NutritionalInfo   string         `json:"nutritional_info"`
	AllergyDisclaimer string         `json:"allergy_disclaimer"`
	Cuisines          []Cuisine      `json:"cuisines" gorm:"many2many:recipe_cuisines;"`
	Diets             []Diet         `json:"diets" gorm:"many2many:recipe_diets;"`
	Appliances        []Appliance    `json:"appliances" gorm:"many2many:recipe_appliances;"`
	Tags              []Tag          `json:"tags" gorm:"many2many:recipe_tags;"`
	Images            datatypes.JSON `json:"images" gorm:"type:json"`
	Difficulty        string         `json:"difficulty"`
	PrepTime          int            `json:"prep_time"`
	CookTime          int            `json:"cooking_time"`
	Servings          int            `json:"servings"`
	AverageRating     float64        `json:"average_rating"`
	RatingCount       int            `json:"rating_count"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	Approved          bool           `json:"approved"`
	Embedding         Float64Slice   `json:"embedding" gorm:"type:json"`
}

// BeforeCreate is a GORM hook that runs before a new record is inserted.
// It ensures that a new UUID is generated if the ID is empty.
func (r *Recipe) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = time.Now()
	}
	return nil
}

// Helper methods for JSON conversion
func (r *Recipe) GetIngredients() ([]Ingredient, error) {
	var ingredients []Ingredient
	if err := json.Unmarshal([]byte(r.Ingredients), &ingredients); err != nil {
		return nil, err
	}
	return ingredients, nil
}

func (r *Recipe) SetIngredients(ingredients []Ingredient) error {
	data, err := json.Marshal(ingredients)
	if err != nil {
		return err
	}
	r.Ingredients = datatypes.JSON(data)
	return nil
}

func (r *Recipe) GetSteps() ([]Step, error) {
	var steps []Step
	if err := json.Unmarshal([]byte(r.Steps), &steps); err != nil {
		return nil, err
	}
	return steps, nil
}

func (r *Recipe) SetSteps(steps []Step) error {
	data, err := json.Marshal(steps)
	if err != nil {
		return err
	}
	r.Steps = datatypes.JSON(data)
	return nil
}
