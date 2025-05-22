package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/integrations"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"gorm.io/gorm"
)

// RecipeHandler handles recipe-related HTTP requests
type RecipeHandler struct {
	db             *gorm.DB
	recipeCache    *repositories.RecipeCache
	deepseekClient *repositories.DeepSeekClient
}

// NewRecipeHandler creates a new RecipeHandler
func NewRecipeHandler(db *gorm.DB, recipeCache *repositories.RecipeCache, deepseekClient *repositories.DeepSeekClient) *RecipeHandler {
	return &RecipeHandler{
		db:             db,
		recipeCache:    recipeCache,
		deepseekClient: deepseekClient,
	}
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
		PromptTokens      int `json:"prompt_tokens"`
		CompletionTokens  int `json:"completion_tokens"`
		TotalTokens       int `json:"total_tokens"`
		PromptCacheHits   int `json:"prompt_cache_hit_tokens"`
		PromptCacheMisses int `json:"prompt_cache_miss_tokens"`
	} `json:"usage"`
}

// Add the Recipe struct after the DeepSeekResponse struct
type Recipe struct {
	ID               string        `json:"id"`
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

	// Remove any trailing newlines but preserve the rest of the content exactly
	return strings.TrimRight(string(data), "\n\r"), nil
}

// GenerateRecipe handles recipe generation requests
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

		// Log detailed usage statistics
		log.Printf("DeepSeek API Usage Statistics:")
		log.Printf("  - Total Tokens: %d", deepseekResp.Usage.TotalTokens)
		log.Printf("  - Prompt Tokens: %d", deepseekResp.Usage.PromptTokens)
		log.Printf("  - Completion Tokens: %d", deepseekResp.Usage.CompletionTokens)
		log.Printf("  - Cache Hit Tokens: %d", deepseekResp.Usage.PromptCacheHits)
		log.Printf("  - Cache Miss Tokens: %d", deepseekResp.Usage.PromptCacheMisses)
		if deepseekResp.Usage.PromptCacheHits > 0 {
			hitPercentage := float64(deepseekResp.Usage.PromptCacheHits) / float64(deepseekResp.Usage.PromptTokens) * 100
			log.Printf("  - Cache Hit Rate: %.2f%%", hitPercentage)
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
		var recipe repositories.Recipe
		if err := json.Unmarshal([]byte(deepseekResp.Choices[0].Message.Content), &recipe); err != nil {
			log.Printf("Failed to parse recipe JSON: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: fmt.Sprintf("Failed to parse recipe JSON: %v", err),
			})
			return
		}

		log.Printf("Attempting to cache recipe in Redis")
		// Cache the recipe in Redis
		recipeID, err := h.recipeCache.CacheRecipe(c.Request.Context(), recipe)
		if err != nil {
			log.Printf("Failed to cache recipe in Redis: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: fmt.Sprintf("Failed to cache recipe: %v", err),
			})
			return
		}

		// Verify the recipe was cached by attempting to retrieve it
		log.Printf("Verifying recipe cache - Attempting to retrieve recipe with ID: %s", recipeID)
		cachedRecipe, err := h.recipeCache.GetRecipe(c.Request.Context(), recipeID)
		if err != nil {
			log.Printf("Warning: Could not verify recipe cache - retrieval failed: %v", err)
		} else {
			log.Printf("Successfully verified recipe cache - Recipe found in Redis with ID: %s", recipeID)
			log.Printf("Cache details - Original Title: %s, Modification Count: %d",
				cachedRecipe.Original.Title,
				cachedRecipe.ModificationCount)
		}

		log.Printf("Successfully generated and cached recipe with ID: %s", recipeID)
		// Set the ID in the recipe object before returning
		cachedRecipe.Current.ID = recipeID
		// Return the parsed recipe object with its ID
		c.JSON(http.StatusOK, gin.H{
			"recipe_id": recipeID,
			"recipe":    cachedRecipe.Current,
			"status":    "pending_approval",
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

// ApproveRecipe handles the approval of a generated recipe
func (h *RecipeHandler) ApproveRecipe(c *gin.Context) {
	// Get recipe ID from path
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Recipe ID is required",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User ID not found in context",
		})
		return
	}

	log.Printf("Approving recipe %s for user %s", recipeID, userID)

	// Get recipe from Redis
	cachedRecipe, err := h.recipeCache.GetRecipe(c.Request.Context(), recipeID)
	if err != nil {
		log.Printf("Failed to get recipe from Redis: %v", err)
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Recipe not found in cache",
		})
		return
	}

	// Prepare text for embedding
	embedText := fmt.Sprintf("%s\n%s\n%s\n%s",
		cachedRecipe.Current.Title,
		cachedRecipe.Current.Description,
		strings.Join(cachedRecipe.Current.Tags, ", "),
		strings.Join(func() []string {
			items := make([]string, len(cachedRecipe.Current.Ingredients))
			for i, ing := range cachedRecipe.Current.Ingredients {
				items[i] = ing.Item
			}
			return items
		}(), ", "),
	)

	log.Printf("Getting embeddings for recipe: %s", cachedRecipe.Current.Title)

	// Get embeddings from OpenAI
	embedding, err := integrations.GenerateEmbedding(
		c.Request.Context(),
		embedText,
	)
	if err != nil {
		log.Printf("Failed to get embeddings: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "EMBEDDING_ERROR",
			Message: "Failed to generate embeddings",
		})
		return
	}

	// Convert embeddings to []float32
	embeddingFloat32 := make([]float32, len(embedding))
	for i, v := range embedding {
		embeddingFloat32[i] = float32(v)
	}

	// Create the recipe model
	recipe := models.Recipe{
		ID:               recipeID,
		Title:            cachedRecipe.Current.Title,
		Description:      cachedRecipe.Current.Description,
		Servings:         cachedRecipe.Current.Servings,
		PrepTimeMinutes:  cachedRecipe.Current.PrepTimeMinutes,
		CookTimeMinutes:  cachedRecipe.Current.CookTimeMinutes,
		TotalTimeMinutes: cachedRecipe.Current.TotalTimeMinutes,
		Ingredients:      convertToModelIngredients(cachedRecipe.Current.Ingredients),
		Instructions:     convertToModelInstructions(cachedRecipe.Current.Instructions),
		Nutrition:        convertToModelNutrition(cachedRecipe.Current.Nutrition),
		Tags:             cachedRecipe.Current.Tags,
		Difficulty:       cachedRecipe.Current.Difficulty,
		Embedding:        embeddingFloat32,
		UserID:           userID.(string),
	}

	log.Printf("Saving recipe to database")

	// Convert all JSONB fields to JSON
	ingredientsJSON, err := json.Marshal(recipe.Ingredients)
	if err != nil {
		log.Printf("Failed to marshal ingredients: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to process ingredients",
		})
		return
	}

	instructionsJSON, err := json.Marshal(recipe.Instructions)
	if err != nil {
		log.Printf("Failed to marshal instructions: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to process instructions",
		})
		return
	}

	nutritionJSON, err := json.Marshal(recipe.Nutrition)
	if err != nil {
		log.Printf("Failed to marshal nutrition: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to process nutrition",
		})
		return
	}

	embeddingJSON, err := json.Marshal(embeddingFloat32)
	if err != nil {
		log.Printf("Failed to marshal embedding: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to process embedding",
		})
		return
	}

	// Convert tags to PostgreSQL array format
	tagsArray := "{" + strings.Join(recipe.Tags, ",") + "}"

	// Save to database using a map to ensure proper JSONB handling
	recipeMap := map[string]interface{}{
		"id":                 recipe.ID,
		"title":              recipe.Title,
		"description":        recipe.Description,
		"servings":           recipe.Servings,
		"prep_time_minutes":  recipe.PrepTimeMinutes,
		"cook_time_minutes":  recipe.CookTimeMinutes,
		"total_time_minutes": recipe.TotalTimeMinutes,
		"ingredients":        ingredientsJSON,
		"instructions":       instructionsJSON,
		"nutrition":          nutritionJSON,
		"tags":               tagsArray,
		"difficulty":         recipe.Difficulty,
		"embedding":          embeddingJSON,
		"user_id":            recipe.UserID,
	}

	// Save to database using the map
	if err := h.db.Model(&models.Recipe{}).Create(recipeMap).Error; err != nil {
		log.Printf("Failed to save recipe to database: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to save recipe",
		})
		return
	}

	log.Printf("Successfully saved recipe %s", recipe.ID)

	c.JSON(http.StatusOK, gin.H{
		"recipe": cachedRecipe.Current,
	})
}

// Convert repositories types to models types
func convertToModelIngredients(repoIngredients []repositories.Ingredient) []models.Ingredient {
	modelIngredients := make([]models.Ingredient, len(repoIngredients))
	for i, ing := range repoIngredients {
		modelIngredients[i] = models.Ingredient{
			Item:   ing.Item,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	return modelIngredients
}

func convertToModelInstructions(repoInstructions []repositories.Instruction) []models.Instruction {
	modelInstructions := make([]models.Instruction, len(repoInstructions))
	for i, inst := range repoInstructions {
		modelInstructions[i] = models.Instruction{
			Step:        inst.Step,
			Description: inst.Description,
		}
	}
	return modelInstructions
}

func convertToModelNutrition(repoNutrition repositories.Nutrition) models.Nutrition {
	return models.Nutrition{
		Calories: repoNutrition.Calories,
		Protein:  repoNutrition.Protein,
		Carbs:    repoNutrition.Carbs,
		Fat:      repoNutrition.Fat,
	}
}

// SearchRecipes handles recipe search requests using hybrid search
func (h *RecipeHandler) SearchRecipes(c *gin.Context) {
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

	log.Printf("Processing search query: %s", req.Query)

	// Get embeddings for the search query from OpenAI
	queryEmbedding, err := integrations.GenerateEmbedding(
		c.Request.Context(),
		req.Query,
	)
	if err != nil {
		log.Printf("Failed to get embeddings: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "EMBEDDING_ERROR",
			Message: "Failed to generate embeddings for search",
		})
		return
	}

	// Convert embeddings to []float32
	queryEmbeddingFloat32 := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		queryEmbeddingFloat32[i] = float32(v)
	}

	// First, try to find exact matches using text search
	var exactMatches []struct {
		ID               string    `gorm:"column:id"`
		Title            string    `gorm:"column:title"`
		Description      string    `gorm:"column:description"`
		Servings         int       `gorm:"column:servings"`
		PrepTimeMinutes  int       `gorm:"column:prep_time_minutes"`
		CookTimeMinutes  int       `gorm:"column:cook_time_minutes"`
		TotalTimeMinutes int       `gorm:"column:total_time_minutes"`
		Ingredients      []byte    `gorm:"column:ingredients"`
		Instructions     []byte    `gorm:"column:instructions"`
		Nutrition        []byte    `gorm:"column:nutrition"`
		Tags             string    `gorm:"column:tags"`
		Difficulty       string    `gorm:"column:difficulty"`
		CreatedAt        time.Time `gorm:"column:created_at"`
		UpdatedAt        time.Time `gorm:"column:updated_at"`
		UserID           string    `gorm:"column:user_id"`
	}

	// Log the search query
	log.Printf("Searching for recipes with query: %s", req.Query)

	// First try a direct text search
	searchQuery := "%" + req.Query + "%"
	log.Printf("Using search pattern: %s", searchQuery)

	// Try searching in title, description, and tags
	if err := h.db.Table("recipes").
		Select("id, title, description, servings, prep_time_minutes, cook_time_minutes, total_time_minutes, ingredients, instructions, nutrition, tags, difficulty, created_at, updated_at, user_id").
		Where("title ILIKE ? OR description ILIKE ? OR tags::text ILIKE ?",
			searchQuery, searchQuery, searchQuery).
		Find(&exactMatches).Error; err != nil {
		log.Printf("Failed to find exact matches: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to search recipes",
		})
		return
	}

	log.Printf("Found %d exact matches", len(exactMatches))
	for _, match := range exactMatches {
		log.Printf("Exact match found - ID: %s, Title: %s", match.ID, match.Title)
	}

	// Convert exact matches to repository types
	convertedExactMatches := make([]repositories.Recipe, 0, len(exactMatches))
	for _, recipe := range exactMatches {
		// Parse ingredients
		var ingredients []models.Ingredient
		if len(recipe.Ingredients) > 0 {
			if err := json.Unmarshal(recipe.Ingredients, &ingredients); err != nil {
				log.Printf("Failed to parse ingredients for recipe %s: %v", recipe.ID, err)
				continue
			}
		}

		// Parse instructions
		var instructions []models.Instruction
		if len(recipe.Instructions) > 0 {
			if err := json.Unmarshal(recipe.Instructions, &instructions); err != nil {
				log.Printf("Failed to parse instructions for recipe %s: %v", recipe.ID, err)
				continue
			}
		}

		// Parse nutrition
		var nutrition models.Nutrition
		if len(recipe.Nutrition) > 0 {
			if err := json.Unmarshal(recipe.Nutrition, &nutrition); err != nil {
				log.Printf("Failed to parse nutrition for recipe %s: %v", recipe.ID, err)
				continue
			}
		}

		// Parse tags from PostgreSQL array format
		var tags []string
		if recipe.Tags != "" {
			// Remove the curly braces and split by comma
			tagsStr := strings.Trim(recipe.Tags, "{}")
			if tagsStr != "" {
				// Split by comma and trim spaces
				tagParts := strings.Split(tagsStr, ",")
				tags = make([]string, len(tagParts))
				for i, tag := range tagParts {
					// Remove quotes if present and trim spaces
					tag = strings.Trim(tag, "\" ")
					tags[i] = tag
				}
			}
		}

		convertedRecipe := repositories.Recipe{
			ID:               recipe.ID,
			Title:            recipe.Title,
			Description:      recipe.Description,
			Servings:         recipe.Servings,
			PrepTimeMinutes:  recipe.PrepTimeMinutes,
			CookTimeMinutes:  recipe.CookTimeMinutes,
			TotalTimeMinutes: recipe.TotalTimeMinutes,
			Ingredients:      convertToRepoIngredients(ingredients),
			Instructions:     convertToRepoInstructions(instructions),
			Nutrition:        convertToRepoNutrition(nutrition),
			Tags:             tags,
			Difficulty:       recipe.Difficulty,
			UserID:           recipe.UserID,
		}

		convertedExactMatches = append(convertedExactMatches, convertedRecipe)
	}

	// Then find similar recipes using vector similarity search
	var similarRecipes []struct {
		ID               string    `gorm:"column:id"`
		Title            string    `gorm:"column:title"`
		Description      string    `gorm:"column:description"`
		Servings         int       `gorm:"column:servings"`
		PrepTimeMinutes  int       `gorm:"column:prep_time_minutes"`
		CookTimeMinutes  int       `gorm:"column:cook_time_minutes"`
		TotalTimeMinutes int       `gorm:"column:total_time_minutes"`
		Ingredients      []byte    `gorm:"column:ingredients"`
		Instructions     []byte    `gorm:"column:instructions"`
		Nutrition        []byte    `gorm:"column:nutrition"`
		Tags             string    `gorm:"column:tags"`
		Difficulty       string    `gorm:"column:difficulty"`
		CreatedAt        time.Time `gorm:"column:created_at"`
		UpdatedAt        time.Time `gorm:"column:updated_at"`
		UserID           string    `gorm:"column:user_id"`
		Similarity       float64   `gorm:"column:similarity"`
	}

	// Log the vector search
	log.Printf("Performing vector similarity search with query: %s", req.Query)

	if err := h.db.Raw(`
		SELECT id, title, description, servings, prep_time_minutes, cook_time_minutes, total_time_minutes, 
		       ingredients, instructions, nutrition, tags, difficulty, created_at, updated_at, user_id,
		       1 - ((embedding->>'data')::float[]::vector <=> $1::float[]::vector) as similarity
		FROM recipes
		WHERE 1 - ((embedding->>'data')::float[]::vector <=> $1::float[]::vector) > 0.5
		ORDER BY similarity DESC
		LIMIT 5
	`, queryEmbeddingFloat32).Scan(&similarRecipes).Error; err != nil {
		log.Printf("Failed to find similar recipes: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to search recipes",
		})
		return
	}

	log.Printf("Found %d similar matches", len(similarRecipes))
	for _, match := range similarRecipes {
		log.Printf("Similar match found - ID: %s, Title: %s, Similarity: %f", match.ID, match.Title, match.Similarity)
	}

	// Convert similar recipes to repository types
	convertedSimilarRecipes := make([]repositories.Recipe, 0, len(similarRecipes))
	for _, recipe := range similarRecipes {
		// Parse ingredients
		var ingredients []models.Ingredient
		if len(recipe.Ingredients) > 0 {
			if err := json.Unmarshal(recipe.Ingredients, &ingredients); err != nil {
				log.Printf("Failed to parse ingredients for recipe %s: %v", recipe.ID, err)
				continue
			}
		}

		// Parse instructions
		var instructions []models.Instruction
		if len(recipe.Instructions) > 0 {
			if err := json.Unmarshal(recipe.Instructions, &instructions); err != nil {
				log.Printf("Failed to parse instructions for recipe %s: %v", recipe.ID, err)
				continue
			}
		}

		// Parse nutrition
		var nutrition models.Nutrition
		if len(recipe.Nutrition) > 0 {
			if err := json.Unmarshal(recipe.Nutrition, &nutrition); err != nil {
				log.Printf("Failed to parse nutrition for recipe %s: %v", recipe.ID, err)
				continue
			}
		}

		// Parse tags from PostgreSQL array format
		var tags []string
		if recipe.Tags != "" {
			// Remove the curly braces and split by comma
			tagsStr := strings.Trim(recipe.Tags, "{}")
			if tagsStr != "" {
				// Split by comma and trim spaces
				tagParts := strings.Split(tagsStr, ",")
				tags = make([]string, len(tagParts))
				for i, tag := range tagParts {
					// Remove quotes if present and trim spaces
					tag = strings.Trim(tag, "\" ")
					tags[i] = tag
				}
			}
		}

		convertedRecipe := repositories.Recipe{
			ID:               recipe.ID,
			Title:            recipe.Title,
			Description:      recipe.Description,
			Servings:         recipe.Servings,
			PrepTimeMinutes:  recipe.PrepTimeMinutes,
			CookTimeMinutes:  recipe.CookTimeMinutes,
			TotalTimeMinutes: recipe.TotalTimeMinutes,
			Ingredients:      convertToRepoIngredients(ingredients),
			Instructions:     convertToRepoInstructions(instructions),
			Nutrition:        convertToRepoNutrition(nutrition),
			Tags:             tags,
			Difficulty:       recipe.Difficulty,
			UserID:           recipe.UserID,
		}

		convertedSimilarRecipes = append(convertedSimilarRecipes, convertedRecipe)
	}

	// Filter out any recipes that are already in exact matches
	filteredSimilar := make([]repositories.Recipe, 0)
	for _, similar := range convertedSimilarRecipes {
		isExactMatch := false
		for _, exact := range convertedExactMatches {
			if similar.ID == exact.ID {
				isExactMatch = true
				break
			}
		}
		if !isExactMatch {
			filteredSimilar = append(filteredSimilar, similar)
		}
	}

	log.Printf("Returning %d exact matches and %d similar matches", len(convertedExactMatches), len(filteredSimilar))

	// Return both exact and similar matches
	c.JSON(http.StatusOK, gin.H{
		"exact_matches":   convertedExactMatches,
		"similar_matches": filteredSimilar,
		"message":         "If you don't find what you're looking for, you can generate a new recipe",
	})
}

// StartRecipeModification handles the request to start modifying a recipe
func (h *RecipeHandler) StartRecipeModification(c *gin.Context) {
	// Get recipe ID from path
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Recipe ID is required",
		})
		return
	}

	// Get recipe from database with specific fields to avoid embedding scan issues
	var recipe struct {
		ID               string    `gorm:"column:id"`
		Title            string    `gorm:"column:title"`
		Description      string    `gorm:"column:description"`
		Servings         int       `gorm:"column:servings"`
		PrepTimeMinutes  int       `gorm:"column:prep_time_minutes"`
		CookTimeMinutes  int       `gorm:"column:cook_time_minutes"`
		TotalTimeMinutes int       `gorm:"column:total_time_minutes"`
		Ingredients      []byte    `gorm:"column:ingredients"`
		Instructions     []byte    `gorm:"column:instructions"`
		Nutrition        []byte    `gorm:"column:nutrition"`
		Tags             string    `gorm:"column:tags"`
		Difficulty       string    `gorm:"column:difficulty"`
		CreatedAt        time.Time `gorm:"column:created_at"`
		UpdatedAt        time.Time `gorm:"column:updated_at"`
	}

	if err := h.db.Table("recipes").
		Select("id, title, description, servings, prep_time_minutes, cook_time_minutes, total_time_minutes, ingredients, instructions, nutrition, tags, difficulty, created_at, updated_at").
		Where("id = ?", recipeID).
		First(&recipe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: "Recipe not found",
			})
			return
		}
		log.Printf("Failed to get recipe from database: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to get recipe",
		})
		return
	}

	// Parse ingredients
	var ingredients []models.Ingredient
	if len(recipe.Ingredients) > 0 {
		if err := json.Unmarshal(recipe.Ingredients, &ingredients); err != nil {
			log.Printf("Failed to parse ingredients for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Ingredients))
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "DATABASE_ERROR",
				Message: "Failed to parse recipe data",
			})
			return
		}
	}

	// Parse instructions
	var instructions []models.Instruction
	if len(recipe.Instructions) > 0 {
		if err := json.Unmarshal(recipe.Instructions, &instructions); err != nil {
			log.Printf("Failed to parse instructions for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Instructions))
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "DATABASE_ERROR",
				Message: "Failed to parse recipe data",
			})
			return
		}
	}

	// Parse nutrition
	var nutrition models.Nutrition
	if len(recipe.Nutrition) > 0 {
		if err := json.Unmarshal(recipe.Nutrition, &nutrition); err != nil {
			log.Printf("Failed to parse nutrition for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Nutrition))
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "DATABASE_ERROR",
				Message: "Failed to parse recipe data",
			})
			return
		}
	}

	// Parse tags from PostgreSQL array format
	var tags []string
	if recipe.Tags != "" {
		// Remove the curly braces and split by comma
		tagsStr := strings.Trim(recipe.Tags, "{}")
		if tagsStr != "" {
			// Split by comma and trim spaces
			tagParts := strings.Split(tagsStr, ",")
			tags = make([]string, len(tagParts))
			for i, tag := range tagParts {
				// Remove quotes if present and trim spaces
				tag = strings.Trim(tag, "\" ")
				tags[i] = tag
			}
		}
	}

	// Convert to repository type
	repoRecipe := repositories.Recipe{
		ID:               recipe.ID,
		Title:            recipe.Title,
		Description:      recipe.Description,
		Servings:         recipe.Servings,
		PrepTimeMinutes:  recipe.PrepTimeMinutes,
		CookTimeMinutes:  recipe.CookTimeMinutes,
		TotalTimeMinutes: recipe.TotalTimeMinutes,
		Ingredients:      convertToRepoIngredients(ingredients),
		Instructions:     convertToRepoInstructions(instructions),
		Nutrition:        convertToRepoNutrition(nutrition),
		Tags:             tags,
		Difficulty:       recipe.Difficulty,
	}

	// Cache the recipe in Redis with modification count 0
	tempID, err := h.recipeCache.CacheRecipe(c.Request.Context(), repoRecipe)
	if err != nil {
		log.Printf("Failed to cache recipe in Redis: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to cache recipe for modification",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recipe_id": tempID,
		"recipe":    repoRecipe,
		"status":    "ready_for_modification",
	})
}

// ModifyRecipe handles recipe modification requests
func (h *RecipeHandler) ModifyRecipe(c *gin.Context) {
	// Get recipe ID from path
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Recipe ID is required",
		})
		return
	}

	// Parse the modification request
	var modification struct {
		Title            *string                    `json:"title,omitempty"`
		Description      *string                    `json:"description,omitempty"`
		Servings         *int                       `json:"servings,omitempty"`
		PrepTimeMinutes  *int                       `json:"prep_time_minutes,omitempty"`
		CookTimeMinutes  *int                       `json:"cook_time_minutes,omitempty"`
		TotalTimeMinutes *int                       `json:"total_time_minutes,omitempty"`
		Ingredients      []repositories.Ingredient  `json:"ingredients,omitempty"`
		Instructions     []repositories.Instruction `json:"instructions,omitempty"`
		Nutrition        *repositories.Nutrition    `json:"nutrition,omitempty"`
		Tags             []string                   `json:"tags,omitempty"`
		Difficulty       *string                    `json:"difficulty,omitempty"`
	}

	if err := c.ShouldBindJSON(&modification); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	// First try to get from Redis
	cachedRecipe, err := h.recipeCache.GetRecipe(c.Request.Context(), recipeID)
	if err == nil {
		// Use the current version from the cache
		repoRecipe := cachedRecipe.Current

		// Apply modifications
		if modification.Title != nil {
			repoRecipe.Title = *modification.Title
		}
		if modification.Description != nil {
			repoRecipe.Description = *modification.Description
		}
		if modification.Servings != nil {
			repoRecipe.Servings = *modification.Servings
		}
		if modification.PrepTimeMinutes != nil {
			repoRecipe.PrepTimeMinutes = *modification.PrepTimeMinutes
		}
		if modification.CookTimeMinutes != nil {
			repoRecipe.CookTimeMinutes = *modification.CookTimeMinutes
		}
		if modification.TotalTimeMinutes != nil {
			repoRecipe.TotalTimeMinutes = *modification.TotalTimeMinutes
		}
		if modification.Ingredients != nil {
			repoRecipe.Ingredients = modification.Ingredients
		}
		if modification.Instructions != nil {
			repoRecipe.Instructions = modification.Instructions
		}
		if modification.Nutrition != nil {
			repoRecipe.Nutrition = *modification.Nutrition
		}
		if modification.Tags != nil {
			repoRecipe.Tags = modification.Tags
		}
		if modification.Difficulty != nil {
			repoRecipe.Difficulty = *modification.Difficulty
		}

		// Update in Redis
		if err := h.recipeCache.UpdateRecipe(c.Request.Context(), recipeID, repoRecipe); err != nil {
			log.Printf("Failed to update recipe in Redis: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update recipe",
			})
			return
		}

		// If the recipe was originally from the database, update it there as well
		if cachedRecipe.ModificationCount == 0 {
			// Prepare text for embedding if text content changed
			if modification.Title != nil || modification.Description != nil || modification.Tags != nil || modification.Ingredients != nil {
				// Prepare text for embedding
				embedText := fmt.Sprintf("%s\n%s\n%s\n%s",
					repoRecipe.Title,
					repoRecipe.Description,
					strings.Join(repoRecipe.Tags, ", "),
					strings.Join(func() []string {
						items := make([]string, len(repoRecipe.Ingredients))
						for i, ing := range repoRecipe.Ingredients {
							items[i] = ing.Item
						}
						return items
					}(), ", "),
				)

				// Get new embeddings
				resp, err := h.deepseekClient.GetEmbeddings(
					c.Request.Context(),
					embedText,
				)
				if err != nil {
					log.Printf("Failed to get embeddings: %v", err)
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "EMBEDDING_ERROR",
						Message: "Failed to generate embeddings",
					})
					return
				}

				if len(resp.Data) == 0 {
					log.Printf("No embeddings returned from DeepSeek")
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "EMBEDDING_ERROR",
						Message: "No embeddings returned",
					})
					return
				}

				// Convert embeddings to []float32
				embedding := make([]float32, len(resp.Data[0].Embedding))
				for i, v := range resp.Data[0].Embedding {
					embedding[i] = float32(v)
				}

				// Convert embedding to JSONB format
				embeddingJSON, err := json.Marshal(embedding)
				if err != nil {
					log.Printf("Failed to marshal embedding: %v", err)
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "INTERNAL_ERROR",
						Message: "Failed to process embedding",
					})
					return
				}

				// Update the recipe in the database
				if err := h.db.Model(&models.Recipe{}).Where("id = ?", recipeID).Updates(map[string]interface{}{
					"title":              repoRecipe.Title,
					"description":        repoRecipe.Description,
					"servings":           repoRecipe.Servings,
					"prep_time_minutes":  repoRecipe.PrepTimeMinutes,
					"cook_time_minutes":  repoRecipe.CookTimeMinutes,
					"total_time_minutes": repoRecipe.TotalTimeMinutes,
					"ingredients":        convertToModelIngredients(repoRecipe.Ingredients),
					"instructions":       convertToModelInstructions(repoRecipe.Instructions),
					"nutrition":          convertToModelNutrition(repoRecipe.Nutrition),
					"tags":               repoRecipe.Tags,
					"difficulty":         repoRecipe.Difficulty,
					"embedding":          embeddingJSON,
				}).Error; err != nil {
					log.Printf("Failed to update recipe in database: %v", err)
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "DATABASE_ERROR",
						Message: "Failed to update recipe in database",
					})
					return
				}
			} else {
				// Update the recipe in the database without changing embeddings
				if err := h.db.Model(&models.Recipe{}).Where("id = ?", recipeID).Updates(map[string]interface{}{
					"title":              repoRecipe.Title,
					"description":        repoRecipe.Description,
					"servings":           repoRecipe.Servings,
					"prep_time_minutes":  repoRecipe.PrepTimeMinutes,
					"cook_time_minutes":  repoRecipe.CookTimeMinutes,
					"total_time_minutes": repoRecipe.TotalTimeMinutes,
					"ingredients":        convertToModelIngredients(repoRecipe.Ingredients),
					"instructions":       convertToModelInstructions(repoRecipe.Instructions),
					"nutrition":          convertToModelNutrition(repoRecipe.Nutrition),
					"tags":               repoRecipe.Tags,
					"difficulty":         repoRecipe.Difficulty,
				}).Error; err != nil {
					log.Printf("Failed to update recipe in database: %v", err)
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "DATABASE_ERROR",
						Message: "Failed to update recipe in database",
					})
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"recipe_id": recipeID,
			"recipe":    repoRecipe,
		})
	} else {
		// If not in Redis, get from database
		var dbRecipe struct {
			ID               string    `gorm:"column:id"`
			Title            string    `gorm:"column:title"`
			Description      string    `gorm:"column:description"`
			Servings         int       `gorm:"column:servings"`
			PrepTimeMinutes  int       `gorm:"column:prep_time_minutes"`
			CookTimeMinutes  int       `gorm:"column:cook_time_minutes"`
			TotalTimeMinutes int       `gorm:"column:total_time_minutes"`
			Ingredients      []byte    `gorm:"column:ingredients"`
			Instructions     []byte    `gorm:"column:instructions"`
			Nutrition        []byte    `gorm:"column:nutrition"`
			Tags             string    `gorm:"column:tags"`
			Difficulty       string    `gorm:"column:difficulty"`
			CreatedAt        time.Time `gorm:"column:created_at"`
			UpdatedAt        time.Time `gorm:"column:updated_at"`
		}

		if err := h.db.Table("recipes").
			Select("id, title, description, servings, prep_time_minutes, cook_time_minutes, total_time_minutes, ingredients, instructions, nutrition, tags, difficulty, created_at, updated_at").
			Where("id = ?", recipeID).
			First(&dbRecipe).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, dtos.ErrorResponse{
					Code:    "NOT_FOUND",
					Message: "Recipe not found",
				})
				return
			}
			log.Printf("Failed to get recipe from database: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "DATABASE_ERROR",
				Message: "Failed to get recipe",
			})
			return
		}

		// Parse ingredients
		var ingredients []models.Ingredient
		if len(dbRecipe.Ingredients) > 0 {
			if err := json.Unmarshal(dbRecipe.Ingredients, &ingredients); err != nil {
				log.Printf("Failed to parse ingredients for recipe %s: %v\nRaw data: %s", dbRecipe.ID, err, string(dbRecipe.Ingredients))
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Code:    "DATABASE_ERROR",
					Message: "Failed to parse recipe data",
				})
				return
			}
		}

		// Parse instructions
		var instructions []models.Instruction
		if len(dbRecipe.Instructions) > 0 {
			if err := json.Unmarshal(dbRecipe.Instructions, &instructions); err != nil {
				log.Printf("Failed to parse instructions for recipe %s: %v\nRaw data: %s", dbRecipe.ID, err, string(dbRecipe.Instructions))
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Code:    "DATABASE_ERROR",
					Message: "Failed to parse recipe data",
				})
				return
			}
		}

		// Parse nutrition
		var nutrition models.Nutrition
		if len(dbRecipe.Nutrition) > 0 {
			if err := json.Unmarshal(dbRecipe.Nutrition, &nutrition); err != nil {
				log.Printf("Failed to parse nutrition for recipe %s: %v\nRaw data: %s", dbRecipe.ID, err, string(dbRecipe.Nutrition))
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Code:    "DATABASE_ERROR",
					Message: "Failed to parse recipe data",
				})
				return
			}
		}

		// Parse tags from PostgreSQL array format
		var tags []string
		if dbRecipe.Tags != "" {
			// Remove the curly braces and split by comma
			tagsStr := strings.Trim(dbRecipe.Tags, "{}")
			if tagsStr != "" {
				// Split by comma and trim spaces
				tagParts := strings.Split(tagsStr, ",")
				tags = make([]string, len(tagParts))
				for i, tag := range tagParts {
					// Remove quotes if present and trim spaces
					tag = strings.Trim(tag, "\" ")
					tags[i] = tag
				}
			}
		}

		// Convert to repository type
		repoRecipe := repositories.Recipe{
			Title:            dbRecipe.Title,
			Description:      dbRecipe.Description,
			Servings:         dbRecipe.Servings,
			PrepTimeMinutes:  dbRecipe.PrepTimeMinutes,
			CookTimeMinutes:  dbRecipe.CookTimeMinutes,
			TotalTimeMinutes: dbRecipe.TotalTimeMinutes,
			Ingredients:      convertToRepoIngredients(ingredients),
			Instructions:     convertToRepoInstructions(instructions),
			Nutrition:        convertToRepoNutrition(nutrition),
			Tags:             tags,
			Difficulty:       dbRecipe.Difficulty,
		}

		// Cache the recipe in Redis
		tempID, err := h.recipeCache.CacheRecipe(c.Request.Context(), repoRecipe)
		if err != nil {
			log.Printf("Failed to cache recipe in Redis: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to cache recipe for modification",
			})
			return
		}

		// Get the cached recipe using the temporary ID
		cachedRecipe, err = h.recipeCache.GetRecipe(c.Request.Context(), tempID)
		if err != nil {
			log.Printf("Failed to get cached recipe: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get cached recipe",
			})
			return
		}

		// Apply modifications
		if modification.Title != nil {
			cachedRecipe.Current.Title = *modification.Title
		}
		if modification.Description != nil {
			cachedRecipe.Current.Description = *modification.Description
		}
		if modification.Servings != nil {
			cachedRecipe.Current.Servings = *modification.Servings
		}
		if modification.PrepTimeMinutes != nil {
			cachedRecipe.Current.PrepTimeMinutes = *modification.PrepTimeMinutes
		}
		if modification.CookTimeMinutes != nil {
			cachedRecipe.Current.CookTimeMinutes = *modification.CookTimeMinutes
		}
		if modification.TotalTimeMinutes != nil {
			cachedRecipe.Current.TotalTimeMinutes = *modification.TotalTimeMinutes
		}
		if modification.Ingredients != nil {
			cachedRecipe.Current.Ingredients = modification.Ingredients
		}
		if modification.Instructions != nil {
			cachedRecipe.Current.Instructions = modification.Instructions
		}
		if modification.Nutrition != nil {
			cachedRecipe.Current.Nutrition = *modification.Nutrition
		}
		if modification.Tags != nil {
			cachedRecipe.Current.Tags = modification.Tags
		}
		if modification.Difficulty != nil {
			cachedRecipe.Current.Difficulty = *modification.Difficulty
		}

		// Update in Redis
		if err := h.recipeCache.UpdateRecipe(c.Request.Context(), recipeID, cachedRecipe.Current); err != nil {
			log.Printf("Failed to update recipe in Redis: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update recipe",
			})
			return
		}

		// If the recipe was originally from the database, update it there as well
		if cachedRecipe.ModificationCount == 0 {
			// Prepare text for embedding if text content changed
			if modification.Title != nil || modification.Description != nil || modification.Tags != nil || modification.Ingredients != nil {
				// Prepare text for embedding
				embedText := fmt.Sprintf("%s\n%s\n%s\n%s",
					cachedRecipe.Current.Title,
					cachedRecipe.Current.Description,
					strings.Join(cachedRecipe.Current.Tags, ", "),
					strings.Join(func() []string {
						items := make([]string, len(cachedRecipe.Current.Ingredients))
						for i, ing := range cachedRecipe.Current.Ingredients {
							items[i] = ing.Item
						}
						return items
					}(), ", "),
				)

				// Get new embeddings
				resp, err := h.deepseekClient.GetEmbeddings(
					c.Request.Context(),
					embedText,
				)
				if err != nil {
					log.Printf("Failed to get embeddings: %v", err)
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "EMBEDDING_ERROR",
						Message: "Failed to generate embeddings",
					})
					return
				}

				if len(resp.Data) == 0 {
					log.Printf("No embeddings returned from DeepSeek")
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "EMBEDDING_ERROR",
						Message: "No embeddings returned",
					})
					return
				}

				// Convert embeddings to []float32
				embedding := make([]float32, len(resp.Data[0].Embedding))
				for i, v := range resp.Data[0].Embedding {
					embedding[i] = float32(v)
				}

				// Convert embedding to JSONB format
				embeddingJSON, err := json.Marshal(embedding)
				if err != nil {
					log.Printf("Failed to marshal embedding: %v", err)
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "INTERNAL_ERROR",
						Message: "Failed to process embedding",
					})
					return
				}

				// Update the recipe in the database
				if err := h.db.Model(&models.Recipe{}).Where("id = ?", recipeID).Updates(map[string]interface{}{
					"title":              cachedRecipe.Current.Title,
					"description":        cachedRecipe.Current.Description,
					"servings":           cachedRecipe.Current.Servings,
					"prep_time_minutes":  cachedRecipe.Current.PrepTimeMinutes,
					"cook_time_minutes":  cachedRecipe.Current.CookTimeMinutes,
					"total_time_minutes": cachedRecipe.Current.TotalTimeMinutes,
					"ingredients":        convertToModelIngredients(cachedRecipe.Current.Ingredients),
					"instructions":       convertToModelInstructions(cachedRecipe.Current.Instructions),
					"nutrition":          convertToModelNutrition(cachedRecipe.Current.Nutrition),
					"tags":               cachedRecipe.Current.Tags,
					"difficulty":         cachedRecipe.Current.Difficulty,
					"embedding":          embeddingJSON,
				}).Error; err != nil {
					log.Printf("Failed to update recipe in database: %v", err)
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "DATABASE_ERROR",
						Message: "Failed to update recipe in database",
					})
					return
				}
			} else {
				// Update the recipe in the database without changing embeddings
				if err := h.db.Model(&models.Recipe{}).Where("id = ?", recipeID).Updates(map[string]interface{}{
					"title":              cachedRecipe.Current.Title,
					"description":        cachedRecipe.Current.Description,
					"servings":           cachedRecipe.Current.Servings,
					"prep_time_minutes":  cachedRecipe.Current.PrepTimeMinutes,
					"cook_time_minutes":  cachedRecipe.Current.CookTimeMinutes,
					"total_time_minutes": cachedRecipe.Current.TotalTimeMinutes,
					"ingredients":        convertToModelIngredients(cachedRecipe.Current.Ingredients),
					"instructions":       convertToModelInstructions(cachedRecipe.Current.Instructions),
					"nutrition":          convertToModelNutrition(cachedRecipe.Current.Nutrition),
					"tags":               cachedRecipe.Current.Tags,
					"difficulty":         cachedRecipe.Current.Difficulty,
				}).Error; err != nil {
					log.Printf("Failed to update recipe in database: %v", err)
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
						Code:    "DATABASE_ERROR",
						Message: "Failed to update recipe in database",
					})
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"recipe_id": recipeID,
			"recipe":    cachedRecipe.Current,
		})
	}
}

// Convert model types to repository types
func convertToRepoIngredients(modelIngredients []models.Ingredient) []repositories.Ingredient {
	repoIngredients := make([]repositories.Ingredient, len(modelIngredients))
	for i, ing := range modelIngredients {
		repoIngredients[i] = repositories.Ingredient{
			Item:   ing.Item,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	return repoIngredients
}

func convertToRepoInstructions(modelInstructions []models.Instruction) []repositories.Instruction {
	repoInstructions := make([]repositories.Instruction, len(modelInstructions))
	for i, inst := range modelInstructions {
		repoInstructions[i] = repositories.Instruction{
			Step:        inst.Step,
			Description: inst.Description,
		}
	}
	return repoInstructions
}

func convertToRepoNutrition(modelNutrition models.Nutrition) repositories.Nutrition {
	return repositories.Nutrition{
		Calories: modelNutrition.Calories,
		Protein:  modelNutrition.Protein,
		Carbs:    modelNutrition.Carbs,
		Fat:      modelNutrition.Fat,
	}
}

// GetRecipe retrieves a single recipe by ID
func (h *RecipeHandler) GetRecipe(c *gin.Context) {
	// Get recipe ID from path
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Recipe ID is required",
		})
		return
	}

	// First try to get from Redis
	cachedRecipe, err := h.recipeCache.GetRecipe(c.Request.Context(), recipeID)
	if err == nil {
		// Ensure the ID is set in the recipe object
		cachedRecipe.Current.ID = recipeID
		c.JSON(http.StatusOK, gin.H{
			"recipe": cachedRecipe.Current,
		})
		return
	}

	// If not in Redis, get from database with specific fields to avoid embedding scan issues
	var recipe struct {
		ID               string    `gorm:"column:id"`
		Title            string    `gorm:"column:title"`
		Description      string    `gorm:"column:description"`
		Servings         int       `gorm:"column:servings"`
		PrepTimeMinutes  int       `gorm:"column:prep_time_minutes"`
		CookTimeMinutes  int       `gorm:"column:cook_time_minutes"`
		TotalTimeMinutes int       `gorm:"column:total_time_minutes"`
		Ingredients      []byte    `gorm:"column:ingredients"`
		Instructions     []byte    `gorm:"column:instructions"`
		Nutrition        []byte    `gorm:"column:nutrition"`
		Tags             string    `gorm:"column:tags"`
		Difficulty       string    `gorm:"column:difficulty"`
		CreatedAt        time.Time `gorm:"column:created_at"`
		UpdatedAt        time.Time `gorm:"column:updated_at"`
	}

	if err := h.db.Table("recipes").
		Select("id, title, description, servings, prep_time_minutes, cook_time_minutes, total_time_minutes, ingredients, instructions, nutrition, tags, difficulty, created_at, updated_at").
		Where("id = ?", recipeID).
		First(&recipe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: "Recipe not found",
			})
			return
		}
		log.Printf("Failed to get recipe from database: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to get recipe",
		})
		return
	}

	// Parse ingredients
	var ingredients []models.Ingredient
	if len(recipe.Ingredients) > 0 {
		if err := json.Unmarshal(recipe.Ingredients, &ingredients); err != nil {
			log.Printf("Failed to parse ingredients for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Ingredients))
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "DATABASE_ERROR",
				Message: "Failed to parse recipe data",
			})
			return
		}
	}

	// Parse instructions
	var instructions []models.Instruction
	if len(recipe.Instructions) > 0 {
		if err := json.Unmarshal(recipe.Instructions, &instructions); err != nil {
			log.Printf("Failed to parse instructions for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Instructions))
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "DATABASE_ERROR",
				Message: "Failed to parse recipe data",
			})
			return
		}
	}

	// Parse nutrition
	var nutrition models.Nutrition
	if len(recipe.Nutrition) > 0 {
		if err := json.Unmarshal(recipe.Nutrition, &nutrition); err != nil {
			log.Printf("Failed to parse nutrition for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Nutrition))
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "DATABASE_ERROR",
				Message: "Failed to parse recipe data",
			})
			return
		}
	}

	// Parse tags from PostgreSQL array format
	var tags []string
	if recipe.Tags != "" {
		// Remove the curly braces and split by comma
		tagsStr := strings.Trim(recipe.Tags, "{}")
		if tagsStr != "" {
			// Split by comma and trim spaces
			tagParts := strings.Split(tagsStr, ",")
			tags = make([]string, len(tagParts))
			for i, tag := range tagParts {
				// Remove quotes if present and trim spaces
				tag = strings.Trim(tag, "\" ")
				tags[i] = tag
			}
		}
	}

	// Convert to repository type
	repoRecipe := repositories.Recipe{
		ID:               recipe.ID,
		Title:            recipe.Title,
		Description:      recipe.Description,
		Servings:         recipe.Servings,
		PrepTimeMinutes:  recipe.PrepTimeMinutes,
		CookTimeMinutes:  recipe.CookTimeMinutes,
		TotalTimeMinutes: recipe.TotalTimeMinutes,
		Ingredients:      convertToRepoIngredients(ingredients),
		Instructions:     convertToRepoInstructions(instructions),
		Nutrition:        convertToRepoNutrition(nutrition),
		Tags:             tags,
		Difficulty:       recipe.Difficulty,
	}

	c.JSON(http.StatusOK, gin.H{
		"recipe": repoRecipe,
	})
}

// ListRecipes retrieves a paginated list of recipes
func (h *RecipeHandler) ListRecipes(c *gin.Context) {
	// Parse pagination parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}
	limitNum, err := strconv.Atoi(limit)
	if err != nil || limitNum < 1 || limitNum > 100 {
		limitNum = 10
	}
	offset := (pageNum - 1) * limitNum

	// Get total count of all recipes
	var total int64
	if err := h.db.Model(&models.Recipe{}).Count(&total).Error; err != nil {
		log.Printf("Failed to get recipe count: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to get recipe count",
		})
		return
	}

	// Get recipes from database with specific fields to avoid embedding scan issues
	var recipes []struct {
		ID               string    `gorm:"column:id"`
		Title            string    `gorm:"column:title"`
		Description      string    `gorm:"column:description"`
		Servings         int       `gorm:"column:servings"`
		PrepTimeMinutes  int       `gorm:"column:prep_time_minutes"`
		CookTimeMinutes  int       `gorm:"column:cook_time_minutes"`
		TotalTimeMinutes int       `gorm:"column:total_time_minutes"`
		Ingredients      []byte    `gorm:"column:ingredients"`
		Instructions     []byte    `gorm:"column:instructions"`
		Nutrition        []byte    `gorm:"column:nutrition"`
		Tags             string    `gorm:"column:tags"`
		Difficulty       string    `gorm:"column:difficulty"`
		CreatedAt        time.Time `gorm:"column:created_at"`
		UpdatedAt        time.Time `gorm:"column:updated_at"`
		UserID           string    `gorm:"column:user_id"`
	}

	if err := h.db.Table("recipes").
		Select("id, title, description, servings, prep_time_minutes, cook_time_minutes, total_time_minutes, ingredients, instructions, nutrition, tags, difficulty, created_at, updated_at, user_id").
		Order("created_at DESC").
		Limit(limitNum).
		Offset(offset).
		Find(&recipes).Error; err != nil {
		log.Printf("Failed to get recipes from database: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to get recipes",
		})
		return
	}

	// Convert to repository types
	repoRecipes := make([]repositories.Recipe, 0, len(recipes))
	for _, recipe := range recipes {
		// Parse ingredients
		var ingredients []models.Ingredient
		if len(recipe.Ingredients) > 0 {
			if err := json.Unmarshal(recipe.Ingredients, &ingredients); err != nil {
				log.Printf("Failed to parse ingredients for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Ingredients))
				continue
			}
		}

		// Parse instructions
		var instructions []models.Instruction
		if len(recipe.Instructions) > 0 {
			if err := json.Unmarshal(recipe.Instructions, &instructions); err != nil {
				log.Printf("Failed to parse instructions for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Instructions))
				continue
			}
		}

		// Parse nutrition
		var nutrition models.Nutrition
		if len(recipe.Nutrition) > 0 {
			if err := json.Unmarshal(recipe.Nutrition, &nutrition); err != nil {
				log.Printf("Failed to parse nutrition for recipe %s: %v\nRaw data: %s", recipe.ID, err, string(recipe.Nutrition))
				continue
			}
		}

		// Parse tags from PostgreSQL array format
		var tags []string
		if recipe.Tags != "" {
			// Remove the curly braces and split by comma
			tagsStr := strings.Trim(recipe.Tags, "{}")
			if tagsStr != "" {
				// Split by comma and trim spaces
				tagParts := strings.Split(tagsStr, ",")
				tags = make([]string, len(tagParts))
				for i, tag := range tagParts {
					// Remove quotes if present and trim spaces
					tag = strings.Trim(tag, "\" ")
					tags[i] = tag
				}
			}
		}

		repoRecipe := repositories.Recipe{
			ID:               recipe.ID,
			Title:            recipe.Title,
			Description:      recipe.Description,
			Servings:         recipe.Servings,
			PrepTimeMinutes:  recipe.PrepTimeMinutes,
			CookTimeMinutes:  recipe.CookTimeMinutes,
			TotalTimeMinutes: recipe.TotalTimeMinutes,
			Ingredients:      convertToRepoIngredients(ingredients),
			Instructions:     convertToRepoInstructions(instructions),
			Nutrition:        convertToRepoNutrition(nutrition),
			Tags:             tags,
			Difficulty:       recipe.Difficulty,
			UserID:           recipe.UserID,
		}

		// Only add recipes that have at least some content
		if repoRecipe.Title != "" || repoRecipe.Description != "" || len(repoRecipe.Ingredients) > 0 {
			repoRecipes = append(repoRecipes, repoRecipe)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"recipes": repoRecipes,
		"pagination": gin.H{
			"total":  total,
			"page":   pageNum,
			"limit":  limitNum,
			"pages":  int(math.Ceil(float64(total) / float64(limitNum))),
			"offset": offset,
		},
	})
}

// ModifyRecipeWithAI handles the request to modify a recipe using AI
func (h *RecipeHandler) ModifyRecipeWithAI(c *gin.Context) {
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Recipe ID is required",
		})
		return
	}

	var req struct {
		ModificationType string `json:"modification_type" binding:"required"`
		AdditionalNotes  string `json:"additional_notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	// First try to get from Redis
	cachedRecipe, err := h.recipeCache.GetRecipe(c.Request.Context(), recipeID)
	if err == nil {
		// Use the current version from the cache
		repoRecipe := cachedRecipe.Current

		// Prepare the prompt for DeepSeek
		prompt := fmt.Sprintf(`Modify the following recipe according to these requirements and respond in JSON format:
Modification Type: %s
Additional Notes: %s

IMPORTANT: You MUST create a new title that reflects the modifications made to the recipe. For example, if making a chocolate cake larger and adding nuts, the new title should be something like "Large Chocolate Nut Mug Cake" or "Family-Sized Chocolate Mug Cake with Nuts".

Original Recipe:
Title: %s
Description: %s
Servings: %d
Prep Time: %d minutes
Cook Time: %d minutes
Total Time: %d minutes
Difficulty: %s

Ingredients:
%s

Instructions:
%s

Nutrition:
%s

Tags: %s

Please provide the modified recipe in JSON format with the following structure:
{
  "title": "New Recipe Name (must reflect the modifications)",
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
			req.ModificationType,
			req.AdditionalNotes,
			repoRecipe.Title,
			repoRecipe.Description,
			repoRecipe.Servings,
			repoRecipe.PrepTimeMinutes,
			repoRecipe.CookTimeMinutes,
			repoRecipe.TotalTimeMinutes,
			repoRecipe.Difficulty,
			formatIngredients(repoRecipe.Ingredients),
			formatInstructions(repoRecipe.Instructions),
			formatNutrition(repoRecipe.Nutrition),
			strings.Join(repoRecipe.Tags, ", "))

		// Call DeepSeek to modify the recipe
		modifiedRecipePtr, err := h.deepseekClient.ModifyRecipe(c.Request.Context(), prompt)
		if err != nil {
			log.Printf("Failed to modify recipe with AI: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "AI_ERROR",
				Message: "Failed to modify recipe with AI",
			})
			return
		}

		// Cache the modified recipe in Redis with modification count 0
		tempID, err := h.recipeCache.CacheRecipe(c.Request.Context(), *modifiedRecipePtr)
		if err != nil {
			log.Printf("Failed to cache modified recipe in Redis: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to cache modified recipe",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"recipe_id": tempID,
			"recipe":    *modifiedRecipePtr,
		})
	} else {
		// If not in Redis, get from database
		var dbRecipe struct {
			ID               string    `gorm:"column:id"`
			Title            string    `gorm:"column:title"`
			Description      string    `gorm:"column:description"`
			Servings         int       `gorm:"column:servings"`
			PrepTimeMinutes  int       `gorm:"column:prep_time_minutes"`
			CookTimeMinutes  int       `gorm:"column:cook_time_minutes"`
			TotalTimeMinutes int       `gorm:"column:total_time_minutes"`
			Ingredients      []byte    `gorm:"column:ingredients"`
			Instructions     []byte    `gorm:"column:instructions"`
			Nutrition        []byte    `gorm:"column:nutrition"`
			Tags             string    `gorm:"column:tags"`
			Difficulty       string    `gorm:"column:difficulty"`
			CreatedAt        time.Time `gorm:"column:created_at"`
			UpdatedAt        time.Time `gorm:"column:updated_at"`
		}

		if err := h.db.Table("recipes").
			Select("id, title, description, servings, prep_time_minutes, cook_time_minutes, total_time_minutes, ingredients, instructions, nutrition, tags, difficulty, created_at, updated_at").
			Where("id = ?", recipeID).
			First(&dbRecipe).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, dtos.ErrorResponse{
					Code:    "NOT_FOUND",
					Message: "Recipe not found",
				})
				return
			}
			log.Printf("Failed to get recipe from database: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "DATABASE_ERROR",
				Message: "Failed to get recipe",
			})
			return
		}

		// Parse ingredients
		var ingredients []models.Ingredient
		if len(dbRecipe.Ingredients) > 0 {
			if err := json.Unmarshal(dbRecipe.Ingredients, &ingredients); err != nil {
				log.Printf("Failed to parse ingredients for recipe %s: %v\nRaw data: %s", dbRecipe.ID, err, string(dbRecipe.Ingredients))
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Code:    "DATABASE_ERROR",
					Message: "Failed to parse recipe data",
				})
				return
			}
		}

		// Parse instructions
		var instructions []models.Instruction
		if len(dbRecipe.Instructions) > 0 {
			if err := json.Unmarshal(dbRecipe.Instructions, &instructions); err != nil {
				log.Printf("Failed to parse instructions for recipe %s: %v\nRaw data: %s", dbRecipe.ID, err, string(dbRecipe.Instructions))
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Code:    "DATABASE_ERROR",
					Message: "Failed to parse recipe data",
				})
				return
			}
		}

		// Parse nutrition
		var nutrition models.Nutrition
		if len(dbRecipe.Nutrition) > 0 {
			if err := json.Unmarshal(dbRecipe.Nutrition, &nutrition); err != nil {
				log.Printf("Failed to parse nutrition for recipe %s: %v\nRaw data: %s", dbRecipe.ID, err, string(dbRecipe.Nutrition))
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Code:    "DATABASE_ERROR",
					Message: "Failed to parse recipe data",
				})
				return
			}
		}

		// Parse tags from PostgreSQL array format
		var tags []string
		if dbRecipe.Tags != "" {
			// Remove the curly braces and split by comma
			tagsStr := strings.Trim(dbRecipe.Tags, "{}")
			if tagsStr != "" {
				// Split by comma and trim spaces
				tagParts := strings.Split(tagsStr, ",")
				tags = make([]string, len(tagParts))
				for i, tag := range tagParts {
					// Remove quotes if present and trim spaces
					tag = strings.Trim(tag, "\" ")
					tags[i] = tag
				}
			}
		}

		// Convert to repository type
		repoRecipe := repositories.Recipe{
			Title:            dbRecipe.Title,
			Description:      dbRecipe.Description,
			Servings:         dbRecipe.Servings,
			PrepTimeMinutes:  dbRecipe.PrepTimeMinutes,
			CookTimeMinutes:  dbRecipe.CookTimeMinutes,
			TotalTimeMinutes: dbRecipe.TotalTimeMinutes,
			Ingredients:      convertToRepoIngredients(ingredients),
			Instructions:     convertToRepoInstructions(instructions),
			Nutrition:        convertToRepoNutrition(nutrition),
			Tags:             tags,
			Difficulty:       dbRecipe.Difficulty,
		}

		// Prepare the prompt for DeepSeek
		prompt := fmt.Sprintf(`Modify the following recipe according to these requirements and respond in JSON format:
Modification Type: %s
Additional Notes: %s

IMPORTANT: You MUST create a new title that reflects the modifications made to the recipe. For example, if making a chocolate cake larger and adding nuts, the new title should be something like "Large Chocolate Nut Mug Cake" or "Family-Sized Chocolate Mug Cake with Nuts".

Original Recipe:
Title: %s
Description: %s
Servings: %d
Prep Time: %d minutes
Cook Time: %d minutes
Total Time: %d minutes
Difficulty: %s

Ingredients:
%s

Instructions:
%s

Nutrition:
%s

Tags: %s

Please provide the modified recipe in JSON format with the following structure:
{
  "title": "New Recipe Name (must reflect the modifications)",
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
			req.ModificationType,
			req.AdditionalNotes,
			repoRecipe.Title,
			repoRecipe.Description,
			repoRecipe.Servings,
			repoRecipe.PrepTimeMinutes,
			repoRecipe.CookTimeMinutes,
			repoRecipe.TotalTimeMinutes,
			repoRecipe.Difficulty,
			formatIngredients(repoRecipe.Ingredients),
			formatInstructions(repoRecipe.Instructions),
			formatNutrition(repoRecipe.Nutrition),
			strings.Join(repoRecipe.Tags, ", "))

		// Call DeepSeek to modify the recipe
		modifiedRecipePtr, err := h.deepseekClient.ModifyRecipe(c.Request.Context(), prompt)
		if err != nil {
			log.Printf("Failed to modify recipe with AI: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "AI_ERROR",
				Message: "Failed to modify recipe with AI",
			})
			return
		}

		// Cache the modified recipe in Redis with modification count 0
		tempID, err := h.recipeCache.CacheRecipe(c.Request.Context(), *modifiedRecipePtr)
		if err != nil {
			log.Printf("Failed to cache modified recipe in Redis: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to cache modified recipe",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"recipe_id": tempID,
			"recipe":    *modifiedRecipePtr,
		})
	}
}

// Helper functions for formatting recipe components
func formatIngredients(ingredients []repositories.Ingredient) string {
	var sb strings.Builder
	for _, ing := range ingredients {
		sb.WriteString(fmt.Sprintf("- %s: %s %s\n", ing.Item, ing.Amount.String(), ing.Unit))
	}
	return sb.String()
}

func formatInstructions(instructions []repositories.Instruction) string {
	var sb strings.Builder
	for _, inst := range instructions {
		sb.WriteString(fmt.Sprintf("%d. %s\n", inst.Step, inst.Description))
	}
	return sb.String()
}

func formatNutrition(nutrition repositories.Nutrition) string {
	return fmt.Sprintf("Calories: %d\nProtein: %s\nCarbs: %s\nFat: %s",
		nutrition.Calories,
		nutrition.Protein,
		nutrition.Carbs,
		nutrition.Fat)
}
