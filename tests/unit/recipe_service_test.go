package unit

import (
	"encoding/json"
	"testing"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
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
	ingredients, err := json.Marshal([]string{"ingredient1", "ingredient2"})
	if err != nil {
		t.Fatalf("Error marshalling ingredients: %v", err)
	}
	steps, err := json.Marshal([]string{"step1", "step2"})
	if err != nil {
		t.Fatalf("Error marshalling steps: %v", err)
	}

	recipe := &models.Recipe{
		ID:          "1",
		Title:       "Test Recipe",
		Ingredients: ingredients,
		Steps:       steps,
	}
	err = services.CreateRecipe(recipe)
	if err == nil {
		t.Error("Expected error for unimplemented CreateRecipe, got nil")
	}
}
