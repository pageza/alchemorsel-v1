package dtos

// RecipeQueryRequest defines the payload for initiating a recipe resolution query.
// It includes a natural language query along with prompt instructions that guide the model,
// and the expected response format.

type RecipeQueryRequest struct {
	Query                  string `json:"query" binding:"required"`
	PromptInstructions     string `json:"promptInstructions" binding:"required"`
	ExpectedResponseFormat string `json:"expectedResponseFormat" binding:"required"`
}
