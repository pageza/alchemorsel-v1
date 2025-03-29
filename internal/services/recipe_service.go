package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"go.uber.org/zap"
)

// RecipeService defines the interface for recipe-related operations
type RecipeService interface {
	// SaveRecipe creates a new recipe
	SaveRecipe(ctx context.Context, recipe *models.Recipe) error

	// GetRecipe retrieves a recipe by ID
	GetRecipe(ctx context.Context, id string) (*models.Recipe, error)

	// UpdateRecipe updates an existing recipe
	UpdateRecipe(ctx context.Context, recipe *models.Recipe) error

	// DeleteRecipe deletes a recipe by ID
	DeleteRecipe(ctx context.Context, id string) error

	// ListRecipes retrieves a list of recipes with pagination and sorting
	ListRecipes(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error)

	// SearchRecipes searches for recipes based on query parameters
	SearchRecipes(ctx context.Context, query string, tags []string, difficulty string) ([]models.Recipe, error)

	// RateRecipe adds a rating to a recipe
	RateRecipe(ctx context.Context, recipeID string, rating float64) error

	// GetRecipeRatings retrieves all ratings for a recipe
	GetRecipeRatings(ctx context.Context, recipeID string) ([]float64, error)

	// ResolveRecipe resolves a recipe query with attributes
	ResolveRecipe(ctx context.Context, query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error)
}

// recipeService is the implementation of RecipeService
type recipeService struct {
	repo             repositories.RecipeRepository
	cuisineService   CuisineService
	dietService      DietService
	applianceService ApplianceService
	tagService       TagService
}

func NewRecipeService(
	repo repositories.RecipeRepository,
	cuisineService CuisineService,
	dietService DietService,
	applianceService ApplianceService,
	tagService TagService,
) RecipeService {
	return &recipeService{
		repo:             repo,
		cuisineService:   cuisineService,
		dietService:      dietService,
		applianceService: applianceService,
		tagService:       tagService,
	}
}

func (s *recipeService) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	return s.repo.GetRecipe(ctx, id)
}

func (s *recipeService) SaveRecipe(ctx context.Context, recipe *models.Recipe) error {
	if recipe == nil {
		return errors.New("recipe cannot be nil")
	}

	// Validate required fields
	if recipe.Title == "" {
		return errors.New("recipe title is required")
	}

	// Ensure the recipe has a valid UUID.
	if recipe.ID == "" {
		recipe.ID = uuid.New().String()
	}

	// Set timestamps if not already set
	if recipe.CreatedAt.IsZero() {
		recipe.CreatedAt = time.Now()
	}
	recipe.UpdatedAt = time.Now()

	// Handle cuisines
	if len(recipe.Cuisines) > 0 {
		for i, cuisine := range recipe.Cuisines {
			if cuisine.ID == "" {
				// Try to find existing cuisine by name or create a new one
				existingCuisine, err := s.cuisineService.GetOrCreate(ctx, cuisine.Name)
				if err != nil {
					return err
				}
				recipe.Cuisines[i] = *existingCuisine
			}
		}
	}

	// Handle diets
	if len(recipe.Diets) > 0 {
		for i, diet := range recipe.Diets {
			if diet.ID == "" {
				// Try to find existing diet by name or create a new one
				existingDiet, err := s.dietService.GetOrCreate(ctx, diet.Name)
				if err != nil {
					return err
				}
				recipe.Diets[i] = *existingDiet
			}
		}
	}

	// Handle appliances
	if len(recipe.Appliances) > 0 {
		for i, appliance := range recipe.Appliances {
			if appliance.ID == "" {
				// Try to find existing appliance by name or create a new one
				existingAppliance, err := s.applianceService.GetOrCreate(ctx, appliance.Name)
				if err != nil {
					return err
				}
				recipe.Appliances[i] = *existingAppliance
			}
		}
	}

	// Handle tags
	if len(recipe.Tags) > 0 {
		for i, tag := range recipe.Tags {
			if tag.ID == "" {
				// Try to find existing tag by name or create a new one
				existingTag, err := s.tagService.GetOrCreate(ctx, tag.Name)
				if err != nil {
					return err
				}
				recipe.Tags[i] = *existingTag
			}
		}
	}

	// Log the operation
	zap.S().Infow("Saving recipe to the database",
		"title", recipe.Title,
		"id", recipe.ID,
		"cuisines", len(recipe.Cuisines),
		"diets", len(recipe.Diets),
		"appliances", len(recipe.Appliances),
		"tags", len(recipe.Tags),
	)

	return s.repo.SaveRecipe(ctx, recipe)
}

func (s *recipeService) ListRecipes(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error) {
	return s.repo.ListRecipes(ctx, page, limit, sort, order)
}

func (s *recipeService) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	if recipe == nil {
		return errors.New("recipe cannot be nil")
	}

	// Validate required fields
	if recipe.Title == "" {
		return errors.New("recipe title is required")
	}

	// Handle cuisines
	if len(recipe.Cuisines) > 0 {
		for i, cuisine := range recipe.Cuisines {
			if cuisine.ID == "" {
				// Try to find existing cuisine by name or create a new one
				existingCuisine, err := s.cuisineService.GetOrCreate(ctx, cuisine.Name)
				if err != nil {
					return err
				}
				recipe.Cuisines[i] = *existingCuisine
			}
		}
	}

	// Handle diets
	if len(recipe.Diets) > 0 {
		for i, diet := range recipe.Diets {
			if diet.ID == "" {
				// Try to find existing diet by name or create a new one
				existingDiet, err := s.dietService.GetOrCreate(ctx, diet.Name)
				if err != nil {
					return err
				}
				recipe.Diets[i] = *existingDiet
			}
		}
	}

	// Handle appliances
	if len(recipe.Appliances) > 0 {
		for i, appliance := range recipe.Appliances {
			if appliance.ID == "" {
				// Try to find existing appliance by name or create a new one
				existingAppliance, err := s.applianceService.GetOrCreate(ctx, appliance.Name)
				if err != nil {
					return err
				}
				recipe.Appliances[i] = *existingAppliance
			}
		}
	}

	// Handle tags
	if len(recipe.Tags) > 0 {
		for i, tag := range recipe.Tags {
			if tag.ID == "" {
				// Try to find existing tag by name or create a new one
				existingTag, err := s.tagService.GetOrCreate(ctx, tag.Name)
				if err != nil {
					return err
				}
				recipe.Tags[i] = *existingTag
			}
		}
	}

	// Log the operation
	zap.S().Infow("Updating recipe in the database",
		"title", recipe.Title,
		"id", recipe.ID,
		"cuisines", len(recipe.Cuisines),
		"diets", len(recipe.Diets),
		"appliances", len(recipe.Appliances),
		"tags", len(recipe.Tags),
	)

	return s.repo.UpdateRecipe(ctx, recipe)
}

func (s *recipeService) DeleteRecipe(ctx context.Context, id string) error {
	return s.repo.DeleteRecipe(ctx, id)
}

func (s *recipeService) SearchRecipes(ctx context.Context, query string, tags []string, difficulty string) ([]models.Recipe, error) {
	return s.repo.SearchRecipes(ctx, query, tags, difficulty)
}

func (s *recipeService) RateRecipe(ctx context.Context, recipeID string, rating float64) error {
	return s.repo.RateRecipe(ctx, recipeID, rating)
}

func (s *recipeService) GetRecipeRatings(ctx context.Context, recipeID string) ([]float64, error) {
	return s.repo.GetRecipeRatings(ctx, recipeID)
}

func (s *recipeService) ResolveRecipe(ctx context.Context, query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	return s.repo.ResolveRecipe(ctx, query, attributes)
}
