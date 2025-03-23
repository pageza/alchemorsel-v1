package dtos

import (
	"encoding/json"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
)

// RecipeResponse defines the payload structure for returning a recipe.
type RecipeResponse struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Ingredients       []string `json:"ingredients"`
	Steps             []string `json:"steps"`
	NutritionalInfo   string   `json:"nutritional_info,omitempty"`
	AllergyDisclaimer string   `json:"allergy_disclaimer,omitempty"`
	// Many-to-many relationships converted to slice of names.
	Cuisines   []string `json:"cuisines,omitempty"`
	Diets      []string `json:"diets,omitempty"`
	Appliances []string `json:"appliances,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	// Additional fields for future enhancements.
	Images        []string  `json:"images,omitempty"`
	AverageRating float64   `json:"average_rating"`
	RatingCount   int       `json:"rating_count"`
	Difficulty    string    `json:"difficulty,omitempty"`
	PrepTime      int       `json:"prep_time"`
	CookTime      int       `json:"cook_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewRecipeResponse converts a models.Recipe into a RecipeResponse DTO.
// It unmarshals JSON fields and maps related models into slices of names.
func NewRecipeResponse(r *models.Recipe) RecipeResponse {
	var ingredients []string
	_ = json.Unmarshal(r.Ingredients, &ingredients)

	var steps []string
	_ = json.Unmarshal(r.Steps, &steps)

	var images []string
	_ = json.Unmarshal(r.Images, &images)

	var cuisines []string
	for _, c := range r.Cuisines {
		cuisines = append(cuisines, c.Name)
	}

	var diets []string
	for _, d := range r.Diets {
		diets = append(diets, d.Name)
	}

	var appliances []string
	for _, a := range r.Appliances {
		appliances = append(appliances, a.Name)
	}

	var tags []string
	for _, t := range r.Tags {
		tags = append(tags, t.Name)
	}

	return RecipeResponse{
		ID:                r.ID,
		Title:             r.Title,
		Ingredients:       ingredients,
		Steps:             steps,
		NutritionalInfo:   r.NutritionalInfo,
		AllergyDisclaimer: r.AllergyDisclaimer,
		Cuisines:          cuisines,
		Diets:             diets,
		Appliances:        appliances,
		Tags:              tags,
		Images:            images,
		AverageRating:     r.AverageRating,
		RatingCount:       r.RatingCount,
		Difficulty:        r.Difficulty,
		PrepTime:          r.PrepTime,
		CookTime:          r.CookTime,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}
