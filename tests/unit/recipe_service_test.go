package unit

import (
	"testing"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"github.com/stretchr/testify/assert"
)

// func TestListRecipes(t *testing.T) {
// 	recipes, err := services.ListRecipes()
// 	if err == nil {
// 		t.Error("Expected error for unimplemented ListRecipes, got nil")
// 	}
// 	if recipes != nil {
// 		t.Error("Expected nil recipes for unimplemented ListRecipes")
// 	}
// }

// MockRecipeRepository is a mock implementation of RecipeRepository for testing.
type MockRecipeRepository struct {
	SaveRecipeFunc func(recipe *models.Recipe) error
}

func (m *MockRecipeRepository) GetRecipe(id string) (*models.Recipe, error) { return nil, nil }
func (m *MockRecipeRepository) SaveRecipe(recipe *models.Recipe) error {
	return m.SaveRecipeFunc(recipe)
}
func (m *MockRecipeRepository) ListRecipes() ([]*models.Recipe, error)              { return nil, nil }
func (m *MockRecipeRepository) UpdateRecipe(id string, recipe *models.Recipe) error { return nil }
func (m *MockRecipeRepository) DeleteRecipe(id string) error                        { return nil }

func TestSaveRecipeSuccess(t *testing.T) {
	// Remove monkey patching (previously used to patch repositories.SaveRecipe)

	// Create a mock repository that simulates a successful save.
	mockRepo := &MockRecipeRepository{
		SaveRecipeFunc: func(recipe *models.Recipe) error {
			// Simulate successful save (e.g., do nothing and return nil)
			return nil
		},
	}

	// Create a new recipe with minimal fields.
	recipe := &models.Recipe{
		Title: "Test Recipe",
	}

	// Instantiate the service with the mock repository.
	service := &services.DefaultRecipeService{Repo: mockRepo}

	err := service.SaveRecipe(recipe)
	assert.Nil(t, err, "Expected no error on saving recipe")
	assert.False(t, recipe.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, recipe.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.True(t, recipe.UpdatedAt.Equal(recipe.CreatedAt) || recipe.UpdatedAt.After(recipe.CreatedAt),
		"UpdatedAt should be equal to or after CreatedAt")
}
