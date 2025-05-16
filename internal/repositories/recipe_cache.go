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
	Original          Recipe    `json:"original"`
	Versions          []Recipe  `json:"versions"`
	Current           Recipe    `json:"current"`
	ModificationCount int       `json:"modification_count"`
	Timestamp         time.Time `json:"timestamp"`
}

// CacheRecipe stores a recipe in Redis with a temporary ID
func (rc *RecipeCache) CacheRecipe(ctx context.Context, recipe Recipe) (string, error) {
	// Generate a temporary ID for the recipe
	tempID := uuid.New().String()

	// Create the cache structure
	cache := CachedRecipe{
		Original:          recipe,
		Versions:          []Recipe{},
		Current:           recipe,
		ModificationCount: 0,
		Timestamp:         time.Now(),
	}

	// Convert to JSON
	cacheJSON, err := json.Marshal(cache)
	if err != nil {
		return "", fmt.Errorf("failed to marshal recipe cache: %v", err)
	}

	// Store in Redis with 1 hour expiration
	key := fmt.Sprintf("recipe:%s", tempID)
	if err := rc.client.Set(ctx, key, cacheJSON, time.Hour).Err(); err != nil {
		return "", fmt.Errorf("failed to cache recipe: %v", err)
	}

	return tempID, nil
}

// GetRecipe retrieves a recipe from Redis by its temporary ID
func (rc *RecipeCache) GetRecipe(ctx context.Context, tempID string) (*CachedRecipe, error) {
	key := fmt.Sprintf("recipe:%s", tempID)

	// Get from Redis
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("recipe not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get recipe from cache: %v", err)
	}

	// Parse JSON
	var cache CachedRecipe
	if err := json.Unmarshal([]byte(val), &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipe cache: %v", err)
	}

	return &cache, nil
}

// UpdateRecipe updates a cached recipe with a new version
func (rc *RecipeCache) UpdateRecipe(ctx context.Context, tempID string, newVersion Recipe) error {
	// Get existing cache
	cache, err := rc.GetRecipe(ctx, tempID)
	if err != nil {
		return err
	}

	// Add current version to history
	cache.Versions = append(cache.Versions, cache.Current)
	cache.Current = newVersion
	cache.ModificationCount++
	cache.Timestamp = time.Now()

	// Convert to JSON
	cacheJSON, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal updated recipe cache: %v", err)
	}

	// Store back in Redis, maintaining the original TTL
	key := fmt.Sprintf("recipe:%s", tempID)
	if err := rc.client.Set(ctx, key, cacheJSON, time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to update recipe cache: %v", err)
	}

	return nil
}

// DeleteRecipe removes a recipe from Redis
func (rc *RecipeCache) DeleteRecipe(ctx context.Context, tempID string) error {
	key := fmt.Sprintf("recipe:%s", tempID)
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete recipe from cache: %v", err)
	}
	return nil
}
