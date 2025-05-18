package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/utils"
)

// OpenAIEmbeddingResponse represents the response from the OpenAI embeddings API
type OpenAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// readSecretFile reads a secret from the Docker secrets directory
func readSecretFile(secretName string) (string, error) {
	log.Printf("Attempting to read secret: %s", secretName)

	// Try Docker secrets path first
	dockerPath := "/run/secrets/" + secretName
	log.Printf("Checking Docker secrets path: %s", dockerPath)
	data, err := os.ReadFile(dockerPath)
	if err != nil {
		log.Printf("Could not read from Docker secrets path: %v", err)

		// Fall back to local secrets directory
		localPath := "./secrets/" + secretName + ".txt"
		log.Printf("Falling back to local path: %s", localPath)
		data, err = os.ReadFile(localPath)
		if err != nil {
			log.Printf("Failed to read secret from local path: %v", err)
			return "", fmt.Errorf("failed to read secret %s: %v", secretName, err)
		}
		log.Printf("Successfully read secret from local path")
	} else {
		log.Printf("Successfully read secret from Docker secrets path")
	}

	// Remove any trailing newlines but preserve the rest of the content exactly
	return strings.TrimRight(string(data), "\n\r"), nil
}

// GenerateEmbedding obtains a numeric embedding for a recipe using the OpenAI API.
func GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	// In test mode, bypass API key check and return a dummy embedding.
	if os.Getenv("TEST_MODE") != "" {
		return []float64{0.1, 0.2, 0.3, 0.4, 0.5}, nil
	}

	// Get API key from Docker secrets
	log.Printf("Retrieving OpenAI API key")
	apiKey, err := readSecretFile("openai_api_key")
	if err != nil {
		log.Printf("Failed to read OpenAI API key: %v", err)
		return nil, fmt.Errorf("failed to read OpenAI API key: %v", err)
	}
	log.Printf("Successfully retrieved OpenAI API key (length: %d)", len(apiKey))

	// Get API URL from Docker secrets
	log.Printf("Retrieving OpenAI API URL")
	apiURL, err := readSecretFile("openai_api_url")
	if err != nil {
		log.Printf("Failed to read OpenAI API URL: %v", err)
		return nil, fmt.Errorf("failed to read OpenAI API URL: %v", err)
	}

	// Clean up the API URL
	apiURL = strings.TrimSpace(apiURL)
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1"
	}
	log.Printf("Using OpenAI API URL: %s", apiURL)

	payload := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var embedding []float64
	err = utils.Retry(3, 2*time.Second, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL+"/embeddings", bytes.NewReader(payloadBytes))
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(body))
		}

		var result OpenAIEmbeddingResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode response: %v", err)
		}

		if len(result.Data) == 0 {
			return errors.New("no embeddings returned from API")
		}

		embedding = result.Data[0].Embedding
		return nil
	})

	return embedding, err
}
