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

// Recipe represents a recipe in the application.
type Recipe struct {
	ID                string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Title             string         `json:"title"`
	Ingredients       datatypes.JSON `json:"ingredients" gorm:"type:json"` // Stored as a JSON array
	Steps             datatypes.JSON `json:"steps" gorm:"type:json"`       // Stored as a JSON array
	NutritionalInfo   string         `json:"nutritional_info"`             // TODO: Consider using a separate struct or table in future.
	AllergyDisclaimer string         `json:"allergy_disclaimer"`           // TODO: Consider using a separate struct or table in future.
	Embedding         Float64Slice   `json:"embedding" gorm:"type:json"`   // custom type for embedding slice
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`

	// New many-to-many relationships for cuisines and diets
	Cuisines   []Cuisine   `json:"cuisines" gorm:"many2many:recipe_cuisines;"`
	Diets      []Diet      `json:"diets" gorm:"many2many:recipe_diets;"`
	Appliances []Appliance `json:"appliances" gorm:"many2many:recipe_appliances;"`

	// New many-to-many relationship for extensible tags/flags
	Tags []Tag `json:"tags" gorm:"many2many:recipe_tags;"`

	// New fields for future enhancements
	Images        datatypes.JSON `json:"images" gorm:"type:json"` // JSON array of image URLs
	AverageRating float64        `json:"average_rating" gorm:"default:0"`
	RatingCount   int            `json:"rating_count" gorm:"default:0"`
	Difficulty    string         `json:"difficulty"` // e.g., "Easy", "Medium", "Hard"
	PrepTime      int            `json:"prep_time"`  // in minutes
	CookTime      int            `json:"cook_time"`  // in minutes
	Approved      bool           `json:"approved" gorm:"default:false"`
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
func (r *Recipe) GetIngredients() ([]string, error) {
	var ingredients []string
	if err := json.Unmarshal([]byte(r.Ingredients), &ingredients); err != nil {
		return nil, err
	}
	return ingredients, nil
}

func (r *Recipe) SetIngredients(ingredients []string) error {
	data, err := json.Marshal(ingredients)
	if err != nil {
		return err
	}
	r.Ingredients = datatypes.JSON(data)
	return nil
}

func (r *Recipe) GetSteps() ([]string, error) {
	var steps []string
	if err := json.Unmarshal([]byte(r.Steps), &steps); err != nil {
		return nil, err
	}
	return steps, nil
}

func (r *Recipe) SetSteps(steps []string) error {
	data, err := json.Marshal(steps)
	if err != nil {
		return err
	}
	r.Steps = datatypes.JSON(data)
	return nil
}
