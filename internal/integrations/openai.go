package integrations

// GenerateEmbedding obtains a numeric embedding for a recipe using the OpenAI API.
func GenerateEmbedding(recipe string) ([]float64, error) {
	// Dummy implementation: return a fixed embedding vector.
	// In a real implementation, call the OpenAI API to obtain an embedding.
	return []float64{0.1, 0.2, 0.3, 0.4, 0.5}, nil
}
