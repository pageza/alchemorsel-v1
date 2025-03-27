package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/integrations"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"go.uber.org/zap"
)

// RecipeHandler handles recipe-related HTTP requests with dependency injection.
type RecipeHandler struct {
	Service services.RecipeServiceInterface
}

// NewRecipeHandler creates a new RecipeHandler with the given service.
func NewRecipeHandler(service services.RecipeServiceInterface) *RecipeHandler {
	return &RecipeHandler{Service: service}
}

// ListRecipes handles GET /v1/recipes
func (h *RecipeHandler) ListRecipes(c *gin.Context) {
	recipes, err := h.Service.ListRecipes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]dtos.RecipeResponse, len(recipes))
	for i, recipe := range recipes {
		response[i] = dtos.NewRecipeResponse(recipe)
	}
	c.JSON(http.StatusOK, response)
}

// GetRecipe handles GET /v1/recipes/:id
func (h *RecipeHandler) GetRecipe(c *gin.Context) {
	id := c.Param("id")
	recipe, err := h.Service.GetRecipe(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	response := dtos.NewRecipeResponse(recipe)
	c.JSON(http.StatusOK, response)
}

// SaveRecipe handles POST /v1/recipes
func (h *RecipeHandler) SaveRecipe(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sanitize input: trim spaces from title
	recipe.Title = strings.TrimSpace(recipe.Title)
	if recipe.Title == "" {
		recipe.Title = "Integration Created Recipe"
	}

	// Get ingredients and steps as strings
	ingredients, err := recipe.GetIngredients()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingredients format"})
		return
	}
	steps, err := recipe.GetSteps()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid steps format"})
		return
	}

	// Sanitize ingredients and steps
	for i, ingredient := range ingredients {
		ingredients[i] = strings.TrimSpace(ingredient)
	}
	for i, step := range steps {
		steps[i] = strings.TrimSpace(step)
	}

	// Set sanitized ingredients and steps back
	if err := recipe.SetIngredients(ingredients); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set ingredients"})
		return
	}
	if err := recipe.SetSteps(steps); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set steps"})
		return
	}

	// Check if the candidate recipe has been approved.
	if !recipe.Approved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe not approved by user"})
		return
	}

	// Log that an approved recipe is being processed.
	zap.S().Infow("User-approved recipe received. Proceeding with save", "title", recipe.Title)

	// Generate a text representation for embedding generation.
	recipeText := recipe.Title + " " + strings.Join(ingredients, " ") + " " + strings.Join(steps, " ")

	// Skip embedding generation in test mode
	if os.Getenv("TEST_MODE") != "true" {
		embedding, err := integrations.GenerateEmbedding(recipeText)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate embedding: " + err.Error()})
			return
		}
		recipe.Embedding = embedding
	}

	// Save the recipe via the service.
	if err := h.Service.SaveRecipe(c.Request.Context(), &recipe); err != nil {
		zap.S().Errorw("SaveRecipe service error", "error", err, "recipeID", recipe.ID, "title", recipe.Title)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create the response DTO
	resp := dtos.NewRecipeResponse(&recipe)
	c.JSON(http.StatusCreated, resp)
}

// UpdateRecipe handles PUT /v1/recipes/:id
func (h *RecipeHandler) UpdateRecipe(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sanitize input
	recipe.Title = strings.TrimSpace(recipe.Title)

	// Get ingredients and steps as strings
	ingredients, err := recipe.GetIngredients()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingredients format"})
		return
	}
	steps, err := recipe.GetSteps()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid steps format"})
		return
	}

	// Sanitize ingredients and steps
	for i, ingredient := range ingredients {
		ingredients[i] = strings.TrimSpace(ingredient)
	}
	for i, step := range steps {
		steps[i] = strings.TrimSpace(step)
	}

	// Set sanitized ingredients and steps back
	if err := recipe.SetIngredients(ingredients); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set ingredients"})
		return
	}
	if err := recipe.SetSteps(steps); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set steps"})
		return
	}

	if err := h.Service.UpdateRecipe(c.Request.Context(), &recipe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := dtos.NewRecipeResponse(&recipe)
	c.JSON(http.StatusOK, resp)
}

// DeleteRecipe handles DELETE /v1/recipes/:id
func (h *RecipeHandler) DeleteRecipe(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.DeleteRecipe(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ResolveRecipe handles POST /v1/recipes/resolve
func (h *RecipeHandler) ResolveRecipe(c *gin.Context) {
	var payload struct {
		Query      string                 `json:"query"`
		Attributes map[string]interface{} `json:"attributes"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service-level ResolveRecipe
	resolved, similar, err := h.Service.ResolveRecipe(payload.Query, payload.Attributes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the resolved recipe and similar recipes.
	c.JSON(http.StatusOK, gin.H{
		"resolved": resolved,
		"similar":  similar,
	})
}
