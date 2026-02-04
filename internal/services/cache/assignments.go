package cache

import (
	"context"
	"encoding/json"

	"unipilot/internal/errors"
	"unipilot/internal/models"

	"gorm.io/gorm"
)

func (c *Cache) SetAssignments(ctx context.Context, assignments []*models.Assignment) error {
	if len(assignments) == 0 {
		return nil
	}

	pipe := c.redis.Pipeline()
	for _, assignment := range assignments {
		cacheKey := FormatKey(KeyAssignment, assignment.ID)
		assignmentJSON, err := json.Marshal(assignment)
		if err != nil {
			continue
		}
		pipe.Set(ctx, cacheKey, assignmentJSON, TTLAssignment)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Error setting assignments in redis")
	}
	return nil
}

func (c *Cache) GetAssignmentsByIDs(ctx context.Context, assignmentIDs []string, db *gorm.DB) ([]*models.Assignment, error) {

	if len(assignmentIDs) == 0 {
		return []*models.Assignment{}, nil
	}

	// Build keys: ["course:1", "course:5", "course:10"]
	keys := make([]string, len(assignmentIDs))
	for i, id := range assignmentIDs {
		keys[i] = FormatKey(KeyAssignment, id)
	}

	// Fetch all courses in a single round trip
	results, err := c.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error getting assignments from redis")
	}

	assignments := make([]*models.Assignment, 0, len(assignmentIDs))
	missingIDs := make([]string, 0)

	// Parse results
	for i, result := range results {
		if result == nil {
			// Cache miss - this assignment wasn't found
			missingIDs = append(missingIDs, assignmentIDs[i])
			continue
		}

		var assignment models.Assignment
		if err := json.Unmarshal([]byte(result.(string)), &assignment); err != nil {
			missingIDs = append(missingIDs, assignmentIDs[i])
			continue
		}
		assignments = append(assignments, &assignment)
	}

	if len(missingIDs) > 0 {
		//
		missingAssignments, err := models.GetAssignmentsByIDs(missingIDs, db)
		if err != nil {
			return nil, errors.Wrap(err, errors.DBQueryFailed, "Error getting assignments from database")
		}

		go c.SetAssignments(context.Background(), missingAssignments)

		assignments = append(assignments, missingAssignments...)
	}

	return assignments, nil

}
func (c *Cache) DeleteAssignment(ctx context.Context, assignmentID string) error {
	return c.redis.Del(ctx, FormatKey(KeyAssignment, assignmentID)).Err()
}

func (c *Cache) SetExpirationAssignment(ctx context.Context, assignmentID string) error {
	return c.redis.Expire(ctx, FormatKey(KeyAssignment, assignmentID), TTLAssignment).Err()
}
