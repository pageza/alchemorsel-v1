package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RecipeError represents a domain-specific error for recipe operations
type RecipeError struct {
	Code    int
	Message string
	Err     error
}

func (e *RecipeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("recipe error: %s (code: %d): %v", e.Message, e.Code, e.Err)
	}
	return fmt.Sprintf("recipe error: %s (code: %d)", e.Message, e.Code)
}

var (
	ErrRecipeNotFound = &RecipeError{Code: 404, Message: "recipe not found"}
	ErrInvalidRecipe  = &RecipeError{Code: 400, Message: "invalid recipe data"}
	ErrDBOperation    = &RecipeError{Code: 500, Message: "database operation failed"}
)

var testRecipes = make(map[string]*models.Recipe)

// ListRecipes retrieves a list of recipes from the database.
func ListRecipes() ([]*models.Recipe, error) {
	logger := logrus.WithField("operation", "ListRecipes")
	logger.Info("retrieving recipes")

	if os.Getenv("TEST_MODE") != "" || DB == nil {
		var recipes []*models.Recipe
		for _, r := range testRecipes {
			recipes = append(recipes, r)
		}
		logger.WithField("count", len(recipes)).Info("retrieved recipes from test store")
		return recipes, nil
	}

	var recipes []*models.Recipe
	if err := DB.Find(&recipes).Error; err != nil {
		logger.WithError(err).Error("failed to retrieve recipes from database")
		return nil, &RecipeError{Code: 500, Message: "failed to retrieve recipes", Err: err}
	}

	logger.WithField("count", len(recipes)).Info("retrieved recipes from database")
	return recipes, nil
}

// GetRecipe retrieves a recipe by ID.
func GetRecipe(id string) (*models.Recipe, error) {
	logger := logrus.WithFields(logrus.Fields{
		"operation": "GetRecipe",
		"recipe_id": id,
	})
	logger.Info("retrieving recipe")

	if id == "" {
		logger.Error("empty recipe ID provided")
		return nil, &RecipeError{Code: 400, Message: "recipe ID is required"}
	}

	if os.Getenv("TEST_MODE") != "" || DB == nil {
		if recipe, ok := testRecipes[id]; ok {
			logger.Info("retrieved recipe from test store")
			return recipe, nil
		}
		logger.Error("recipe not found in test store")
		return nil, ErrRecipeNotFound
	}

	var recipe models.Recipe
	if err := DB.First(&recipe, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("recipe not found in database")
			return nil, ErrRecipeNotFound
		}
		logger.WithError(err).Error("failed to retrieve recipe from database")
		return nil, &RecipeError{Code: 500, Message: "failed to retrieve recipe", Err: err}
	}

	logger.Info("retrieved recipe from database")
	return &recipe, nil
}

// SaveRecipe inserts a new recipe into the database.
func SaveRecipe(recipe *models.Recipe) error {
	if recipe == nil {
		return &RecipeError{Code: 400, Message: "recipe cannot be nil"}
	}

	logger := logrus.WithFields(logrus.Fields{
		"operation": "SaveRecipe",
		"recipe_id": recipe.ID,
		"title":     recipe.Title,
	})
	logger.Info("saving recipe")

	// Validate required fields
	if recipe.Title == "" {
		logger.Error("recipe title is required")
		return &RecipeError{Code: 400, Message: "recipe title is required"}
	}

	if recipe.ID == "" {
		recipe.ID = uuid.New().String()
		logger.WithField("new_id", recipe.ID).Info("generated new recipe ID")
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

	// Use in-memory store for tests
	if os.Getenv("TEST_MODE") == "true" {
		testRecipes[recipe.ID] = recipe
		logger.Info("saved recipe to test store")
		return nil
	}

	// Use transaction for database operations
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(recipe).Error; err != nil {
			logger.WithError(err).Error("failed to save recipe to database")
			return &RecipeError{Code: 500, Message: "failed to save recipe", Err: err}
		}
		logger.Info("saved recipe to database")
		return nil
	})

	return err
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
	if err := r.db.WithContext(ctx).
		Preload("Cuisines").
		Preload("Diets").
		Preload("Appliances").
		Preload("Tags").
		First(&recipe, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &recipe, nil
}

func (r *DefaultRecipeRepository) SaveRecipe(ctx context.Context, recipe *models.Recipe) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Set timestamps
		if recipe.CreatedAt.IsZero() {
			recipe.CreatedAt = time.Now()
		}
		recipe.UpdatedAt = time.Now()

		// Generate UUID if not set
		if recipe.ID == "" {
			recipe.ID = uuid.New().String()
		}

		// Create the recipe first
		if err := tx.Create(recipe).Error; err != nil {
			return err
		}

		// Handle many-to-many relationships
		if len(recipe.Cuisines) > 0 {
			if err := tx.Model(recipe).Association("Cuisines").Replace(recipe.Cuisines); err != nil {
				return err
			}
		}

		if len(recipe.Diets) > 0 {
			if err := tx.Model(recipe).Association("Diets").Replace(recipe.Diets); err != nil {
				return err
			}
		}

		if len(recipe.Appliances) > 0 {
			if err := tx.Model(recipe).Association("Appliances").Replace(recipe.Appliances); err != nil {
				return err
			}
		}

		if len(recipe.Tags) > 0 {
			if err := tx.Model(recipe).Association("Tags").Replace(recipe.Tags); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *DefaultRecipeRepository) ListRecipes(ctx context.Context) ([]*models.Recipe, error) {
	var recipes []*models.Recipe
	if err := r.db.WithContext(ctx).
		Preload("Cuisines").
		Preload("Diets").
		Preload("Appliances").
		Preload("Tags").
		Find(&recipes).Error; err != nil {
		return nil, err
	}
	return recipes, nil
}

func (r *DefaultRecipeRepository) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update timestamp
		recipe.UpdatedAt = time.Now()

		// Update the recipe first
		if err := tx.Save(recipe).Error; err != nil {
			return err
		}

		// Handle many-to-many relationships
		if len(recipe.Cuisines) > 0 {
			if err := tx.Model(recipe).Association("Cuisines").Replace(recipe.Cuisines); err != nil {
				return err
			}
		}

		if len(recipe.Diets) > 0 {
			if err := tx.Model(recipe).Association("Diets").Replace(recipe.Diets); err != nil {
				return err
			}
		}

		if len(recipe.Appliances) > 0 {
			if err := tx.Model(recipe).Association("Appliances").Replace(recipe.Appliances); err != nil {
				return err
			}
		}

		if len(recipe.Tags) > 0 {
			if err := tx.Model(recipe).Association("Tags").Replace(recipe.Tags); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *DefaultRecipeRepository) DeleteRecipe(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		recipe := &models.Recipe{ID: id}

		// Clear all associations first
		if err := tx.Model(recipe).Association("Cuisines").Clear(); err != nil {
			return err
		}
		if err := tx.Model(recipe).Association("Diets").Clear(); err != nil {
			return err
		}
		if err := tx.Model(recipe).Association("Appliances").Clear(); err != nil {
			return err
		}
		if err := tx.Model(recipe).Association("Tags").Clear(); err != nil {
			return err
		}

		// Delete the recipe
		return tx.Delete(recipe).Error
	})
}
