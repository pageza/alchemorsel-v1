package unit

import (
	"testing"

	"recipeservice/internal/models"
	"recipeservice/internal/services"
)

func TestListRecipes(t *testing.T) {
	recipes, err := services.ListRecipes()
	if err == nil {
		t.Error("Expected error for unimplemented ListRecipes, got nil")
	}
	if recipes != nil {
		t.Error("Expected nil recipes for unimplemented ListRecipes")
	}
}

func TestCreateRecipe(t *testing.T) {
	recipe := &models.Recipe{
		ID:          "1",
		Title:       "Test Recipe",
		Ingredients: []string{"ingredient1", "ingredient2"},
		Steps:       []string{"step1", "step2"},
	}
	err := services.CreateRecipe(recipe)
	if err == nil {
		t.Error("Expected error for unimplemented CreateRecipe, got nil")
	}
}
