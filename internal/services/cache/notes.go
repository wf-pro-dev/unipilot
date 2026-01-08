package cache

import (
	"context"
	"encoding/json"
	"strconv"

	"unipilot/internal/errors"
	"unipilot/internal/models"

	"gorm.io/gorm"
)

func (c *Cache) GetNote(ctx context.Context, noteID uint) (*models.Note, error) {
	cacheKey := FormatKey(KeyNote, strconv.Itoa(int(noteID)))
	noteJSON, err := c.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error getting note from redis")
	}
	var note models.Note
	if err := json.Unmarshal([]byte(noteJSON), &note); err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error unmarshalling note from redis")
	}
	return &note, nil
}

func (c *Cache) SetNotes(ctx context.Context, notes []*models.Note) error {

	pipe := c.redis.Pipeline()
	for _, note := range notes {
		cacheKey := FormatKey(KeyNote, strconv.Itoa(int(note.ID)))
		noteJSON, err := json.Marshal(note)
		if err != nil {
			continue
		}
		pipe.Set(ctx, cacheKey, noteJSON, TTLNote)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Error setting notes in redis")
	}
	return nil
}

func (c *Cache) GetNotesByIDs(ctx context.Context, noteIDs []uint, db *gorm.DB) ([]*models.Note, error) {
	if len(noteIDs) == 0 {
		return []*models.Note{}, nil
	}

	keys := make([]string, len(noteIDs))
	for i, id := range noteIDs {
		keys[i] = FormatKey(KeyNote, strconv.Itoa(int(id)))
	}

	results, err := c.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error getting notes from redis")
	}

	notes := make([]*models.Note, 0, len(noteIDs))
	missingIDs := make([]uint, 0)

	for i, result := range results {
		if result == nil {
			missingIDs = append(missingIDs, noteIDs[i])
			continue
		}
		var note models.Note
		if err := json.Unmarshal([]byte(result.(string)), &note); err != nil {
			missingIDs = append(missingIDs, noteIDs[i])
			continue
		}
		notes = append(notes, &note)
	}
	if len(missingIDs) > 0 {
		missingNotes, err := models.GetNotesByIDs(missingIDs, db)
		if err != nil {
			return nil, errors.Wrap(err, errors.DBQueryFailed, "Error getting notes from database")
		}
		notes = append(notes, missingNotes...)

		go c.SetNotes(context.Background(), missingNotes)
	}
	return notes, nil
}

func (c *Cache) DeleteNote(ctx context.Context, noteID uint) error {
	return c.redis.Del(ctx, FormatKey(KeyNote, strconv.Itoa(int(noteID)))).Err()
}

func (c *Cache) SetExpirationNote(ctx context.Context, noteID uint) error {
	return c.redis.Expire(ctx, FormatKey(KeyNote, strconv.Itoa(int(noteID))), TTLNote).Err()
}
