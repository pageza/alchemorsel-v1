package dtos

// RecipeModificationRequest defines the payload for modifying an existing recipe response from the model.
// It includes the candidate recipe, any alternative proposals, and a field for user-provided modification instructions,
// which can be sent back to the model to refine the recipe further.

type RecipeModificationRequest struct {
	Candidate                string   `json:"candidate" binding:"required"`
	Alternatives             []string `json:"alternatives,omitempty"`
	ModificationInstructions string   `json:"modification_instructions,omitempty"`
}
