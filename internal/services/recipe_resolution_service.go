package services

import (
	"errors"

	"github.com/pageza/alchemorsel-v1/internal/models"
)

// ResolveRecipe searches for a matching recipe; if not found, generates one using external APIs.
func ResolveRecipe(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	// TODO: 1. Search database for matching recipes using pgvector.
	// TODO: 2. If no close match, call external Deepseek API to generate a recipe.
	// TODO: 3. Call OpenAI API to obtain an embedding.
	// TODO: 4. Insert generated recipe into the database.
	return nil, nil, errors.New("not implemented")
}
