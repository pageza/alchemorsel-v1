package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RecipeCache struct {
	Original          Recipe    `json:"original"`
	Versions          []Recipe  `json:"versions"`
	Current           Recipe    `json:"current"`
	ModificationCount int       `json:"modification_count"`
	Timestamp         time.Time `json:"timestamp"`
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisClient{client: client}, nil
}

func (rc *RedisClient) CacheRecipe(ctx context.Context, recipe Recipe) (string, error) {
	// Generate a temporary ID for the recipe
	tempID := uuid.New().String()

	// Create the cache structure
	cache := RecipeCache{
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

func (rc *RedisClient) GetRecipe(ctx context.Context, tempID string) (*RecipeCache, error) {
	key := fmt.Sprintf("recipe:%s", tempID)

	// Get from Redis
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("recipe not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get recipe from cache: %v", err)
	}

	// Parse JSON
	var cache RecipeCache
	if err := json.Unmarshal([]byte(val), &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipe cache: %v", err)
	}

	return &cache, nil
}

func (rc *RedisClient) UpdateRecipe(ctx context.Context, tempID string, newVersion Recipe) error {
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
