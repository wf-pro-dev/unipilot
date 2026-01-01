package cache

import (
	"context"
	"encoding/json"
	"strconv"

	"unipilot/internal/errors"
	"unipilot/internal/models"
)

// Users provides caching operations for users resource.

// GetUsers retrieves all users from cache (hash structure).
// Returns map of userID -> User JSON, empty map on cache miss.
func (c *Cache) GetUsers(ctx context.Context) (map[string]string, error) {
	usersHash, err := c.redis.HGetAll(ctx, KeyUsers).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error getting users from redis")
	}
	return usersHash, nil
}

// SetUsers stores a user in cache (hash structure).
func (c *Cache) SetUsers(ctx context.Context, userID uint, user *models.User) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return errors.Wrap(err, errors.ProcJSONMarshalFailed, "Error marshalling user to json")
	}
	return c.redis.HSet(ctx, KeyUsers, strconv.Itoa(int(userID)), userJSON).Err()
}

// DeleteUsers removes a user from cache.
func (c *Cache) DeleteUsers(ctx context.Context, userID uint) error {
	return c.redis.HDel(ctx, KeyUsers, strconv.Itoa(int(userID))).Err()
}

// SetExpirationUsers sets TTL for the entire users hash.
func (c *Cache) SetExpirationUsers(ctx context.Context) error {
	return c.redis.Expire(ctx, KeyUsers, TTLUsers).Err()
}
