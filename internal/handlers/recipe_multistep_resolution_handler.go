package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/parsers"
	"github.com/pageza/alchemorsel-v1/internal/services"
)

// RecipeMultistepResolutionHandler handles the multi-step recipe resolution process.
type RecipeMultistepResolutionHandler struct {
	service services.RecipeResolutionService
}

// NewRecipeMultistepResolutionHandler creates a new instance of RecipeMultistepResolutionHandler.
func NewRecipeMultistepResolutionHandler(service services.RecipeResolutionService) *RecipeMultistepResolutionHandler {
	return &RecipeMultistepResolutionHandler{
		service: service,
	}
}

// QueryRecipe handles the initial natural language query, incorporating user directives and profile details.
// It first checks the database for exact or close matches using a structured query built from the parsed natural language input.
// If no acceptable match is found, it builds a composite prompt and calls the external model to generate a recipe recommendation.
func (h *RecipeMultistepResolutionHandler) QueryRecipe(c *gin.Context) {
	var req dtos.RecipeQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Parse the user's freeform query into structured parameters using the parser
	parsedQuery, err := parsers.ParseRecipeQuery(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing query: " + err.Error()})
		return
	}

	// Retrieve user profile information (simulate for now)
	profileData := map[string]interface{}{
		"allergens":            []string{"peanuts"},
		"dietary_restrictions": "vegetarian",
	}

	ctx := c.Request.Context()

	// Instead of two separate database calls, retrieve close matches first
	closeMatches, err := h.service.FindCloseMatches(ctx, parsedQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while searching for matches: " + err.Error()})
		return
	}
	if len(closeMatches) > 0 {
		// Helper function to decide if a close match qualifies as an exact match.
		// For demonstration, assume an exact match if the recipe string exactly equals the original query.
		isExactMatch := func(recipe string, pq *parsers.ParsedQuery) bool {
			// TODO: Implement proper exact matching logic based on parsed query details.
			return recipe == req.Query
		}

		if isExactMatch(closeMatches[0], parsedQuery) {
			exactMatch := closeMatches[0]
			alternatives := closeMatches[1:]
			c.JSON(http.StatusOK, gin.H{
				"match_type":   "exact",
				"recipe":       exactMatch,
				"alternatives": alternatives,
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"match_type": "close",
				"recipes":    closeMatches,
			})
			return
		}
	}

	if len(closeMatches) == 0 {
		// No matches found, so build a composite prompt and call the external model
		compositePrompt, err := h.service.BuildCompositePrompt(req.Query, req.PromptInstructions, req.ExpectedResponseFormat, profileData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while building composite prompt: " + err.Error()})
			return
		}

		candidate, alternatives, err := h.service.ResolveRecipeByModel(ctx, compositePrompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while resolving recipe by model: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"match_type":   "generated",
			"candidate":    candidate,
			"alternatives": alternatives,
		})
		return
	}
}

// ModifyRecipe handles iterative modifications based on the user's feedback.
// It receives a structured response from the model alongside modification instructions and sends the request back to the model
// for further refinement until the recipe is approved by the user.
func (h *RecipeMultistepResolutionHandler) ModifyRecipe(c *gin.Context) {
	var req dtos.RecipeModificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Process the candidate recipe and alternatives from the model and apply any user-provided modification instructions.

	// TODO: Optionally, re-invoke the external model with the modifications to generate an updated recipe.

	c.JSON(http.StatusOK, gin.H{
		"message": "ModifyRecipe not implemented yet",
	})
}
