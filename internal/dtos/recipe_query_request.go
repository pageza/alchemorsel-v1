package dtos

// RecipeQueryRequest defines the payload for initiating a recipe resolution query.
// It includes a natural language query along with prompt instructions that guide the model,
// and the expected response format.

type RecipeQueryRequest struct {
	Query                  string `json:"query" binding:"required"`
	PromptInstructions     string `json:"prompt_instructions" binding:"required"`
	ExpectedResponseFormat string `json:"expected_response_format" binding:"required"`
}
