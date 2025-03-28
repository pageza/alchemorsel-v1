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
) RecipeServiceInterface {
	return &RecipeService{
		repo:             repo,
		cuisineService:   cuisineService,
		dietService:      dietService,
		applianceService: applianceService,
		tagService:       tagService,
	}
}

func (s *RecipeService) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	return s.repo.GetRecipe(ctx, id)
}

func (s *RecipeService) SaveRecipe(ctx context.Context, recipe *models.Recipe) error {
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

func (s *RecipeService) ListRecipes(ctx context.Context) ([]*models.Recipe, error) {
	return s.repo.ListRecipes(ctx)
}

func (s *RecipeService) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
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

func (s *RecipeService) DeleteRecipe(ctx context.Context, id string) error {
	return s.repo.DeleteRecipe(ctx, id)
}

func (s *RecipeService) ResolveRecipe(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	return ResolveRecipe(query, attributes)
}
