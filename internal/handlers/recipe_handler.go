package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/integrations"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"go.uber.org/zap"
)

// ListRecipes handles GET /v1/recipes
func ListRecipes(c *gin.Context) {
	// TODO: Retrieve list of recipes, applying any required filtering.
	c.JSON(http.StatusOK, gin.H{"message": "ListRecipes endpoint - TODO: implement logic"})
}

// GetRecipe handles GET /v1/recipes/:id
func GetRecipe(c *gin.Context) {
	// TODO: Retrieve specific recipe details by ID.
	c.JSON(http.StatusOK, gin.H{"message": "GetRecipe endpoint - TODO: implement logic"})
}

// SaveRecipe handles POST /v1/recipes
func SaveRecipe(c *gin.Context) {
	var req dtos.RecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sanitize input: trim spaces from title, ingredients, and steps.
	req.Title = strings.TrimSpace(req.Title)
	for i, ingredient := range req.Ingredients {
		req.Ingredients[i] = strings.TrimSpace(ingredient)
	}
	for i, step := range req.Steps {
		req.Steps[i] = strings.TrimSpace(step)
	}

	// Check if the candidate recipe has been approved.
	if !req.Approved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe not approved by user"})
		return
	}

	// Log that an approved recipe is being processed.
	zap.S().Infow("User-approved recipe received. Proceeding with save", "title", req.Title)

	// Convert slices to JSON for persistence.
	ingredientsJSON, err := json.Marshal(req.Ingredients)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal ingredients"})
		return
	}
	stepsJSON, err := json.Marshal(req.Steps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal steps"})
		return
	}

	// Build a Recipe model from the request.
	recipe := models.Recipe{
		Title:             req.Title,
		Ingredients:       ingredientsJSON,
		Steps:             stepsJSON,
		NutritionalInfo:   req.NutritionalInfo,
		AllergyDisclaimer: req.AllergyDisclaimer,
		// For this MVP, we are not resolving Appliances, Cuisines, or Diets.
	}

	// Generate a text representation for embedding generation.
	recipeText := req.Title + " " + strings.Join(req.Ingredients, " ") + " " + strings.Join(req.Steps, " ")
	embedding, err := integrations.GenerateEmbedding(recipeText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate embedding: " + err.Error()})
		return
	}
	recipe.Embedding = embedding

	// Save the recipe via the service.
	if err := services.SaveRecipe(&recipe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the saved recipe as a RecipeResponse DTO.
	resp := dtos.NewRecipeResponse(&recipe)
	c.JSON(http.StatusCreated, resp)
}

// UpdateRecipe handles PUT /v1/recipes/:id
func UpdateRecipe(c *gin.Context) {
	// TODO: Parse request and update an existing recipe.
	c.JSON(http.StatusOK, gin.H{"message": "UpdateRecipe endpoint - TODO: implement logic"})
}

// DeleteRecipe handles DELETE /v1/recipes/:id
func DeleteRecipe(c *gin.Context) {
	// TODO: Delete recipe by ID.
	c.JSON(http.StatusOK, gin.H{"message": "DeleteRecipe endpoint - TODO: implement logic"})
}
