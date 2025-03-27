package services

import (
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"go.uber.org/zap"
)

// RecipeServiceInterface defines the methods for recipe-related business logic.
type RecipeServiceInterface interface {
	GetRecipe(id string) (*models.Recipe, error)
	SaveRecipe(recipe *models.Recipe) error
	ListRecipes() ([]*models.Recipe, error)
	UpdateRecipe(id string, recipe *models.Recipe) error
	DeleteRecipe(id string) error
}

// DefaultRecipeService is the default implementation of RecipeServiceInterface.
// It delegates calls to the repositories layer and adds business logic where necessary.
type DefaultRecipeService struct {
	Repo repositories.RecipeRepository
}

func (s *DefaultRecipeService) GetRecipe(id string) (*models.Recipe, error) {
	return s.Repo.GetRecipe(id)
}

func (s *DefaultRecipeService) SaveRecipe(recipe *models.Recipe) error {
	now := time.Now()
	if recipe.CreatedAt.IsZero() {
		recipe.CreatedAt = now
	}
	recipe.UpdatedAt = now

	// Log the database save operation for monitoring.
	zap.S().Infow("Saving recipe to the database", "title", recipe.Title)

	return s.Repo.SaveRecipe(recipe)
}

func (s *DefaultRecipeService) ListRecipes() ([]*models.Recipe, error) {
	return s.Repo.ListRecipes()
}

func (s *DefaultRecipeService) UpdateRecipe(id string, recipe *models.Recipe) error {
	return s.Repo.UpdateRecipe(id, recipe)
}

func (s *DefaultRecipeService) DeleteRecipe(id string) error {
	return s.Repo.DeleteRecipe(id)
}
