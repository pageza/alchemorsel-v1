package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
)

// RecipeHandler handles HTTP requests for recipe generation
type RecipeHandler struct{}

// NewRecipeHandler creates a new instance of RecipeHandler
func NewRecipeHandler() *RecipeHandler {
	return &RecipeHandler{}
}

// DeepSeekResponse represents the response from the DeepSeek API
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Add the Recipe struct after the DeepSeekResponse struct
type Recipe struct {
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Servings         int           `json:"servings"`
	PrepTimeMinutes  int           `json:"prep_time_minutes"`
	CookTimeMinutes  int           `json:"cook_time_minutes"`
	TotalTimeMinutes int           `json:"total_time_minutes"`
	Ingredients      []Ingredient  `json:"ingredients"`
	Instructions     []Instruction `json:"instructions"`
	Nutrition        Nutrition     `json:"nutrition"`
	Tags             []string      `json:"tags"`
	Difficulty       string        `json:"difficulty"`
}

type Ingredient struct {
	Item   string      `json:"item"`
	Amount json.Number `json:"amount"`
	Unit   string      `json:"unit"`
}

type Instruction struct {
	Step        int    `json:"step"`
	Description string `json:"description"`
}

type Nutrition struct {
	Calories int    `json:"calories"`
	Protein  string `json:"protein"`
	Carbs    string `json:"carbs"`
	Fat      string `json:"fat"`
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

	return string(bytes.TrimSpace(data)), nil
}

// GenerateRecipe handles the request to generate a recipe
func (h *RecipeHandler) GenerateRecipe(c *gin.Context) {
	log.Printf("Received recipe generation request")

	var req struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body: query field is required",
		})
		return
	}

	log.Printf("Processing query: %s", req.Query)

	// Build the DeepSeek payload
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You are a recipe generation assistant that always responds in JSON format. Format your responses as valid JSON objects with the following structure:
{
  "title": "Recipe Name",
  "description": "Brief description",
  "servings": 4,
  "prep_time_minutes": 15,
  "cook_time_minutes": 30,
  "total_time_minutes": 45,
  "ingredients": [
    {
      "item": "ingredient name",
      "amount": 2,
      "unit": "cups"
    }
  ],
  "instructions": [
    {
      "step": 1,
      "description": "Step description"
    }
  ],
  "nutrition": {
    "calories": 300,
    "protein": "10g",
    "carbs": "45g",
    "fat": "12g"
  },
  "tags": ["vegetarian", "easy", "quick"],
  "difficulty": "easy"
}`,
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Generate a recipe in JSON format for: %s", req.Query),
			},
		},
		"temperature": 0.7,
		"max_tokens":  2048,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal payload: %v", err),
		})
		return
	}

	// Get API key from Docker secrets
	log.Printf("Retrieving DeepSeek API key")
	apiKey, err := readSecretFile("deepseek_api_key")
	if err != nil {
		log.Printf("Failed to read DeepSeek API key: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to read DeepSeek API key: %v", err),
		})
		return
	}
	log.Printf("Successfully retrieved DeepSeek API key (length: %d)", len(apiKey))

	// Get API URL from Docker secrets
	log.Printf("Retrieving DeepSeek API URL")
	apiURL, err := readSecretFile("deepseek_api_url")
	if err != nil {
		log.Printf("Failed to read DeepSeek API URL: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to read DeepSeek API URL: %v", err),
		})
		return
	}

	// Clean up the API URL
	apiURL = strings.TrimSpace(apiURL)
	log.Printf("Successfully retrieved DeepSeek API URL: %s", apiURL)

	// Ensure the API URL ends with /chat/completions
	if !bytes.HasSuffix([]byte(apiURL), []byte("/chat/completions")) {
		apiURL = apiURL + "/chat/completions"
		log.Printf("Appended /chat/completions to API URL: %s", apiURL)
	}

	// Create HTTP client with increased timeout (90 seconds instead of 30)
	client := &http.Client{
		Timeout: 90 * time.Second,
	}

	// Create context with timeout for better cancellation handling
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	// Create request with context
	log.Printf("Creating HTTP request to: %s", apiURL)
	req2, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to create request: %v", err),
		})
		return
	}

	// Set headers
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req2.Header.Set("Accept", "application/json")
	log.Printf("Set request headers")

	// Make request
	log.Printf("Sending request to DeepSeek API")
	resp, err := client.Do(req2)
	if err != nil {
		log.Printf("Failed to make request to DeepSeek API: %v", err)

		// Handle timeout errors specifically
		if os.IsTimeout(err) || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			c.JSON(http.StatusGatewayTimeout, dtos.ErrorResponse{
				Code:    "TIMEOUT_ERROR",
				Message: "The request to DeepSeek API timed out. Please try again later or with a simpler query.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to make request to DeepSeek API: %v", err),
		})
		return
	}
	defer resp.Body.Close()
	log.Printf("Received response with status code: %d", resp.StatusCode)

	// Read response body with a timeout context
	bodyChannel := make(chan []byte, 1)
	errChannel := make(chan error, 1)

	go func() {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errChannel <- err
			return
		}
		bodyChannel <- body
	}()

	// Wait for either the body to be read or a timeout
	select {
	case body := <-bodyChannel:
		// Process the response body

		// Check response status code
		if resp.StatusCode != http.StatusOK {
			log.Printf("DeepSeek API returned error status code %d: %s", resp.StatusCode, string(body))
			c.JSON(resp.StatusCode, dtos.ErrorResponse{
				Code:    "DEEPSEEK_API_ERROR",
				Message: fmt.Sprintf("DeepSeek API returned error: %s", string(body)),
			})
			return
		}

		// Parse response
		log.Printf("Parsing DeepSeek API response")
		var deepseekResp DeepSeekResponse
		if err := json.Unmarshal(body, &deepseekResp); err != nil {
			log.Printf("Failed to parse DeepSeek response: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: fmt.Sprintf("Failed to parse DeepSeek response: %v", err),
			})
			return
		}

		// Check if there are any choices in the response
		if len(deepseekResp.Choices) == 0 {
			log.Printf("DeepSeek API returned no choices")
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "DeepSeek API returned no choices",
			})
			return
		}

		// Parse the recipe JSON string into our Recipe struct
		var recipe Recipe
		if err := json.Unmarshal([]byte(deepseekResp.Choices[0].Message.Content), &recipe); err != nil {
			log.Printf("Failed to parse recipe JSON: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: fmt.Sprintf("Failed to parse recipe JSON: %v", err),
			})
			return
		}

		log.Printf("Successfully generated recipe, returning response")
		// Return the parsed recipe object
		c.JSON(http.StatusOK, gin.H{
			"recipe": recipe,
			"usage": gin.H{
				"prompt_tokens":     deepseekResp.Usage.PromptTokens,
				"completion_tokens": deepseekResp.Usage.CompletionTokens,
				"total_tokens":      deepseekResp.Usage.TotalTokens,
			},
		})

	case err := <-errChannel:
		log.Printf("Failed to read response body: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to read response body: %v", err),
		})
		return

	case <-ctx.Done():
		log.Printf("Context deadline exceeded while reading response body")
		c.JSON(http.StatusGatewayTimeout, dtos.ErrorResponse{
			Code:    "TIMEOUT_ERROR",
			Message: "Timeout while reading response from DeepSeek API. Please try again later or with a simpler query.",
		})
		return
	}
}
