package repositories

import (
	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/models"
)

// ListRecipes retrieves a list of recipes from the database.
func ListRecipes() ([]*models.Recipe, error) {
	// TODO: Implement database operation to list recipes.
	return nil, nil
}

// GetRecipe retrieves a recipe by ID.
func GetRecipe(id string) (*models.Recipe, error) {
	var recipe models.Recipe
	if err := DB.First(&recipe, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &recipe, nil
}

// SaveRecipe inserts a new recipe into the database.
func SaveRecipe(recipe *models.Recipe) error {
	return DB.Create(recipe).Error
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

// RecipeRepository defines the repository interface for recipes.
type RecipeRepository interface {
	GetRecipe(id string) (*models.Recipe, error)
	SaveRecipe(recipe *models.Recipe) error
	ListRecipes() ([]*models.Recipe, error)
	UpdateRecipe(id string, recipe *models.Recipe) error
	DeleteRecipe(id string) error
}

// DefaultRecipeRepository is the default implementation of RecipeRepository
type DefaultRecipeRepository struct{}

func (r *DefaultRecipeRepository) GetRecipe(id string) (*models.Recipe, error) {
	return GetRecipe(id)
}

func (r *DefaultRecipeRepository) SaveRecipe(recipe *models.Recipe) error {
	if recipe.ID == "" {
		recipe.ID = uuid.NewString()
	}
	return SaveRecipe(recipe)
}

func (r *DefaultRecipeRepository) ListRecipes() ([]*models.Recipe, error) {
	return ListRecipes()
}

func (r *DefaultRecipeRepository) UpdateRecipe(id string, recipe *models.Recipe) error {
	return UpdateRecipe(id, recipe)
}

func (r *DefaultRecipeRepository) DeleteRecipe(id string) error {
	return DeleteRecipe(id)
}
