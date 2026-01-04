package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"unipilot/internal/errors"
	"unipilot/internal/models"
)

// Followers provides caching operations for user followers resource.

// GetUserFollowers retrieves all followers for a user from cache (hash structure).
// Returns map of followerID -> User JSON, empty map on cache miss.
func (c *Cache) GetUserFollowers(ctx context.Context, userID uint) ([]models.User, error) {
	cacheKey := FormatKey(KeyUserFollowers, strconv.Itoa(int(userID)))
	followersHash, err := c.redis.HGetAll(ctx, cacheKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, fmt.Sprintf("Error getting user followers from redis for user %d", userID))
	}
	var followers []models.User
	for _, followerJSON := range followersHash {
		var follower models.User
		if err := json.Unmarshal([]byte(followerJSON), &follower); err == nil {
			followers = append(followers, follower)
		}
	}
	if len(followers) == 0 {
		return nil, errors.NewAppError(errors.CacheMiss, "No followers found", nil)
	}

	return followers, nil
}

// SetUserFollowers stores a follower in cache (hash structure).
func (c *Cache) SetUserFollowers(ctx context.Context, userID, followerID uint, follower *models.User) error {
	cacheKey := FormatKey(KeyUserFollowers, strconv.Itoa(int(userID)))
	followerJSON, err := json.Marshal(follower)
	if err != nil {
		return errors.Wrap(err, errors.ProcJSONMarshalFailed, "Error marshalling follower to json")
	}
	return c.redis.HSet(ctx, cacheKey, strconv.Itoa(int(followerID)), followerJSON).Err()
}

// DeleteUserFollowers removes a follower from cache.
func (c *Cache) DeleteUserFollowers(ctx context.Context, userID, followerID uint) error {
	cacheKey := FormatKey(KeyUserFollowers, strconv.Itoa(int(userID)))
	return c.redis.HDel(ctx, cacheKey, strconv.Itoa(int(followerID))).Err()
}

// SetExpirationUserFollowers sets TTL for the user followers hash.
func (c *Cache) SetExpirationUserFollowers(ctx context.Context, userID uint) error {
	cacheKey := FormatKey(KeyUserFollowers, strconv.Itoa(int(userID)))
	return c.redis.Expire(ctx, cacheKey, TTLUserFollowers).Err()
}
