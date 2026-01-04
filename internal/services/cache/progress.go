package cache

import (
	"context"
	"encoding/json"
	"time"

	"unipilot/internal/errors"

	"github.com/redis/go-redis/v9"
)

type UploadProgress struct {
	UploadID   string    `json:"upload_id"`
	Percentage float64   `json:"percentage"`
	Status     string    `json:"status"` // "starting", "uploading", "processing", "complete", "error"
	Message    string    `json:"message,omitempty"`
	TotalBytes int64     `json:"total_bytes"`
	SentBytes  int64     `json:"sent_bytes"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// Set initializes or updates progress for an upload
func (c *Cache) Set(ctx context.Context, uploadID string, progress *UploadProgress) error {
	key := FormatKey(KeyProgress, uploadID)

	// Ensure timestamps are set
	if progress.CreatedAt.IsZero() {
		progress.CreatedAt = time.Now()
	}
	progress.UpdatedAt = time.Now()

	data, err := json.Marshal(progress)
	if err != nil {
		return errors.Wrap(err, errors.ProcJSONMarshalFailed, "Failed to marshal progress data")
	}

	if err := c.redis.Set(ctx, key, data, TTLProgress).Err(); err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Failed to set progress in cache")
	}

	return nil
}

// Get retrieves progress for an upload
func (c *Cache) Get(ctx context.Context, uploadID string) (*UploadProgress, error) {
	key := FormatKey(KeyProgress, uploadID)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.Wrap(err, errors.CacheOperationFailed, "Progress not found")
		}
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Failed to get progress from cache")
	}

	var progress UploadProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal progress data")
	}

	return &progress, nil
}

// UpdatePercentage updates just the percentage and status
func (c *Cache) UpdatePercentage(ctx context.Context, uploadID string, percentage float64, status string) error {
	// Get existing progress
	progress, err := c.Get(ctx, uploadID)
	if err != nil {
		// If not found, create new
		progress = &UploadProgress{
			UploadID:   uploadID,
			Percentage: percentage,
			Status:     status,
			UpdatedAt:  time.Now(),
			CreatedAt:  time.Now(),
		}
	} else {
		progress.Percentage = percentage
		progress.Status = status
		progress.UpdatedAt = time.Now()
	}

	return c.Set(ctx, uploadID, progress)
}

// UpdateBytes updates the bytes sent
func (c *Cache) UpdateBytes(ctx context.Context, uploadID string, sentBytes, totalBytes int64) error {
	progress, err := c.Get(ctx, uploadID)
	if err != nil {
		// If not found, create new
		progress = &UploadProgress{
			UploadID:   uploadID,
			SentBytes:  sentBytes,
			TotalBytes: totalBytes,
			Percentage: float64(sentBytes) / float64(totalBytes) * 100,
			Status:     "uploading",
			UpdatedAt:  time.Now(),
			CreatedAt:  time.Now(),
		}
	} else {
		progress.SentBytes = sentBytes
		progress.TotalBytes = totalBytes
		progress.Percentage = float64(sentBytes) / float64(totalBytes) * 100
		progress.Status = "uploading"
		progress.UpdatedAt = time.Now()
	}

	return c.Set(ctx, uploadID, progress)
}

// SetStatus updates only the status
func (c *Cache) SetStatus(ctx context.Context, uploadID, status, message string) error {
	progress, err := c.Get(ctx, uploadID)
	if err != nil {
		// If not found, create new with minimal data
		progress = &UploadProgress{
			UploadID:  uploadID,
			Status:    status,
			Message:   message,
			UpdatedAt: time.Now(),
			CreatedAt: time.Now(),
		}
	} else {
		progress.Status = status
		progress.Message = message
		progress.UpdatedAt = time.Now()
	}

	return c.Set(ctx, uploadID, progress)
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
