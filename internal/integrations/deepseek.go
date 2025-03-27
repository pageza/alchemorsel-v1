package integrations

import (
	"time"

	"github.com/pageza/alchemorsel-v1/internal/utils"
)

// GenerateRecipe generates a recipe using the Deepseek API with retry logic.
func GenerateRecipe(query string, attributes map[string]interface{}) (string, error) {
	var recipe string
	err := utils.Retry(3, 2*time.Second, func() error {
		// Dummy implementation: simulate calling the Deepseek API to generate a recipe.
		recipe = "Generated recipe for " + query
		return nil
	})
	return recipe, err
}
