package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/integrations"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/parsers"
)

// RecipeResolutionService defines the functions for the multi-step recipe resolution flow.
// It handles database searches for exact and close matches, builds composite prompts that include
// the user's query, prompt instructions, expected response format, and user profile data,
// and sends the composite prompt to an external model.

type RecipeResolutionService interface {
	// FindExactMatch should search the database for an exact recipe match for the given parsed query.
	FindExactMatch(ctx context.Context, parsedQuery *parsers.ParsedQuery) (string, error)
	// FindCloseMatches should return a list of close matches from the recipe database based on the parsed query.
	FindCloseMatches(ctx context.Context, parsedQuery *parsers.ParsedQuery) ([]string, error)
	// BuildCompositePrompt creates a composite prompt using the user's query, prompt instructions,
	// expected response format, and additional profile data (e.g., allergen and diet restrictions).
	BuildCompositePrompt(query string, promptInstructions string, expectedResponseFormat string, profile map[string]interface{}) (string, error)
	// ResolveRecipeByModel sends the composite prompt to the external model and returns
	// a candidate recipe along with alternative proposals.
	ResolveRecipeByModel(ctx context.Context, compositePrompt string) (string, []string, error)
}

// recipeResolutionService is a default implementation of RecipeResolutionService.
// All methods are currently scaffolded with TODO comments.

type recipeResolutionService struct{}

// NewRecipeResolutionService creates a new instance of RecipeResolutionService.
func NewRecipeResolutionService() RecipeResolutionService {
	return &recipeResolutionService{}
}

func (s *recipeResolutionService) FindExactMatch(ctx context.Context, parsedQuery *parsers.ParsedQuery) (string, error) {
	// TODO: Implement logic to search for an exact match in the recipe database using the parsed query with GORM filtering.
	return "", nil
}

func (s *recipeResolutionService) FindCloseMatches(ctx context.Context, parsedQuery *parsers.ParsedQuery) ([]string, error) {
	// Pseudo-code for building a GORM query:
	// -----------------------------------------------------------
	// var recipes []models.Recipe
	// db := GetDBFromContext(ctx) // Assume a function that returns *gorm.DB from the context
	// query := db.Model(&models.Recipe{})
	//
	// // Filter by cuisine
	// if parsedQuery.Cuisine != "" && parsedQuery.Cuisine != "unknown" {
	//     query = query.Where("cuisine = ?", parsedQuery.Cuisine)
	// }
	//
	// // Filter by dietary restrictions
	// if parsedQuery.DietaryRestrictions != "" && parsedQuery.DietaryRestrictions != "none" {
	//     query = query.Where("dietary_restrictions = ?", parsedQuery.DietaryRestrictions)
	// }
	//
	// // Filter by included ingredients (this is simplistic; real implementation may need more robust matching)
	// if len(parsedQuery.Ingredients) > 0 {
	//     // For example, joining ingredients with wildcards
	//     query = query.Where("ingredients LIKE ?", "%" + strings.Join(parsedQuery.Ingredients, "%") + "%")
	// }
	//
	// // Exclude recipes containing any excluded ingredients
	// if len(parsedQuery.Exclusions) > 0 {
	//     for _, exclusion := range parsedQuery.Exclusions {
	//         query = query.Where("ingredients NOT LIKE ?", "%" + exclusion + "%")
	//     }
	// }
	//
	// // Additional optional filters:
	// if parsedQuery.Timing > 0 {
	//     // Assuming Timing is the maximum total time (prep_time + cooking_time)
	//     query = query.Where("(prep_time + cooking_time) <= ?", parsedQuery.Timing)
	// }
	// if parsedQuery.Servings > 0 {
	//     query = query.Where("servings = ?", parsedQuery.Servings)
	// }
	// if parsedQuery.Difficulty != "" {
	//     query = query.Where("difficulty = ?", parsedQuery.Difficulty)
	// }
	// if parsedQuery.CaloriesPerServing > 0 {
	//     query = query.Where("calories_per_serving <= ?", parsedQuery.CaloriesPerServing)
	// }
	// if parsedQuery.ServingSize != "" {
	//     query = query.Where("serving_size = ?", parsedQuery.ServingSize)
	// }
	//
	// err := query.Find(&recipes).Error
	// if err != nil {
	//     return nil, err
	// }
	//
	// // Convert recipes to a slice of identifiers or summaries (for now, assume recipe titles)
	// var recipeTitles []string
	// for _, r := range recipes {
	//     recipeTitles = append(recipeTitles, r.Title)
	// }
	// return recipeTitles, nil
	// -----------------------------------------------------------

	// Currently, this is a placeholder for the actual GORM query logic
	return nil, nil
}

// Define hardcoded defaults for prompt instructions and expected response format
const (
	DefaultExpectedResponseFormat = "{\"title\": string, \"description\": string, \"ingredients\": [{\"name\": string, \"amount\": number, \"unit\": string}], \"steps\": [{\"order\": number, \"description\": string}], \"nutritional_info\": string, \"allergy_disclaimer\": string, \"cuisines\": [string], \"diets\": [string], \"appliances\": [string], \"tags\": [string], \"images\": [string], \"difficulty\": string, \"prep_time\": number, \"cooking_time\": number, \"servings\": number, \"approved\": boolean}"
	DefaultPromptInstructions     = "Act as a professional personal chef. Provide detailed, step-by-step recipes with clear instructions and precise measurements."
)

// BuildCompositePrompt constructs the composite prompt using the user's query and profile details with hardcoded prompt instructions and expected response format.
func (s *recipeResolutionService) BuildCompositePrompt(query string, promptInstructions string, expectedResponseFormat string, profile map[string]interface{}) (string, error) {
	// Check if promptInstructions and expectedResponseFormat are provided; if not, use the defaults
	if promptInstructions == "" {
		promptInstructions = DefaultPromptInstructions
		fmt.Println("PromptInstructions missing in request, using default prompt instructions")
	}
	if expectedResponseFormat == "" {
		expectedResponseFormat = DefaultExpectedResponseFormat
		fmt.Println("ExpectedResponseFormat missing in request, using default expected response format")
	}

	compositePrompt := "=== Composite Prompt for Recipe Resolution ===\n\n"
	compositePrompt += "User Query:\n" + query + "\n\n"
	compositePrompt += "Prompt Instructions:\n" + promptInstructions + "\n\n"
	compositePrompt += "Expected Response Format:\n" + expectedResponseFormat + "\n\n"
	compositePrompt += "User Profile:\n"
	for key, value := range profile {
		compositePrompt += " - " + key + ": " + fmt.Sprintf("%v", value) + "\n"
	}
	compositePrompt += "\n=== End of Prompt ==="
	return compositePrompt, nil
}

func (s *recipeResolutionService) ResolveRecipeByModel(ctx context.Context, compositePrompt string) (string, []string, error) {
	response, err := callExternalAPI(compositePrompt)
	if err != nil {
		return "", nil, err
	}
	// For now, just return the raw response as the candidate and an empty slice for alternatives.
	return response, []string{}, nil
}

// ResolveRecipe searches for a matching recipe; if not found, generates one using external APIs.
func ResolveRecipe(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	// Construct a prompt by prefixing the user's request with instructions
	promptPrefix := "You are a professional chef's assistant to help the chef create dishes using the parameters specified. The expected response format is JSON with the following keys: title (string), description (string), ingredients (array of objects with keys: name, amount, unit), steps (array of objects with keys: order, description), nutritional_info (string), allergy_disclaimer (string), cuisines (array of strings), diets (array of strings), appliances (array of strings), tags (array of strings), images (array of strings), difficulty (string), prep_time (integer), cooking_time (integer), servings (integer), approved (boolean)."

	// Build the prompt
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return nil, nil, err
	}
	prompt := promptPrefix + "\nParameters:\n" + string(attributesJSON)
	if query != "" {
		prompt += "\nTitle: " + query
	}

	// Call the external API to generate the recipe
	generatedResponse, err := callExternalAPI(prompt)
	if err != nil {
		return nil, nil, err
	}

	// Instead of parsing JSON, directly use the raw generated response
	generatedRecipe := &models.Recipe{
		Title:       "Generated Recipe",
		Description: generatedResponse,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return generatedRecipe, []*models.Recipe{}, nil
}

// Consolidate DeepSeek integration: delegate the call to integrations.GenerateRecipe
func callExternalAPI(prompt string) (string, error) {
	return integrations.GenerateRecipe(prompt, make(map[string]interface{}))
}
