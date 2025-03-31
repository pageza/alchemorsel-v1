package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/services"
)

// ResolveRecipe handles POST /v1/recipes/resolve
func ResolveRecipe(c *gin.Context) {
	var resolutionReq dtos.RecipeResolutionRequest
	if err := c.ShouldBindJSON(&resolutionReq); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// Construct a map of extra attributes from the request.
	attributes := map[string]interface{}{
		"ingredients":               resolutionReq.Ingredients,
		"steps":                     resolutionReq.Steps,
		"cuisines":                  resolutionReq.Cuisines,
		"diets":                     resolutionReq.Diets,
		"allergy_disclaimer":        resolutionReq.AllergyDisclaimer,
		"modification_instructions": resolutionReq.ModificationInstructions,
	}

	candidate, alternatives, err := services.ResolveRecipe(resolutionReq.Title, attributes)
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
