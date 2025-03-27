package integrations

import (
	"time"

	"github.com/pageza/alchemorsel-v1/internal/utils"
)

// GenerateEmbedding obtains a numeric embedding for a recipe using the OpenAI API.
func GenerateEmbedding(recipe string) ([]float64, error) {
	var embedding []float64
	err := utils.Retry(3, 2*time.Second, func() error {
		// Dummy implementation: simulate a call to the OpenAI API to obtain an embedding.
		embedding = []float64{0.1, 0.2, 0.3, 0.4, 0.5}
		return nil
	})
	return embedding, err
}
