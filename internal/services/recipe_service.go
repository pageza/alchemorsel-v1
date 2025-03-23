package services

import (
	"errors"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
)

// ListRecipes retrieves a list of recipes.
func ListRecipes() ([]*models.Recipe, error) {
	// TODO: Implement logic to list or search recipes.
	return nil, errors.New("not implemented")
}

// GetRecipe retrieves a recipe by ID.
func GetRecipe(id string) (*models.Recipe, error) {
	// TODO: Implement logic to retrieve a specific recipe.
	return nil, errors.New("not implemented")
}

// SaveRecipe saves a new recipe into the database.
func SaveRecipe(recipe *models.Recipe) error {
	now := time.Now()
	if recipe.CreatedAt.IsZero() {
		recipe.CreatedAt = now
	}
	recipe.UpdatedAt = now
	// Insert recipe into the database via the repository.
	return repositories.SaveRecipe(recipe)
}

// UpdateRecipe updates an existing recipe.
func UpdateRecipe(id string, recipe *models.Recipe) error {
	// TODO: Implement logic to update a recipe.
	return errors.New("not implemented")
}

// DeleteRecipe deletes a recipe by ID.
func DeleteRecipe(id string) error {
	// TODO: Implement logic to delete a recipe.
	return errors.New("not implemented")
}
