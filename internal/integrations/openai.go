package integrations

import (
	"errors"
	"os"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/utils"
)

// GenerateEmbedding obtains a numeric embedding for a recipe using the OpenAI API.
func GenerateEmbedding(recipe string) ([]float64, error) {
	// In test mode, bypass API key check and return a dummy embedding.
	if os.Getenv("TEST_MODE") != "" {
		return []float64{0.1, 0.2, 0.3, 0.4, 0.5}, nil
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	// Normally, use apiKey with an HTTP client to call the OpenAI API.
	var embedding []float64
	err := utils.Retry(3, 2*time.Second, func() error {
		// Dummy implementation: simulate a call to the OpenAI API using the apiKey to obtain an embedding.
		embedding = []float64{0.1, 0.2, 0.3, 0.4, 0.5}
		return nil
	})
	return embedding, err
}
