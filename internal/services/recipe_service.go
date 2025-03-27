package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"go.uber.org/zap"
)

// RecipeServiceInterface defines the methods for recipe-related business logic.
type RecipeServiceInterface interface {
	GetRecipe(ctx context.Context, id string) (*models.Recipe, error)
	SaveRecipe(ctx context.Context, recipe *models.Recipe) error
	ListRecipes(ctx context.Context) ([]*models.Recipe, error)
	UpdateRecipe(ctx context.Context, recipe *models.Recipe) error
	DeleteRecipe(ctx context.Context, id string) error
	ResolveRecipe(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error)
}

// RecipeService is the implementation of RecipeServiceInterface.
type RecipeService struct {
	repo repositories.RecipeRepository
}

func NewRecipeService(repo repositories.RecipeRepository) RecipeServiceInterface {
	return &RecipeService{repo: repo}
}

func (s *RecipeService) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	return s.repo.GetRecipe(ctx, id)
}

func (s *RecipeService) SaveRecipe(ctx context.Context, recipe *models.Recipe) error {
	// Ensure the recipe has a valid UUID.
	if recipe.ID == "" {
		recipe.ID = uuid.New().String()
	}

	// Set timestamps if not already set
	if recipe.CreatedAt.IsZero() {
		recipe.CreatedAt = time.Now()
	}
	recipe.UpdatedAt = time.Now()

	// Log the operation
	zap.S().Infow("Saving recipe to the database", "title", recipe.Title, "id", recipe.ID)

	return s.repo.SaveRecipe(ctx, recipe)
}

func (s *RecipeService) ListRecipes(ctx context.Context) ([]*models.Recipe, error) {
	return s.repo.ListRecipes(ctx)
}

func (s *RecipeService) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	return s.repo.UpdateRecipe(ctx, recipe)
}

func (s *RecipeService) DeleteRecipe(ctx context.Context, id string) error {
	return s.repo.DeleteRecipe(ctx, id)
}

func (s *RecipeService) ResolveRecipe(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	return ResolveRecipe(query, attributes)
}
