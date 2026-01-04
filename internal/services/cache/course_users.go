package cache

import (
	"context"
	"fmt"
	"strconv"

	"unipilot/internal/errors"

	"github.com/google/uuid"
)

// GetLinkUsers retrieves all user IDs sharing a course via LinkID from cache.
func (c *Cache) GetCourseUsers(ctx context.Context, courseID uint, excludeUserID uint) ([]uint, error) {
	cacheKey := FormatKey(KeyCoursesLinkedUsers, fmt.Sprintf("%d", courseID))

	// Get all user IDs from Redis Set
	userIDStrings, err := c.redis.SMembers(ctx, cacheKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error getting link users from cache")
	}

	// Convert to uint slice and exclude the specified user
	var userIDs []uint
	for _, idStr := range userIDStrings {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			continue
		}
		userID := uint(id)
		if userID != excludeUserID {
			userIDs = append(userIDs, userID)
		}
	}

	return userIDs, nil
}

// SetCourseUsers sets the course users in cache.
func (c *Cache) SetCourseUsers(ctx context.Context, courseID uint, userIDs []uint) error {
	cacheKey := FormatKey(KeyCoursesLinkedUsers, fmt.Sprintf("%d", courseID))
	if err := c.redis.SAdd(ctx, cacheKey, userIDs).Err(); err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Error setting course users in cache")
	}
	return c.redis.Expire(ctx, cacheKey, TTLLinkUsers).Err()
}

// AddLinkUser adds a user to a LinkID set (when course is linked).
// Optimized for write-heavy workflow: O(1) operation.
func (c *Cache) AddCourseUser(ctx context.Context, courseID uint, userID uint) error {
	cacheKey := FormatKey(KeyCoursesLinkedUsers, fmt.Sprintf("%d", courseID))

	// Add user ID to set
	if err := c.redis.SAdd(ctx, cacheKey, strconv.Itoa(int(userID))).Err(); err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Error adding user to link")
	}

	// Set expiration (long TTL since this is write-maintained)
	return c.redis.Expire(ctx, cacheKey, TTLLinkUsers).Err()
}

// RemoveUserFromCourse removes a user from a Course set (when course is unlinked).
func (c *Cache) RemoveCourseUser(ctx context.Context, courseID uint, userID uint) error {
	cacheKey := FormatKey(KeyCoursesLinkedUsers, fmt.Sprintf("%d", courseID))
	return c.redis.SRem(ctx, cacheKey, strconv.Itoa(int(userID))).Err()
}

// FormatLinkKey formats a cache key with LinkID (UUID).
func FormatLinkKey(pattern string, linkID uuid.UUID) string {
	return fmt.Sprintf(pattern, linkID.String())
}
