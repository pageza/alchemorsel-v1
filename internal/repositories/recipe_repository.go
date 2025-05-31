package repositories

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/errors"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
		return errors.NewValidationError("recipe cannot be nil")
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
		return errors.NewValidationError("recipe title is required")
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
			return errors.NewDatabaseError("failed to save recipe").WithFields(zap.String("recipe_title", recipe.Title))
		}
		logger.Info("saved recipe to database")
		return nil
	})

	return err
}

// UpdateRecipe modifies an existing recipe in the database.
func UpdateRecipe(id string, recipe *models.Recipe) error {
	if id == "" {
		return errors.NewValidationError("recipe ID is required")
	}
	if recipe == nil {
		return errors.NewValidationError("recipe cannot be nil")
	}
	
	if os.Getenv("TEST_MODE") == "true" {
		if _, exists := testRecipes[id]; !exists {
			return errors.NewNotFoundError("recipe not found")
		}
		testRecipes[id] = recipe
		return nil
	}
	
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(recipe).Error; err != nil {
			return errors.NewDatabaseError("failed to update recipe").WithFields(zap.String("recipe_id", id))
		}
		return nil
	})
}

// DeleteRecipe removes a recipe by ID from the database.
func DeleteRecipe(id string) error {
	if id == "" {
		return errors.NewValidationError("recipe ID is required")
	}
	
	if os.Getenv("TEST_MODE") == "true" {
		if _, exists := testRecipes[id]; !exists {
			return errors.NewNotFoundError("recipe not found")
		}
		delete(testRecipes, id)
		return nil
	}
	
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Recipe{}, "id = ?", id).Error; err != nil {
			return errors.NewDatabaseError("failed to delete recipe").WithFields(zap.String("recipe_id", id))
		}
		return nil
	})
}

type RecipeRepository interface {
	GetRecipe(ctx context.Context, id string) (*models.Recipe, error)
	SaveRecipe(ctx context.Context, recipe *models.Recipe) error
	ListRecipes(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error)
	UpdateRecipe(ctx context.Context, recipe *models.Recipe) error
	DeleteRecipe(ctx context.Context, id string) error
	SearchRecipes(ctx context.Context, query string, tags []string, difficulty string) ([]models.Recipe, error)
	RateRecipe(ctx context.Context, recipeID string, rating float64) error
	GetRecipeRatings(ctx context.Context, recipeID string) ([]float64, error)
	ResolveRecipe(ctx context.Context, query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error)
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
	if recipe == nil {
		return errors.NewValidationError("recipe cannot be nil")
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
		return errors.NewValidationError("recipe title is required")
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

	// Use transaction for database operations
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(recipe).Error; err != nil {
			logger.WithError(err).Error("failed to save recipe to database")
			return errors.NewDatabaseError("failed to save recipe").WithFields(zap.String("recipe_title", recipe.Title))
		}
		logger.Info("saved recipe to database")
		return nil
	})

	return err
}

func (r *DefaultRecipeRepository) ListRecipes(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error) {
	var recipes []models.Recipe
	query := r.db.WithContext(ctx).
		Preload("Cuisines").
		Preload("Diets").
		Preload("Appliances").
		Preload("Tags")

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Apply sorting
	if sort != "" {
		if order != "asc" && order != "desc" {
			order = "desc"
		}
		query = query.Order(fmt.Sprintf("%s %s", sort, order))
	}

	if err := query.Find(&recipes).Error; err != nil {
		return nil, err
	}

	return recipes, nil
}

func (r *DefaultRecipeRepository) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	if recipe == nil {
		return errors.NewValidationError("recipe cannot be nil")
	}

	logger := logrus.WithFields(logrus.Fields{
		"operation": "UpdateRecipe",
		"recipe_id": recipe.ID,
		"title":     recipe.Title,
	})
	logger.Info("updating recipe")

	// Validate required fields
	if recipe.Title == "" {
		logger.Error("recipe title is required")
		return errors.NewValidationError("recipe title is required")
	}

	recipe.UpdatedAt = time.Now()

	// Use transaction for database operations
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(recipe).Error; err != nil {
			logger.WithError(err).Error("failed to update recipe in database")
			return errors.NewDatabaseError("failed to update recipe").WithFields(zap.String("recipe_id", recipe.ID))
		}
		logger.Info("updated recipe in database")
		return nil
	})

	return err
}

func (r *DefaultRecipeRepository) DeleteRecipe(ctx context.Context, id string) error {
	logger := logrus.WithFields(logrus.Fields{
		"operation": "DeleteRecipe",
		"recipe_id": id,
	})
	logger.Info("deleting recipe")

	if id == "" {
		logger.Error("recipe ID is required")
		return errors.NewValidationError("recipe ID is required")
	}

	// Use transaction for database operations
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Recipe{}, "id = ?", id).Error; err != nil {
			logger.WithError(err).Error("failed to delete recipe from database")
			return errors.NewDatabaseError("failed to delete recipe").WithFields(zap.String("recipe_id", id))
		}
		logger.Info("deleted recipe from database")
		return nil
	})

	return err
}

func (r *DefaultRecipeRepository) SearchRecipes(ctx context.Context, query string, tags []string, difficulty string) ([]models.Recipe, error) {
	var recipes []models.Recipe
	db := r.db.WithContext(ctx).
		Preload("Cuisines").
		Preload("Diets").
		Preload("Appliances").
		Preload("Tags")

	if query != "" {
		db = db.Where("title LIKE ? OR description LIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if len(tags) > 0 {
		db = db.Joins("JOIN recipe_tags ON recipes.id = recipe_tags.recipe_id").
			Joins("JOIN tags ON recipe_tags.tag_id = tags.id").
			Where("tags.name IN ?", tags)
	}

	if difficulty != "" {
		db = db.Where("difficulty = ?", difficulty)
	}

	if err := db.Find(&recipes).Error; err != nil {
		return nil, err
	}

	return recipes, nil
}

func (r *DefaultRecipeRepository) RateRecipe(ctx context.Context, recipeID string, rating float64) error {
	logger := logrus.WithFields(logrus.Fields{
		"operation": "RateRecipe",
		"recipe_id": recipeID,
		"rating":    rating,
	})
	logger.Info("rating recipe")

	if recipeID == "" {
		logger.Error("recipe ID is required")
		return errors.NewValidationError("recipe ID is required")
	}

	// Use transaction for database operations
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var recipe models.Recipe
		if err := tx.First(&recipe, "id = ?", recipeID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				logger.Error("recipe not found")
				return errors.NewNotFoundError("recipe not found").WithFields(zap.String("recipe_id", recipeID))
			}
			logger.WithError(err).Error("failed to retrieve recipe from database")
			return errors.NewDatabaseError("failed to retrieve recipe").WithFields(zap.String("recipe_id", recipeID))
		}

		// Update rating
		recipe.AverageRating = ((recipe.AverageRating * float64(recipe.RatingCount)) + rating) / float64(recipe.RatingCount+1)
		recipe.RatingCount++

		if err := tx.Save(&recipe).Error; err != nil {
			logger.WithError(err).Error("failed to update recipe rating in database")
			return errors.NewDatabaseError("failed to update recipe rating").WithFields(zap.String("recipe_id", recipeID))
		}

		logger.Info("updated recipe rating in database")
		return nil
	})

	return err
}

func (r *DefaultRecipeRepository) GetRecipeRatings(ctx context.Context, recipeID string) ([]float64, error) {
	logger := logrus.WithFields(logrus.Fields{
		"operation": "GetRecipeRatings",
		"recipe_id": recipeID,
	})
	logger.Info("retrieving recipe ratings")

	if recipeID == "" {
		logger.Error("recipe ID is required")
		return nil, errors.NewValidationError("recipe ID is required")
	}

	var recipe models.Recipe
	if err := r.db.WithContext(ctx).First(&recipe, "id = ?", recipeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("recipe not found")
			return nil, errors.NewNotFoundError("recipe not found").WithFields(zap.String("recipe_id", recipeID))
		}
		logger.WithError(err).Error("failed to retrieve recipe from database")
		return nil, errors.NewDatabaseError("failed to retrieve recipe").WithFields(zap.String("recipe_id", recipeID))
	}

	// For now, we'll just return a slice with the average rating repeated RatingCount times
	ratings := make([]float64, recipe.RatingCount)
	for i := range ratings {
		ratings[i] = recipe.AverageRating
	}

	return ratings, nil
}

func (r *DefaultRecipeRepository) ResolveRecipe(ctx context.Context, query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	logger := logrus.WithFields(logrus.Fields{
		"operation": "ResolveRecipe",
		"query":     query,
	})
	logger.Info("resolving recipe")

	// First, try to find an exact match
	var exactMatch models.Recipe
	db := r.db.WithContext(ctx).
		Preload("Cuisines").
		Preload("Diets").
		Preload("Appliances").
		Preload("Tags")

	if err := db.Where("title = ?", query).First(&exactMatch).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.WithError(err).Error("failed to search for exact match")
			return nil, nil, errors.NewDatabaseError("failed to search for exact match").WithFields(zap.String("query", query))
		}
	} else {
		return &exactMatch, nil, nil
	}

	// If no exact match, find similar recipes
	var similarRecipes []*models.Recipe
	if err := db.Where("title LIKE ?", "%"+query+"%").
		Or("description LIKE ?", "%"+query+"%").
		Find(&similarRecipes).Error; err != nil {
		logger.WithError(err).Error("failed to search for similar recipes")
		return nil, nil, errors.NewDatabaseError("failed to search for similar recipes").WithFields(zap.String("query", query))
	}

	return nil, similarRecipes, nil
}
