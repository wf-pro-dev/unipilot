package cache

import (
	"github.com/redis/go-redis/v9"
)

// Cache provides centralized Redis caching for all resources.
// Resources are organized by namespace for clear separation of concerns.
type Cache struct {
	redis *redis.Client
}

// New creates a new centralized cache service.
func New(redisClient *redis.Client) *Cache {
	return &Cache{
		redis: redisClient,
	}
}
