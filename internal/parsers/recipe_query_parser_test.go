package parsers

import (
	"testing"
)

func TestParseRecipeQueryEmpty(t *testing.T) {
	_, err := ParseRecipeQuery("   ")
	if err == nil {
		t.Error("Expected error for empty query, got nil")
	}
}

// TestParseRecipeQueryMexicanVegan tests the ParseRecipeQuery function for a Mexican vegan dish.
func TestParseRecipeQueryMexicanVegan(t *testing.T) {
	query := "I want a Mexican vegan dish with tomatoes and onions"
	parsed, err := ParseRecipeQuery(query)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if parsed == nil { //nolint:SA5011
		t.Fatal("ParseRecipeQuery returned nil pointer")
	}

	if parsed.Cuisine != "mexican" {
		t.Errorf("Expected cuisine 'mexican', got %s", parsed.Cuisine)
	}

	if parsed.DietaryRestrictions != "vegan" {
		t.Errorf("Expected dietary 'vegan', got %s", parsed.DietaryRestrictions)
	}

	if len(parsed.Ingredients) == 0 {
		t.Error("Expected non-empty ingredients list, but got empty")
	}
}
