package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"unipilot/internal/errors"
	"unipilot/internal/services/fileops/progress"

	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
)

// Set initializes or updates progress for an upload
func (c *Cache) SetProgress(ctx context.Context, progressID datatypes.UUID, snapshot *progress.TrackerSnapshot) error {
	key := FormatKey(KeyProgress, progressID)

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
func (c *Cache) GetProgressChannel(ctx context.Context, progressID datatypes.UUID) (*redis.PubSub, error) {
	key := FormatKey(KeyProgress, progressID)

	pubsub := c.redis.Subscribe(ctx, key)
	if pubsub == nil {
		return nil, errors.Wrap(
			fmt.Errorf("failed to subscribe to progress channel %s", key),
			errors.CacheOperationFailed,
			"Failed to subscribe to progress",
		)
	}

	// Wait for confirmation that subscription is created
	_, err := pubsub.Receive(ctx)
	if err != nil {
		pubsub.Close()
		return nil, errors.Wrap(
			fmt.Errorf("failed to receive subscription confirmation: %v", err),
			errors.CacheOperationFailed,
			"Failed to establish progress subscription",
		)
	}

	return pubsub, nil
}

func (c *Cache) PublishProgress(ctx context.Context, progressID datatypes.UUID, progress *progress.TrackerSnapshot) error {

	key := FormatKey(KeyProgress, progressID)
	data, err := json.Marshal(progress)
	if err != nil {
		return errors.Wrap(err, errors.ProcJSONMarshalFailed, "Failed to marshal progress data")
	}
	if err := c.redis.Publish(ctx, key, data).Err(); err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Failed to publish progress")
	}

	return nil
}

// Exists checks if progress entry exists
func (c *Cache) ProcessExists(ctx context.Context, progressID datatypes.UUID) (bool, error) {
	key := FormatKey(KeyProgress, progressID)
	exists, err := c.redis.Exists(ctx, key).Result()
	return exists > 0, err
}
