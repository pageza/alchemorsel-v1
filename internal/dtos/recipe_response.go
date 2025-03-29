package dtos

import (
	"encoding/json"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
)

// RecipeResponse defines the payload structure for returning a recipe.
type RecipeResponse struct {
	ID                string       `json:"id"`
	Title             string       `json:"title"`
	Description       string       `json:"description,omitempty"`
	Ingredients       []Ingredient `json:"ingredients"`
	Steps             []Step       `json:"steps"`
	NutritionalInfo   string       `json:"nutritional_info,omitempty"`
	AllergyDisclaimer string       `json:"allergy_disclaimer,omitempty"`
	// Many-to-many relationships converted to slice of names.
	Cuisines   []string `json:"cuisines,omitempty"`
	Diets      []string `json:"diets,omitempty"`
	Appliances []string `json:"appliances,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	// Additional fields for future enhancements.
	Images        []string  `json:"images,omitempty"`
	Difficulty    string    `json:"difficulty,omitempty"`
	PrepTime      int       `json:"prep_time,omitempty"`
	CookTime      int       `json:"cooking_time,omitempty"`
	Servings      int       `json:"servings,omitempty"`
	AverageRating float64   `json:"average_rating,omitempty"`
	RatingCount   int       `json:"rating_count,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Approved      bool      `json:"approved,omitempty"`
}

// RecipeListResponse wraps a list of recipes in a response object
type RecipeListResponse struct {
	Recipes []RecipeResponse `json:"recipes"`
}

// NewRecipeResponse converts a models.Recipe into a RecipeResponse DTO.
// It unmarshals JSON fields and maps related models into slices of names.
func NewRecipeResponse(recipe *models.Recipe) *RecipeResponse {
	response := &RecipeResponse{
		ID:                recipe.ID,
		Title:             recipe.Title,
		Description:       recipe.Description,
		NutritionalInfo:   recipe.NutritionalInfo,
		AllergyDisclaimer: recipe.AllergyDisclaimer,
		Difficulty:        recipe.Difficulty,
		PrepTime:          recipe.PrepTime,
		CookTime:          recipe.CookTime,
		Servings:          recipe.Servings,
		Approved:          recipe.Approved,
		CreatedAt:         recipe.CreatedAt,
		UpdatedAt:         recipe.UpdatedAt,
	}

	// Convert ingredients JSON to array
	var ingredients []Ingredient
	if err := json.Unmarshal(recipe.Ingredients, &ingredients); err == nil {
		response.Ingredients = ingredients
	}

	// Convert steps JSON to array
	var steps []Step
	if err := json.Unmarshal(recipe.Steps, &steps); err == nil {
		response.Steps = steps
	}

	// Map related models to slices of names
	response.Cuisines = make([]string, len(recipe.Cuisines))
	for i, cuisine := range recipe.Cuisines {
		response.Cuisines[i] = cuisine.Name
	}

	response.Diets = make([]string, len(recipe.Diets))
	for i, diet := range recipe.Diets {
		response.Diets[i] = diet.Name
	}

	response.Appliances = make([]string, len(recipe.Appliances))
	for i, appliance := range recipe.Appliances {
		response.Appliances[i] = appliance.Name
	}

	response.Tags = make([]string, len(recipe.Tags))
	for i, tag := range recipe.Tags {
		response.Tags[i] = tag.Name
	}

	return response
}
