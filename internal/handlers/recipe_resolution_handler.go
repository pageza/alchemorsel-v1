package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/services"
)

// RecipeResolutionHandler handles requests related to recipe resolution.
type RecipeResolutionHandler struct {
	service services.RecipeService
}

// NewRecipeResolutionHandler creates a new instance of RecipeResolutionHandler.
func NewRecipeResolutionHandler(service services.RecipeService) *RecipeResolutionHandler {
	return &RecipeResolutionHandler{service: service}
}

// ResolveRecipe processes recipe resolution requests.
func (h *RecipeResolutionHandler) ResolveRecipe(c *gin.Context) {
	var req dtos.RecipeResolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	attributes := map[string]interface{}{
		"ingredients":               req.Ingredients,
		"steps":                     req.Steps,
		"cuisines":                  req.Cuisines,
		"diets":                     req.Diets,
		"allergy_disclaimer":        req.AllergyDisclaimer,
		"modification_instructions": req.ModificationInstructions,
	}

	// Prepend a prefix prompt instructing the model on expected behavior and response format
	const promptPrefix = "Instruction: Provide a recipe recommendation that satisfies the user's request. The response should be in JSON format with keys 'candidate' and 'alternatives'."
	var userInput string
	if req.Query != "" {
		userInput = req.Query
	} else {
		userInput = req.Title
	}
	finalQuery := promptPrefix + " " + userInput

	candidate, alternatives, err := h.service.ResolveRecipe(c.Request.Context(), finalQuery, attributes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"candidate":    candidate,
		"alternatives": alternatives,
	})
}
