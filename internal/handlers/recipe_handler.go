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

// @Summary List all recipes
// @Description Get a list of all recipes
// @Tags recipes
// @Accept json
// @Produce json
// @Success 200 {array} models.Recipe
// @Failure 500 {object} ErrorResponse
// @Router /v1/recipes [get]
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

// @Summary Get a recipe by ID
// @Description Get a recipe by its unique ID
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 200 {object} models.Recipe
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/recipes/{id} [get]
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

// @Summary Create a new recipe
// @Description Create a new recipe with the provided details
// @Tags recipes
// @Accept json
// @Produce json
// @Param recipe body models.Recipe true "Recipe object"
// @Success 201 {object} models.Recipe
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/recipes [post]
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

// @Summary Update a recipe
// @Description Update an existing recipe with new details
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Param recipe body models.Recipe true "Updated recipe object"
// @Success 200 {object} models.Recipe
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/recipes/{id} [put]
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

// @Summary Delete a recipe
// @Description Delete a recipe by its ID
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/recipes/{id} [delete]
func (h *RecipeHandler) DeleteRecipe(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.DeleteRecipe(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Resolve a recipe
// @Description Resolve a recipe based on a query and attributes
// @Tags recipes
// @Accept json
// @Produce json
// @Param request body ResolveRecipeRequest true "Resolve recipe request"
// @Success 200 {object} ResolveRecipeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/recipes/resolve [post]
func (h *RecipeHandler) ResolveRecipe(c *gin.Context) {
	var req ResolveRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resolved, similar, err := h.Service.ResolveRecipe(req.Query, req.Attributes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ResolveRecipeResponse{
		Resolved: resolved,
		Similar:  similar,
	})
}

// ResolveRecipeRequest represents the request body for recipe resolution
type ResolveRecipeRequest struct {
	Query      string                 `json:"query" binding:"required"`
	Attributes map[string]interface{} `json:"attributes"`
}

// ResolveRecipeResponse represents the response for recipe resolution
type ResolveRecipeResponse struct {
	Resolved *models.Recipe   `json:"resolved"`
	Similar  []*models.Recipe `json:"similar"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}
