package repositories

import (
	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new RedisClient
func NewRedisClient(client *redis.Client) *RedisClient {
	return &RedisClient{
		client: client,
	}
}

// GetClient returns the underlying Redis client
func (c *RedisClient) GetClient() *redis.Client {
	return c.client
}
