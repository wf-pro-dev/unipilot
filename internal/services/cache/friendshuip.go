package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"

	"unipilot/internal/models"
)

// GetUserFriends retrieves cached friends for a user
func (c *Cache) GetUserFriends(ctx context.Context, userID datatypes.UUID) ([]models.User, error) {
	key := fmt.Sprintf("friends:%s", userID.String())

	// Get all friends from hash
	friendsMap, err := c.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(friendsMap) == 0 {
		return nil, fmt.Errorf("cache miss")
	}

	friends := make([]models.User, 0, len(friendsMap))
	for _, friendJSON := range friendsMap {
		var user models.User
		if err := json.Unmarshal([]byte(friendJSON), &user); err != nil {
			continue // Skip invalid entries
		}
		friends = append(friends, user)
	}

	return friends, nil
}

// SetUserFriends caches a single friend for a user
func (c *Cache) SetUserFriends(ctx context.Context, userID, friendID datatypes.UUID, user *models.User) error {
	key := fmt.Sprintf("friends:%s", userID.String())
	field := friendID.String()

	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return c.redis.HSet(ctx, key, field, userJSON).Err()
}

// DeleteUserFriends clears the entire friends cache for a user
func (c *Cache) DeleteUserFriends(ctx context.Context, userID datatypes.UUID) error {
	key := fmt.Sprintf("friends:%s", userID.String())
	return c.redis.Del(ctx, key).Err()
}

// SetExpirationUserFriends sets TTL on friends cache
func (c *Cache) SetExpirationUserFriends(ctx context.Context, userID datatypes.UUID) error {
	key := fmt.Sprintf("friends:%s", userID.String())
	return c.redis.Expire(ctx, key, 30*time.Minute).Err()
}
