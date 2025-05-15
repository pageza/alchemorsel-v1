package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// EmbeddingVector is a custom type for []float32 that implements sql.Scanner and driver.Valuer
type EmbeddingVector []float32

// Value implements the driver.Valuer interface for EmbeddingVector
func (e EmbeddingVector) Value() (driver.Value, error) {
	if len(e) == 0 {
		return nil, nil
	}

	// Convert []float32 to string in PostgreSQL vector format: "[1,2,3]"
	strValues := make([]string, len(e))
	for i, v := range e {
		strValues[i] = fmt.Sprintf("%f", v)
	}
	return fmt.Sprintf("[%s]", strings.Join(strValues, ",")), nil
}

// Scan implements the sql.Scanner interface for EmbeddingVector
func (e *EmbeddingVector) Scan(value interface{}) error {
	if value == nil {
		*e = nil
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return errors.New("type assertion to string failed")
	}

	// Remove brackets and split by comma
	str = strings.Trim(str, "[]")
	if str == "" {
		*e = make([]float32, 0)
		return nil
	}

	values := strings.Split(str, ",")
	result := make([]float32, len(values))
	for i, v := range values {
		var f float32
		_, err := fmt.Sscanf(strings.TrimSpace(v), "%f", &f)
		if err != nil {
			return fmt.Errorf("failed to parse float: %v", err)
		}
		result[i] = f
	}

	*e = result
	return nil
}

type Recipe struct {
	ID               string          `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Title            string          `gorm:"not null" json:"title"`
	Description      string          `gorm:"type:text" json:"description"`
	Servings         int             `json:"servings"`
	PrepTimeMinutes  int             `json:"prep_time_minutes"`
	CookTimeMinutes  int             `json:"cook_time_minutes"`
	TotalTimeMinutes int             `json:"total_time_minutes"`
	Ingredients      Ingredients     `gorm:"type:jsonb" json:"ingredients"`
	Instructions     Instructions    `gorm:"type:jsonb" json:"instructions"`
	Nutrition        Nutrition       `gorm:"type:jsonb" json:"nutrition"`
	Tags             pq.StringArray  `gorm:"type:text[]" json:"tags"`
	Difficulty       string          `json:"difficulty"`
	Embedding        EmbeddingVector `gorm:"type:vector(1536)" json:"-"` // Ada-2 embeddings are 1536 dimensions
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	UserID           string          `gorm:"type:uuid;not null" json:"user_id"`
}

// Value implements the driver.Valuer interface for Recipe
func (r Recipe) Value() (driver.Value, error) {
	// Create a map to store all fields except embedding
	recipeMap := map[string]interface{}{
		"id":                 r.ID,
		"title":              r.Title,
		"description":        r.Description,
		"servings":           r.Servings,
		"prep_time_minutes":  r.PrepTimeMinutes,
		"cook_time_minutes":  r.CookTimeMinutes,
		"total_time_minutes": r.TotalTimeMinutes,
		"ingredients":        r.Ingredients,
		"instructions":       r.Instructions,
		"nutrition":          r.Nutrition,
		"tags":               r.Tags,
		"difficulty":         r.Difficulty,
		"created_at":         r.CreatedAt,
		"updated_at":         r.UpdatedAt,
		"user_id":            r.UserID,
	}

	// Marshal the map to JSON
	jsonData, err := json.Marshal(recipeMap)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// Scan implements the sql.Scanner interface for Recipe
func (r *Recipe) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	// Create a temporary struct to unmarshal the JSON
	var temp struct {
		ID               string         `json:"id"`
		Title            string         `json:"title"`
		Description      string         `json:"description"`
		Servings         int            `json:"servings"`
		PrepTimeMinutes  int            `json:"prep_time_minutes"`
		CookTimeMinutes  int            `json:"cook_time_minutes"`
		TotalTimeMinutes int            `json:"total_time_minutes"`
		Ingredients      Ingredients    `json:"ingredients"`
		Instructions     Instructions   `json:"instructions"`
		Nutrition        Nutrition      `json:"nutrition"`
		Tags             pq.StringArray `json:"tags"`
		Difficulty       string         `json:"difficulty"`
		CreatedAt        time.Time      `json:"created_at"`
		UpdatedAt        time.Time      `json:"updated_at"`
		UserID           string         `json:"user_id"`
	}

	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}

	// Copy values to the recipe
	r.ID = temp.ID
	r.Title = temp.Title
	r.Description = temp.Description
	r.Servings = temp.Servings
	r.PrepTimeMinutes = temp.PrepTimeMinutes
	r.CookTimeMinutes = temp.CookTimeMinutes
	r.TotalTimeMinutes = temp.TotalTimeMinutes
	r.Ingredients = temp.Ingredients
	r.Instructions = temp.Instructions
	r.Nutrition = temp.Nutrition
	r.Tags = temp.Tags
	r.Difficulty = temp.Difficulty
	r.CreatedAt = temp.CreatedAt
	r.UpdatedAt = temp.UpdatedAt
	r.UserID = temp.UserID

	return nil
}

type Ingredient struct {
	Item   string      `json:"item"`
	Amount json.Number `json:"amount"`
	Unit   string      `json:"unit"`
}

// Ingredients is a custom type for []Ingredient that implements sql.Scanner and driver.Valuer
type Ingredients []Ingredient

// Value implements the driver.Valuer interface for Ingredients
func (i Ingredients) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// Scan implements the sql.Scanner interface for Ingredients
func (i *Ingredients) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, i)
}

type Instruction struct {
	Step        int    `json:"step"`
	Description string `json:"description"`
}

// Instructions is a custom type for []Instruction that implements sql.Scanner and driver.Valuer
type Instructions []Instruction

// Value implements the driver.Valuer interface for Instructions
func (i Instructions) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// Scan implements the sql.Scanner interface for Instructions
func (i *Instructions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, i)
}

type Nutrition struct {
	Calories int    `json:"calories"`
	Protein  string `json:"protein"`
	Carbs    string `json:"carbs"`
	Fat      string `json:"fat"`
}

// Value implements the driver.Valuer interface for Nutrition
func (n Nutrition) Value() (driver.Value, error) {
	return json.Marshal(n)
}

// Scan implements the sql.Scanner interface for Nutrition
func (n *Nutrition) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, n)
}
