package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RecipeHandler handles recipe-related HTTP requests with dependency injection.
type RecipeHandler struct {
	Service services.RecipeService
}

// NewRecipeHandler creates a new RecipeHandler with the given service.
func NewRecipeHandler(service services.RecipeService) *RecipeHandler {
	return &RecipeHandler{Service: service}
}

// @Summary List all recipes
// @Description Get a list of all recipes
// @Tags recipes
// @Accept json
// @Produce json
// @Success 200 {object} dtos.RecipeListResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/recipes [get]
func (h *RecipeHandler) ListRecipes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	recipes, err := h.Service.ListRecipes(c.Request.Context(), page, limit, sort, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	// Convert to response DTOs
	var response dtos.RecipeListResponse
	response.Recipes = make([]dtos.RecipeResponse, len(recipes))
	for i, recipe := range recipes {
		response.Recipes[i] = *dtos.NewRecipeResponse(&recipe)
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
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Code: "NOT_FOUND", Message: "Recipe not found"})
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
// @Param recipe body dtos.RecipeRequest true "Recipe details"
// @Success 201 {object} dtos.RecipeResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Router /v1/recipes [post]
func (h *RecipeHandler) SaveRecipe(c *gin.Context) {
	var recipeReq dtos.RecipeRequest
	if err := c.ShouldBindJSON(&recipeReq); err != nil {
		logrus.WithError(err).Error("Failed to bind JSON request")
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid request body: " + err.Error()})
		return
	}

	// Collect validation errors
	var validationErrors []string
	if recipeReq.Title == "" {
		validationErrors = append(validationErrors, "Title is required")
	}
	if len(recipeReq.Ingredients) == 0 {
		validationErrors = append(validationErrors, "At least one ingredient is required")
	}
	if len(recipeReq.Steps) == 0 {
		validationErrors = append(validationErrors, "At least one step is required")
	}

	// Validate ingredients
	for i, ing := range recipeReq.Ingredients {
		if ing.Name == "" || ing.Amount == "" || ing.Unit == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid ingredient at index %d: name, amount, and unit are required", i))
		}
	}

	// Validate steps
	for i, step := range recipeReq.Steps {
		if step.Description == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid step at index %d: description is required", i))
		}
	}

	// If there are validation errors, return them all at once
	if len(validationErrors) > 0 {
		logrus.WithField("errors", validationErrors).Error("Validation failed")
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: strings.Join(validationErrors, "; "),
		})
		return
	}

	// Create recipe model
	recipe := &models.Recipe{
		Title:             recipeReq.Title,
		Description:       recipeReq.Description,
		NutritionalInfo:   recipeReq.NutritionalInfo,
		AllergyDisclaimer: recipeReq.AllergyDisclaimer,
		Difficulty:        recipeReq.Difficulty,
		PrepTime:          recipeReq.PrepTime,
		CookTime:          recipeReq.CookTime,
		Servings:          recipeReq.Servings,
		Approved:          recipeReq.Approved,
	}

	// Convert ingredients
	ingredients := make([]models.Ingredient, len(recipeReq.Ingredients))
	for i, ing := range recipeReq.Ingredients {
		ingredients[i] = models.Ingredient{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	if err := recipe.SetIngredients(ingredients); err != nil {
		logrus.WithError(err).Error("Failed to set ingredients")
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Failed to set ingredients: " + err.Error()})
		return
	}

	// Convert steps
	steps := make([]models.Step, len(recipeReq.Steps))
	for i, step := range recipeReq.Steps {
		steps[i] = models.Step{
			Order:       step.Order,
			Description: step.Description,
		}
	}
	if err := recipe.SetSteps(steps); err != nil {
		logrus.WithError(err).Error("Failed to set steps")
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Failed to set steps: " + err.Error()})
		return
	}

	// Convert string arrays to models
	for _, name := range recipeReq.Cuisines {
		recipe.Cuisines = append(recipe.Cuisines, models.Cuisine{Name: name})
	}
	for _, name := range recipeReq.Diets {
		recipe.Diets = append(recipe.Diets, models.Diet{Name: name})
	}
	for _, name := range recipeReq.Appliances {
		recipe.Appliances = append(recipe.Appliances, models.Appliance{Name: name})
	}
	for _, name := range recipeReq.Tags {
		recipe.Tags = append(recipe.Tags, models.Tag{Name: name})
	}

	// Save recipe
	if err := h.Service.SaveRecipe(c.Request.Context(), recipe); err != nil {
		logrus.WithError(err).Error("Failed to save recipe")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to save recipe: " + err.Error()})
		return
	}

	// Return created recipe
	response := dtos.NewRecipeResponse(recipe)
	c.JSON(http.StatusCreated, response)
}

// @Summary Update a recipe
// @Description Update an existing recipe with the provided details
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Param recipe body dtos.RecipeRequest true "Recipe details"
// @Success 200 {object} dtos.RecipeResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /v1/recipes/{id} [put]
func (h *RecipeHandler) UpdateRecipe(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Recipe ID is required"})
		return
	}

	var recipeReq dtos.RecipeRequest
	if err := c.ShouldBindJSON(&recipeReq); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: err.Error()})
		return
	}

	// Get existing recipe
	recipe, err := h.Service.GetRecipe(c.Request.Context(), id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Code: "NOT_FOUND", Message: "Recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	// Update recipe fields
	recipe.Title = recipeReq.Title
	recipe.Description = recipeReq.Description
	recipe.NutritionalInfo = recipeReq.NutritionalInfo
	recipe.AllergyDisclaimer = recipeReq.AllergyDisclaimer
	recipe.Difficulty = recipeReq.Difficulty
	recipe.PrepTime = recipeReq.PrepTime
	recipe.CookTime = recipeReq.CookTime
	recipe.Servings = recipeReq.Servings
	recipe.Approved = recipeReq.Approved

	// Convert and validate ingredients
	ingredients := make([]models.Ingredient, len(recipeReq.Ingredients))
	for i, ing := range recipeReq.Ingredients {
		if ing.Name == "" || ing.Amount == "" || ing.Unit == "" {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid ingredient: name, amount, and unit are required"})
			return
		}
		ingredients[i] = models.Ingredient{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	if err := recipe.SetIngredients(ingredients); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Failed to set ingredients: " + err.Error()})
		return
	}

	// Convert and validate steps
	steps := make([]models.Step, len(recipeReq.Steps))
	for i, step := range recipeReq.Steps {
		if step.Description == "" {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid step: description is required"})
			return
		}
		steps[i] = models.Step{
			Order:       step.Order,
			Description: step.Description,
		}
	}
	if err := recipe.SetSteps(steps); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Failed to set steps: " + err.Error()})
		return
	}

	// Convert string arrays to models
	recipe.Cuisines = nil
	recipe.Diets = nil
	recipe.Appliances = nil
	recipe.Tags = nil

	for _, name := range recipeReq.Cuisines {
		recipe.Cuisines = append(recipe.Cuisines, models.Cuisine{Name: name})
	}
	for _, name := range recipeReq.Diets {
		recipe.Diets = append(recipe.Diets, models.Diet{Name: name})
	}
	for _, name := range recipeReq.Appliances {
		recipe.Appliances = append(recipe.Appliances, models.Appliance{Name: name})
	}
	for _, name := range recipeReq.Tags {
		recipe.Tags = append(recipe.Tags, models.Tag{Name: name})
	}

	// Update recipe
	if err := h.Service.UpdateRecipe(c.Request.Context(), recipe); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Code: "NOT_FOUND", Message: "Recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	// Convert to response DTO
	response := dtos.NewRecipeResponse(recipe)
	c.JSON(http.StatusOK, response)
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
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Code: "NOT_FOUND", Message: "Recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to delete recipe: " + err.Error()})
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
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: err.Error()})
		return
	}

	resolved, similar, err := h.Service.ResolveRecipe(c.Request.Context(), req.Query, req.Attributes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to resolve recipe: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, ResolveRecipeResponse{
		Resolved: resolved,
		Similar:  similar,
	})
}

// RateRecipe handles rating a recipe.
// @Summary Rate a recipe
// @Description Add a rating to a recipe
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Param rating body float64 true "Rating value"
// @Success 200 {object} dtos.RecipeResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /v1/recipes/{id}/rate [post]
func (h *RecipeHandler) RateRecipe(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Recipe ID is required"})
		return
	}

	var rating float64
	if err := c.ShouldBindJSON(&rating); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid rating value"})
		return
	}

	if rating < 0 || rating > 5 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Rating must be between 0 and 5"})
		return
	}

	if err := h.Service.RateRecipe(c.Request.Context(), id, rating); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Code: "NOT_FOUND", Message: "Recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to rate recipe: " + err.Error()})
		return
	}

	// Get updated recipe
	recipe, err := h.Service.GetRecipe(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	response := dtos.NewRecipeResponse(recipe)
	c.JSON(http.StatusOK, response)
}

// GetRecipeRatings handles retrieving ratings for a recipe.
// @Summary Get recipe ratings
// @Description Get all ratings for a recipe
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 200 {array} float64
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /v1/recipes/{id}/ratings [get]
func (h *RecipeHandler) GetRecipeRatings(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Code: "BAD_REQUEST", Message: "Recipe ID is required"})
		return
	}

	ratings, err := h.Service.GetRecipeRatings(c.Request.Context(), id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Code: "NOT_FOUND", Message: "Recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ratings)
}

// SearchRecipes handles searching for recipes.
// @Summary Search recipes
// @Description Search for recipes based on query parameters
// @Tags recipes
// @Accept json
// @Produce json
// @Param q query string false "Search query"
// @Param tags query []string false "Filter by tags"
// @Param difficulty query string false "Filter by difficulty"
// @Success 200 {object} dtos.RecipeListResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Router /v1/recipes/search [get]
func (h *RecipeHandler) SearchRecipes(c *gin.Context) {
	query := c.Query("q")
	tags := c.QueryArray("tags")
	difficulty := c.Query("difficulty")

	recipes, err := h.Service.SearchRecipes(c.Request.Context(), query, tags, difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	// Convert to response DTOs
	var response dtos.RecipeListResponse
	response.Recipes = make([]dtos.RecipeResponse, len(recipes))
	for i, recipe := range recipes {
		response.Recipes[i] = *dtos.NewRecipeResponse(&recipe)
	}

	c.JSON(http.StatusOK, response)
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
