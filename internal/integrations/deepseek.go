package integrations

import (
	"errors"
	"os"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/utils"
)

// GenerateRecipe generates a recipe using the Deepseek API with retry logic.
func GenerateRecipe(query string, attributes map[string]interface{}) (string, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return "", errors.New("DEEPSEEK_API_KEY is not set")
	}

	// Normally, use apiKey with an HTTP client to call the Deepseek API.
	var recipe string
	err := utils.Retry(3, 2*time.Second, func() error {
		// Dummy implementation: simulate calling the Deepseek API using the apiKey to generate a recipe.
		recipe = "Generated recipe for " + query
		return nil
	})
	return recipe, err
}
