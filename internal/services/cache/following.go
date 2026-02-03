package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"unipilot/internal/errors"
	"unipilot/internal/models"

	"gorm.io/datatypes"
)

// Following provides caching operations for user following resource.

// GetUserFollowing retrieves all following users for a user from cache (hash structure).
// Returns map of followedID -> User JSON, empty map on cache miss.
func (c *Cache) GetUserFollowing(ctx context.Context, userID datatypes.UUID) ([]models.User, error) {
	cacheKey := FormatKey(KeyUserFollowing, userID)
	followingHash, err := c.redis.HGetAll(ctx, cacheKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, fmt.Sprintf("Error getting user following from redis for user %d", userID))
	}
	var followings []models.User
	for _, followingJSON := range followingHash {
		var following models.User
		if err := json.Unmarshal([]byte(followingJSON), &following); err == nil {
			followings = append(followings, following)
		}
	}
	if len(followings) == 0 {
		return nil, errors.NewAppError(errors.CacheMiss, "No following found", nil)
	}
	return followings, nil
}

// SetUserFollowing stores a followed user in cache (hash structure).
func (c *Cache) SetUserFollowing(ctx context.Context, userID, followedID datatypes.UUID, followed *models.User) error {
	cacheKey := FormatKey(KeyUserFollowing, userID)
	followedJSON, err := json.Marshal(followed)
	if err != nil {
		return errors.Wrap(err, errors.ProcJSONMarshalFailed, "Error marshalling followed user to json")
	}
	return c.redis.HSet(ctx, cacheKey, followedID.String(), followedJSON).Err()
}

// DeleteUserFollowing removes a followed user from cache.
func (c *Cache) DeleteUserFollowing(ctx context.Context, userID, followedID datatypes.UUID) error {
	cacheKey := FormatKey(KeyUserFollowing, userID)
	return c.redis.HDel(ctx, cacheKey, followedID.String()).Err()
}

// SetExpirationUserFollowing sets TTL for the user following hash.
func (c *Cache) SetExpirationUserFollowing(ctx context.Context, userID datatypes.UUID) error {
	cacheKey := FormatKey(KeyUserFollowing, userID)
	return c.redis.Expire(ctx, cacheKey, TTLUserFollowing).Err()
}
