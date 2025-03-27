package repositories

import (
	"context"
	"os"
	"time"

	"errors"

	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var testRecipes = make(map[string]*models.Recipe)

// ListRecipes retrieves a list of recipes from the database.
func ListRecipes() ([]*models.Recipe, error) {
	if os.Getenv("TEST_MODE") != "" || DB == nil {
		var recipes []*models.Recipe
		for _, r := range testRecipes {
			recipes = append(recipes, r)
		}
		return recipes, nil
	}
	// TODO: Implement database operation to list recipes.
	return nil, nil
}

// GetRecipe retrieves a recipe by ID.
func GetRecipe(id string) (*models.Recipe, error) {
	if os.Getenv("TEST_MODE") != "" || DB == nil {
		if recipe, ok := testRecipes[id]; ok {
			return recipe, nil
		}
		return nil, errors.New("recipe not found")
	}
	var recipe models.Recipe
	if err := DB.First(&recipe, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &recipe, nil
}

// SaveRecipe inserts a new recipe into the database.
func SaveRecipe(recipe *models.Recipe) error {
	if recipe.ID == "" {
		recipe.ID = uuid.New().String()
	}

	// Set timestamps if not already set
	if recipe.CreatedAt.IsZero() {
		recipe.CreatedAt = time.Now()
	}
	recipe.UpdatedAt = time.Now()

	// Ensure ingredients and steps are JSON
	if len(recipe.Ingredients) == 0 {
		recipe.Ingredients = []byte("[]")
	}
	if len(recipe.Steps) == 0 {
		recipe.Steps = []byte("[]")
	}

	// Log the save operation.
	logrus.WithFields(logrus.Fields{"recipeID": recipe.ID, "recipeTitle": recipe.Title}).Info("Global SaveRecipe called")

	// Use in-memory store for tests.
	if os.Getenv("TEST_MODE") == "true" {
		testRecipes[recipe.ID] = recipe
		return nil
	}

	if err := DB.Create(recipe).Error; err != nil {
		return err
	}
	return nil
}

// UpdateRecipe modifies an existing recipe in the database.
func UpdateRecipe(id string, recipe *models.Recipe) error {
	// TODO: Implement database operation to update a recipe.
	return nil
}

// DeleteRecipe removes a recipe by ID from the database.
func DeleteRecipe(id string) error {
	// TODO: Implement database operation to delete a recipe.
	return nil
}

type RecipeRepository interface {
	GetRecipe(ctx context.Context, id string) (*models.Recipe, error)
	SaveRecipe(ctx context.Context, recipe *models.Recipe) error
	ListRecipes(ctx context.Context) ([]*models.Recipe, error)
	UpdateRecipe(ctx context.Context, recipe *models.Recipe) error
	DeleteRecipe(ctx context.Context, id string) error
}

type DefaultRecipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) RecipeRepository {
	return &DefaultRecipeRepository{db: db}
}

func (r *DefaultRecipeRepository) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	var recipe models.Recipe
	if err := r.db.WithContext(ctx).First(&recipe, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &recipe, nil
}

func (r *DefaultRecipeRepository) SaveRecipe(ctx context.Context, recipe *models.Recipe) error {
	return r.db.WithContext(ctx).Create(recipe).Error
}

func (r *DefaultRecipeRepository) ListRecipes(ctx context.Context) ([]*models.Recipe, error) {
	var recipes []*models.Recipe
	if err := r.db.WithContext(ctx).Find(&recipes).Error; err != nil {
		return nil, err
	}
	return recipes, nil
}

func (r *DefaultRecipeRepository) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	return r.db.WithContext(ctx).Save(recipe).Error
}

func (r *DefaultRecipeRepository) DeleteRecipe(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Recipe{}, "id = ?", id).Error
}
