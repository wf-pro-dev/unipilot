package cache

import (
	"context"
	"encoding/json"

	"unipilot/internal/errors"
	"unipilot/internal/models"
)

// LinkedCourses provides caching operations for user linked courses resource.
// read-heavy complex query with nested preloads.

// GetUserLinkedCourses retrieves user's linked courses from cache.
// Returns nil on cache miss (not an error).
func (c *Cache) GetCoursesLinked(ctx context.Context, userID uint) ([]models.Course, error) {
	cacheKey := FormatKey(KeyCoursesLinked, userID)
	cachedData, err := c.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err // Cache miss
	}

	var courses []models.Course
	if err := json.Unmarshal([]byte(cachedData), &courses); err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Invalid cache data for user linked courses")
	}
	return courses, nil
}

// SetUserLinkedCourses stores user's linked courses in cache.
func (c *Cache) SetCoursesLinked(ctx context.Context, userID uint, courses []models.Course) error {
	cacheKey := FormatKey(KeyCoursesLinked, userID)
	coursesJSON, err := json.Marshal(courses)
	if err != nil {
		return errors.Wrap(err, errors.ProcJSONMarshalFailed, "Error marshalling user linked courses")
	}
	return c.redis.Set(ctx, cacheKey, coursesJSON, TTLUserLinkedCourses).Err()
}

// DeleteUserLinkedCourses invalidates user's linked courses cache.
func (c *Cache) DeleteCoursesLinked(ctx context.Context, userID uint) error {
	cacheKey := FormatKey(KeyCoursesLinked, userID)
	return c.redis.Del(ctx, cacheKey).Err()
}

func (c *Cache) SetExpirationCoursesLinked(ctx context.Context, userID uint) error {
	cacheKey := FormatKey(KeyCoursesLinked, userID)
	return c.redis.Expire(ctx, cacheKey, TTLUserLinkedCourses).Err()
}
