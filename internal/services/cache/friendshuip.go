package cache

import (
	"context"
)

func (c *Cache) GetUserFriends(ctx context.Context, userID string) ([]string, error) {
	cacheKey := FormatKey(KeyUserFriends, userID)
	return c.redis.SMembers(ctx, cacheKey).Result()
}

func (c *Cache) SetUserFriends(ctx context.Context, userID string, friendIDs []string) error {
	pipe := c.redis.Pipeline()
	for _, friendID := range friendIDs {
		pipe.SAdd(ctx, FormatKey(KeyUserFriends, userID), friendID).Err()
	}
	pipe.Expire(ctx, FormatKey(KeyUserFriends, userID), TTLUserFriends)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *Cache) AddUserFriend(ctx context.Context, userID string, friendID string) error {
	cacheKey := FormatKey(KeyUserFriends, userID)
	c.redis.SAdd(ctx, cacheKey, friendID).Err()
	return c.SetExpirationUserFriends(ctx, userID)
}

func (c *Cache) RemoveUserFriend(ctx context.Context, userID string, friendID string) error {
	return c.redis.SRem(ctx, FormatKey(KeyUserFriends, userID), friendID).Err()
}

func (c *Cache) DeleteUserFriends(ctx context.Context, userID string) error {
	return c.redis.Del(ctx, FormatKey(KeyUserFriends, userID)).Err()
}

func (c *Cache) SetExpirationUserFriends(ctx context.Context, userID string) error {
	return c.redis.Expire(ctx, FormatKey(KeyUserFriends, userID), TTLUserFriends).Err()
}
