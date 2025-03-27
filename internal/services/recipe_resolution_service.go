package services

import (
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
)

// ResolveRecipe searches for a matching recipe; if not found, generates one using external APIs.
func ResolveRecipe(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	// TODO: 1. Search database for matching recipes using pgvector.
	// TODO: 2. If no close match, call external Deepseek API to generate a recipe.
	// TODO: 3. Call OpenAI API to obtain an embedding.
	// TODO: 4. Insert generated recipe into the database.

	// For now, assume no close match found.

	// TODO: 2. Call external Deepseek API to generate a recipe.
	// Simulate a generated recipe.
	generatedRecipe := &models.Recipe{
		ID:          "generated-id",
		Title:       "Generated Recipe for " + query,
		Ingredients: []byte("[\"ingredient1\", \"ingredient2\"]"),
		Steps:       []byte("[\"step1\", \"step2\"]"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// TODO: 3. Call OpenAI API to obtain an embedding.
	// For demonstration, assign a dummy embedding as a slice of float64.
	generatedRecipe.Embedding = models.Float64Slice{0.1, 0.2, 0.3}

	// TODO: 4. Insert generated recipe into the database (skipped for now).

	// For similar recipes, return an empty list.
	similarRecipes := []*models.Recipe{}

	return generatedRecipe, similarRecipes, nil
}
