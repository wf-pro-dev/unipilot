package cache

import (
	"context"
	"encoding/json"

	"unipilot/internal/errors"
	"unipilot/internal/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Users provides caching operations for users resource.

func (c *Cache) GetUsersByIDs(ctx context.Context, userIDs []string, cursor *models.Cursor, limit int, db *gorm.DB) (*models.PageResponse[models.User], error) {

	if len(userIDs) == 0 {
		return nil, nil
	}

	// Build keys: ["course:1", "course:5", "course:10"]
	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = FormatKey(KeyUser, id)
	}

	// Fetch all courses in a single round trip
	results, err := c.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error getting users from redis")
	}

	users := make([]*models.User, 0, len(userIDs))
	missingIDs := make([]string, 0)

	// Parse results
	for i, result := range results {
		if result == nil {
			// Cache miss - this assignment wasn't found
			missingIDs = append(missingIDs, userIDs[i])
			continue
		}

		var user models.User
		if err := json.Unmarshal([]byte(result.(string)), &user); err != nil {
			missingIDs = append(missingIDs, userIDs[i])
			continue
		}
		users = append(users, &user)
	}

	if len(missingIDs) > 0 {
		//
		missingUsers, err := models.GetUsersByIDs(missingIDs, db)
		if err != nil {
			return nil, errors.Wrap(err, errors.DBQueryFailed, "Error getting assignments from database")
		}

		go c.SetUsers(context.Background(), missingUsers)

		users = append(users, missingUsers...)
	}

	return &models.PageResponse[models.User]{
		Data:    users,
		Cursor:  cursor,
		HasMore: len(userIDs) > limit,
	}, nil

}

// Set SetUsers stores a user in cache (hash structure).
func (c *Cache) SetUsers(ctx context.Context, users []*models.User) error {
	pipe := c.redis.Pipeline()
	for _, user := range users {
		cacheKey := FormatKey(KeyUser, user.ID)
		userJSON, err := json.Marshal(user)
		if err != nil {
			continue
		}
		pipe.Set(ctx, cacheKey, userJSON, TTLUser)

		pipe.Expire(ctx, cacheKey, TTLUser)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Error setting users in redis")
	}
	return nil
}

func (c *Cache) GetUserClusterIDs(ctx context.Context, userID string, db *gorm.DB) ([]string, error) {

	cacheKey := FormatKey(KeyUserClusters, userID)

	// 1. Try to get from Redis
	result, err := c.redis.SMembers(ctx, cacheKey).Result()
	if err != nil && err != redis.Nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "failed to get anchors from redis")
	}

	// 2. If Cache Hit: Parse and return
	if len(result) > 0 {
		return result, nil
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

func (c *Cache) AddUserCluster(ctx context.Context, userID string, clusterID string) error {
	cacheKey := FormatKey(KeyUserClusters, userID)
	c.redis.SAdd(ctx, cacheKey, clusterID).Err()
	return c.SetExpirationUserClusters(ctx, userID)
}

func (c *Cache) RemoveUserCluster(ctx context.Context, userID string, clusterID string) error {
	cacheKey := FormatKey(KeyUserClusters, userID)
	c.redis.SRem(ctx, cacheKey, clusterID).Err()
	return c.SetExpirationUserClusters(ctx, userID)
}

func (c *Cache) DeleteUserClusters(ctx context.Context, userID string) error {
	return c.redis.Del(ctx, FormatKey(KeyUserClusters, userID)).Err()
}

func (c *Cache) SetExpirationUserClusters(ctx context.Context, userID string) error {
	return c.redis.Expire(ctx, FormatKey(KeyUserClusters, userID), TTLUserCoursesLinked).Err()
}
