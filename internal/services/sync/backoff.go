package sync

import (
	"time"

	"unipilot/internal/errors"
)

// ShouldRetry determines if an operation should be retried based on error type
func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a network/connection error
	if errors.HasCode(err, errors.NetworkOffline) {
		return true
	}

	// Check if it's a sync-related error that can be retried
	if errors.HasCode(err, errors.SyncFailed) {
		return true
	}

	// Check if it's a client request error (network issues)
	if errors.HasCode(err, errors.ClientRequestFailed) {
		return true
	}

	// For now, retry all errors (conservative approach)
	return true
}

// CalculateBackoffDuration calculates exponential backoff duration
func CalculateBackoffDuration(retryCount int, baseDelay, maxDelay time.Duration) time.Duration {
	delay := baseDelay * time.Duration(1<<retryCount)
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// GetBackoffDuration calculates exponential backoff using default values
func GetBackoffDuration(retryCount int) time.Duration {
	return CalculateBackoffDuration(retryCount, 1*time.Second, 5*time.Minute)
}
