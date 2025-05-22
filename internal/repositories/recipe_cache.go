package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RecipeCache represents a cached recipe with version history
type RecipeCache struct {
	client *redis.Client
}

// NewRecipeCache creates a new RecipeCache instance
func NewRecipeCache(client *redis.Client) *RecipeCache {
	return &RecipeCache{
		client: client,
	}
}

// CachedRecipe represents a recipe stored in Redis with version history
type CachedRecipe struct {
	Original          Recipe `json:"original"`
	Current           Recipe `json:"current"`
	ModificationCount int    `json:"modification_count"`
	LastModified      int64  `json:"last_modified"`
}

// CacheRecipe stores a recipe in Redis with a UUID
func (c *RecipeCache) CacheRecipe(ctx context.Context, recipe Recipe) (string, error) {
	// Generate UUID for the recipe
	recipeID := uuid.New().String()

	// Set the ID in both original and current recipe objects
	recipe.ID = recipeID

	cachedRecipe := CachedRecipe{
		Original:          recipe,
		Current:           recipe,
		ModificationCount: 0,
		LastModified:      time.Now().Unix(),
	}

	data, err := json.Marshal(cachedRecipe)
	if err != nil {
		return "", fmt.Errorf("failed to marshal recipe: %v", err)
	}

	key := fmt.Sprintf("recipe:%s", recipeID)
	if err := c.client.Set(ctx, key, data, 1*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("failed to cache recipe: %v", err)
	}

	return recipeID, nil
}

// GetRecipe retrieves a recipe from Redis by its UUID
func (c *RecipeCache) GetRecipe(ctx context.Context, recipeID string) (*CachedRecipe, error) {
	key := fmt.Sprintf("recipe:%s", recipeID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe from cache: %v", err)
	}

	var cachedRecipe CachedRecipe
	if err := json.Unmarshal(data, &cachedRecipe); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipe: %v", err)
	}

	return &cachedRecipe, nil
}

// UpdateRecipe updates a recipe in Redis by its UUID
func (c *RecipeCache) UpdateRecipe(ctx context.Context, recipeID string, updatedRecipe Recipe) error {
	cachedRecipe, err := c.GetRecipe(ctx, recipeID)
	if err != nil {
		return err
	}

	// Ensure the ID is preserved
	updatedRecipe.ID = recipeID

	cachedRecipe.Current = updatedRecipe
	cachedRecipe.ModificationCount++
	cachedRecipe.LastModified = time.Now().Unix()

	data, err := json.Marshal(cachedRecipe)
	if err != nil {
		return fmt.Errorf("failed to marshal updated recipe: %v", err)
	}

	key := fmt.Sprintf("recipe:%s", recipeID)
	if err := c.client.Set(ctx, key, data, 1*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to update recipe in cache: %v", err)
	}

	return nil
}

// DeleteRecipe removes a recipe from Redis by its UUID
func (c *RecipeCache) DeleteRecipe(ctx context.Context, recipeID string) error {
	key := fmt.Sprintf("recipe:%s", recipeID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete recipe from cache: %v", err)
	}
	return nil
}
