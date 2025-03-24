package unit

import (
	"testing"

	"bou.ke/monkey"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
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

func TestSaveRecipeSuccess(t *testing.T) {
	// Monkey-patch repositories.SaveRecipe to simulate a successful database save.
	patch := monkey.Patch(repositories.SaveRecipe, func(recipe *models.Recipe) error {
		return nil
	})
	defer patch.Unpatch()

	// Create a new recipe with minimal fields.
	recipe := &models.Recipe{
		Title: "Test Recipe",
	}

	// Call the service SaveRecipe (this will set CreatedAt and UpdatedAt).
	err := services.SaveRecipe(recipe)
	assert.Nil(t, err, "Expected no error on saving recipe")
	assert.False(t, recipe.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, recipe.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.True(t, recipe.UpdatedAt.Equal(recipe.CreatedAt) || recipe.UpdatedAt.After(recipe.CreatedAt),
		"UpdatedAt should be equal to or after CreatedAt")
}
