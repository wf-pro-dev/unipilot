package cache

import (
	"context"
	"encoding/json"

	"unipilot/internal/errors"
	"unipilot/internal/services/fileops/progress"

	"github.com/redis/go-redis/v9"
)

// Set initializes or updates progress for an upload
func (c *Cache) SetProgress(ctx context.Context, uploadID string, snapshot *progress.TrackerSnapshot) error {
	key := FormatKey(KeyProgress, uploadID)

	data, err := json.Marshal(snapshot)
	if err != nil {
		return errors.Wrap(err, errors.ProcJSONMarshalFailed, "Failed to marshal progress data")
	}

	if err := c.redis.Set(ctx, key, data, TTLProgress).Err(); err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Failed to set progress in cache")
	}

	return nil
}

// Get retrieves progress for an upload
func (c *Cache) GetProgress(ctx context.Context, uploadID string) (*progress.TrackerSnapshot, error) {
	key := FormatKey(KeyProgress, uploadID)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.Wrap(err, errors.CacheOperationFailed, "Progress not found")
		}
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Failed to get progress from cache")
	}

	var progress progress.TrackerSnapshot
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal progress data")
	}

	return &progress, nil
}

// UpdatePercentage updates just the percentage and status
func (c *Cache) UpdatePercentage(ctx context.Context, uploadID string, percentage float64) error {
	// Get existing progress
	progress, err := c.GetProgress(ctx, uploadID)
	if err != nil {
		// If not found, create new
		return errors.Wrap(err, errors.CacheOperationFailed, "Progress not found")
	} else {
		progress.Percentage = percentage
	}

	return c.SetProgress(ctx, uploadID, progress)
}

// SetStatus updates only the status
func (c *Cache) SetStatus(ctx context.Context, uploadID, status, message string) error {
	progress, err := c.GetProgress(ctx, uploadID)
	if err != nil {
		// If not found, create new with minimal data
		return errors.Wrap(err, errors.CacheOperationFailed, "Progress not found")
	} else {
		progress.Status = status
	}

	return c.SetProgress(ctx, uploadID, progress)
}

// Delete removes progress entry
func (c *Cache) Delete(ctx context.Context, uploadID string) error {
	key := FormatKey(KeyProgress, uploadID)
	return c.redis.Del(ctx, key).Err()
}

// Exists checks if progress entry exists
func (c *Cache) Exists(ctx context.Context, uploadID string) (bool, error) {
	key := FormatKey(KeyProgress, uploadID)
	exists, err := c.redis.Exists(ctx, key).Result()
	return exists > 0, err
}
