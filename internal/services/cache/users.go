package cache

import (
	"context"
	"encoding/json"
	"strconv"

	"unipilot/internal/errors"
	"unipilot/internal/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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

func (c *Cache) GetUserClusterIDs(ctx context.Context, userID uint, db *gorm.DB) ([]uint, error) {

	cacheKey := FormatKey(KeyUserClusters, strconv.Itoa(int(userID)))

	// 1. Try to get from Redis
	result, err := c.redis.SMembers(ctx, cacheKey).Result()
	if err != nil && err != redis.Nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "failed to get anchors from redis")
	}

	// 2. If Cache Hit: Parse and return
	if len(result) > 0 {
		return parseUintSlice(result), nil
	}

	// 3. If Cache Miss: Fallback to Database
	clusterIDs, err := models.GetUserClusterIDs(userID, db)
	if err != nil {
		return nil, err
	}

	// 4. If we found anchors, warm the cache asynchronously
	if len(clusterIDs) > 0 {
		go func() {
			// Convert to interface slice for SAdd
			interfaces := make([]interface{}, len(clusterIDs))
			for i, v := range clusterIDs {
				interfaces[i] = v
			}
			c.redis.SAdd(context.Background(), cacheKey, interfaces...)
			c.redis.Expire(context.Background(), cacheKey, TTLUserCoursesLinked)
		}()
	}

	return clusterIDs, nil
}

func (c *Cache) AddUserCluster(ctx context.Context, userID uint, clusterID uint) error {
	cacheKey := FormatKey(KeyUserClusters, strconv.Itoa(int(userID)))
	c.redis.SAdd(ctx, cacheKey, clusterID).Err()
	return c.SetExpirationUserClusters(ctx, userID)
}

func (c *Cache) RemoveUserCluster(ctx context.Context, userID uint, clusterID uint) error {
	cacheKey := FormatKey(KeyUserClusters, strconv.Itoa(int(userID)))
	c.redis.SRem(ctx, cacheKey, clusterID).Err()
	return c.SetExpirationUserClusters(ctx, userID)
}

func (c *Cache) DeleteUserClusters(ctx context.Context, userID uint) error {
	return c.redis.Del(ctx, FormatKey(KeyUserClusters, strconv.Itoa(int(userID)))).Err()
}

func (c *Cache) SetExpirationUserClusters(ctx context.Context, userID uint) error {
	return c.redis.Expire(ctx, FormatKey(KeyUserClusters, strconv.Itoa(int(userID))), TTLUserCoursesLinked).Err()
}
