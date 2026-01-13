package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/server"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func (c *Cache) SetCourses(ctx context.Context, courses []*models.Course) error {
	pipe := c.redis.Pipeline()

	for _, course := range courses {

		if course.Assignments != nil {
			originalAssignmentIDs := make([]interface{}, 0)

			for _, assignment := range course.Assignments {
				if assignment.ParentID == 0 {
					assignmentKey := FormatKey(KeyAssignment, strconv.Itoa(int(assignment.ID)))
					assignmentJSON, err := json.Marshal(&assignment)
					if err != nil {
						continue
					}
					pipe.Set(ctx, assignmentKey, assignmentJSON, TTLAssignment)
					originalAssignmentIDs = append(originalAssignmentIDs, assignment.ID)
				}
			}

			// Store only original IDs
			if len(originalAssignmentIDs) > 0 {
				assignmentsKey := FormatKey(KeyCourseAssignments, strconv.Itoa(int(course.ID)))
				pipe.Del(ctx, assignmentsKey)
				pipe.SAdd(ctx, assignmentsKey, originalAssignmentIDs...)
				pipe.Expire(ctx, assignmentsKey, TTLCourseAssignments)
			}
		}

		if course.Notes != nil {
			originalNoteIDs := make([]interface{}, 0)

			for _, note := range course.Notes {
				if note.ParentID == 0 {
					note.Content = ""
					noteKey := FormatKey(KeyNote, strconv.Itoa(int(note.ID)))
					noteJSON, err := json.Marshal(&note)
					if err != nil {
						continue
					}
					pipe.Set(ctx, noteKey, noteJSON, TTLNote)
					originalNoteIDs = append(originalNoteIDs, note.ID)
				}
			}

			if len(originalNoteIDs) > 0 {
				notesKey := FormatKey(KeyCourseNotes, strconv.Itoa(int(course.ID)))
				pipe.Del(ctx, notesKey)
				pipe.SAdd(ctx, notesKey, originalNoteIDs...)
				pipe.Expire(ctx, notesKey, TTLCourseNotes)
			}
		}

		// Clean dependencies
		course.Assignments = nil
		course.Notes = nil

		cacheKey := FormatKey(KeyCourse, strconv.Itoa(int(course.ID)))
		courseJSON, err := json.Marshal(course)
		if err != nil {
			continue
		}
		pipe.Set(ctx, cacheKey, courseJSON, TTLCourse)

		pipe.Expire(ctx, cacheKey, TTLCourse)
	}

	// Execute all commands in one round trip
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, errors.CacheOperationFailed, "Error setting courses in redis")
	}

	return nil
}

func (c *Cache) DeleteCourse(ctx context.Context, courseID uint) error {
	return c.redis.Del(ctx, FormatKey(KeyCourse, strconv.Itoa(int(courseID)))).Err()
}

// GetClusterCourses retrieves all Course IDs in a cluster by the Root ID
func (c *Cache) GetClusterCourses(ctx context.Context, rootID uint, db *gorm.DB) ([]uint, error) {
	cacheKey := FormatKey(KeyClusterCourses, strconv.Itoa(int(rootID)))

	// 1. Fetch from Redis
	result, err := c.redis.SMembers(ctx, cacheKey).Result()
	if err != nil && err != redis.Nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "failed to get cluster courses")
	}

	// 2. Cache Hit: Return the Course IDs
	if len(result) > 0 {
		return parseUintSlice(result), nil
	}

	// 3. Cache Miss: Fallback to DB
	courseIDs, err := models.GetClusterCourses(rootID, db)
	if err != nil {
		return nil, err
	}

	// 4. Warm Cache if courses exist
	if len(courseIDs) > 0 {
		go c.SetClusterCourses(context.Background(), rootID, courseIDs)
	}

	return courseIDs, nil
}

// SetClusterCourses warms the cluster cache
func (c *Cache) SetClusterCourses(ctx context.Context, rootID uint, courseIDs []uint) error {
	cacheKey := FormatKey(KeyClusterCourses, strconv.Itoa(int(rootID)))

	interfaces := make([]interface{}, len(courseIDs))
	for i, v := range courseIDs {
		interfaces[i] = v
	}

	pipe := c.redis.Pipeline()
	pipe.SAdd(ctx, cacheKey, interfaces...)
	pipe.Expire(ctx, cacheKey, TTLCourseLinks)
	_, err := pipe.Exec(ctx)
	return err
}

// AddCourseToCluster adds a single Course ID to a cluster (used on Link Accept)
func (c *Cache) AddClusterCourse(ctx context.Context, rootID uint, courseID uint) error {
	cacheKey := FormatKey(KeyClusterCourses, strconv.Itoa(int(rootID)))
	c.redis.SAdd(ctx, cacheKey, courseID).Err()
	return c.SetExpirationClusterCourses(ctx, rootID)
}

// RemoveCourseFromCluster removes a Course ID (used on Course Delete/Link Break)
func (c *Cache) RemoveClusterCourse(ctx context.Context, rootID uint, courseID uint) error {
	cacheKey := FormatKey(KeyClusterCourses, strconv.Itoa(int(rootID)))
	c.redis.SRem(ctx, cacheKey, courseID).Err()
	return c.SetExpirationClusterCourses(ctx, rootID)
}

func (c *Cache) SetExpirationClusterCourses(ctx context.Context, rootID uint) error {
	return c.redis.Expire(ctx, FormatKey(KeyClusterCourses, strconv.Itoa(int(rootID))), TTLCourseLinks).Err()
}

func (c *Cache) GetClusterUsers(ctx context.Context, rootID uint, db *gorm.DB) ([]uint, error) {
	cacheKey := FormatKey(KeyClusterUsers, strconv.Itoa(int(rootID)))

	// 1. Fetch from Redis
	result, err := c.redis.SMembers(ctx, cacheKey).Result()
	if err != nil && err != redis.Nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "failed to get cluster users")
	}

	// 2. Cache Hit: Return the User IDs
	if len(result) > 0 {
		return parseUintSlice(result), nil
	}

	// 3. Cache Miss: Fallback to DB
	userIDs, err := models.GetClusterUserIDs(rootID, db)
	if err != nil {
		return nil, err
	}

	// 4. Warm Cache if courses exist
	if len(userIDs) > 0 {
		go c.SetClusterUsers(context.Background(), rootID, userIDs)
	}
	return userIDs, nil
}

func (c *Cache) SetClusterUsers(ctx context.Context, rootID uint, userIDs []uint) error {
	cacheKey := FormatKey(KeyClusterUsers, strconv.Itoa(int(rootID)))
	interfaces := make([]interface{}, len(userIDs))
	for i, v := range userIDs {
		interfaces[i] = v
	}
	return c.redis.SAdd(ctx, cacheKey, interfaces...).Err()
}

func (c *Cache) AddClusterUser(ctx context.Context, rootID uint, userID uint) error {
	cacheKey := FormatKey(KeyClusterUsers, strconv.Itoa(int(rootID)))
	c.redis.SAdd(ctx, cacheKey, userID).Err()
	return c.SetExpirationClusterUsers(ctx, rootID)
}

func (c *Cache) RemoveClusterUser(ctx context.Context, rootID uint, userID uint) error {
	cacheKey := FormatKey(KeyClusterUsers, strconv.Itoa(int(rootID)))
	c.redis.SRem(ctx, cacheKey, userID).Err()
	return c.SetExpirationClusterUsers(ctx, rootID)
}

func (c *Cache) DeleteClusterUsers(ctx context.Context, rootID uint) error {
	return c.redis.Del(ctx, FormatKey(KeyClusterUsers, strconv.Itoa(int(rootID)))).Err()
}

func (c *Cache) SetExpirationClusterUsers(ctx context.Context, rootID uint) error {
	return c.redis.Expire(ctx, FormatKey(KeyClusterUsers, strconv.Itoa(int(rootID))), TTLClusterUsers).Err()
}

func (cache *Cache) GetCoursesByIDs(ctx context.Context, courseIDs []uint, db *gorm.DB) ([]*models.Course, error) {
	if len(courseIDs) == 0 {
		return []*models.Course{}, nil
	}

	// Step 1: Pipeline to fetch everything we need
	pipe := cache.redis.Pipeline()

	courseCmds := make([]*redis.StringCmd, len(courseIDs))
	assignmentCmds := make([]*redis.StringSliceCmd, len(courseIDs))
	noteCmds := make([]*redis.StringSliceCmd, len(courseIDs))

	for i, courseID := range courseIDs {
		courseCmds[i] = pipe.Get(ctx, FormatKey(KeyCourse, strconv.Itoa(int(courseID))))
		assignmentCmds[i] = pipe.SMembers(ctx, FormatKey(KeyCourseAssignments, strconv.Itoa(int(courseID))))
		noteCmds[i] = pipe.SMembers(ctx, FormatKey(KeyCourseNotes, strconv.Itoa(int(courseID))))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Real error - fail fast
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Pipeline failed")
	}

	// Step 2: Parse course objects and collect dependency IDs
	var (
		courses             []*models.Course
		missingCourseIDs    []uint
		allAssignmentIDs    []uint
		allNoteIDs          []uint
		courseAssignmentMap = make(map[uint][]uint)
		courseNoteMap       = make(map[uint][]uint)
	)

	expPipe := cache.redis.Pipeline()
	expCmds := make([]*redis.BoolCmd, 2*len(courseIDs))

	for i, courseID := range courseIDs {
		// Try to get course object
		courseJSON, err := courseCmds[i].Result()
		if err != nil {
			// Cache miss or error - need to fetch from DB
			missingCourseIDs = append(missingCourseIDs, courseID)

			// Log only real errors (not cache misses)
			if err != redis.Nil {
				server.LogWarn(ctx, errors.Wrap(err, errors.CacheOperationFailed,
					fmt.Sprintf("Redis error fetching course:%d", courseID)).ToServerError(fiber.StatusInternalServerError))
			}
			continue
		}

		// Unmarshal course
		var course models.Course
		if err := json.Unmarshal([]byte(courseJSON), &course); err != nil {
			// Corrupted cache - treat as miss
			missingCourseIDs = append(missingCourseIDs, courseID)
			server.LogWarn(ctx, errors.Wrap(err, errors.ProcJSONUnmarshalFailed,
				fmt.Sprintf("Corrupted cache data for course:%d", courseID)).ToServerError(fiber.StatusInternalServerError))
			continue
		}

		// Course found - collect its dependency IDs
		// Note: SMEMBERS never returns error for missing sets, just empty slice
		assignmentIDStrs, _ := assignmentCmds[i].Result()
		noteIDStrs, _ := noteCmds[i].Result()

		// Store assignment IDs
		assignmentIDs := parseUintSlice(assignmentIDStrs)
		courseAssignmentMap[courseID] = assignmentIDs
		allAssignmentIDs = append(allAssignmentIDs, assignmentIDs...)

		expCmds[i] = expPipe.Expire(ctx, FormatKey(KeyCourseAssignments, strconv.Itoa(int(courseID))), TTLCourseAssignments)

		// Store note IDs
		noteIDs := parseUintSlice(noteIDStrs)
		courseNoteMap[courseID] = noteIDs
		allNoteIDs = append(allNoteIDs, noteIDs...)

		expCmds[i+len(courseIDs)] = expPipe.Expire(ctx, FormatKey(KeyCourseNotes, strconv.Itoa(int(courseID))), TTLCourseNotes)

		courses = append(courses, &course)
	}

	_, err = expPipe.Exec(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error setting expiration for courses")
	}

	// Step 3: Batch fetch assignment objects
	assignments, err := cache.GetAssignmentsByIDs(ctx, allAssignmentIDs, db)
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error getting assignments from redis")
	}

	// Step 4: Batch fetch note objects
	notes, err := cache.GetNotesByIDs(ctx, allNoteIDs, db)
	if err != nil {
		return nil, errors.Wrap(err, errors.CacheOperationFailed, "Error getting notes from redis")
	}

	assignmentsMap := buildMap(assignments, func(a *models.Assignment) uint { return a.ID })
	notesMap := buildMap(notes, func(n *models.Note) uint { return n.ID })

	// Step 5: Attach dependencies to courses
	for _, course := range courses {
		course.Assignments = buildList(courseAssignmentMap[course.ID], assignmentsMap)
		course.Notes = buildList(courseNoteMap[course.ID], notesMap)
	}

	if len(missingCourseIDs) > 0 {
		missingCourses, err := models.GetCoursesByIDs(missingCourseIDs,
			db.
				Preload("User", func(db *gorm.DB) *gorm.DB {
					return db.Select("id, username, avatar, email")
				}).
				Preload("Assignments", "parent_id IS NULL").
				Preload("Notes", "parent_id IS NULL"))
		if err != nil {
			return nil, errors.Wrap(err, errors.DBQueryFailed, "Error getting courses from database")
		}
		go cache.SetCourses(context.Background(), missingCourses)

		courses = append(courses, missingCourses...)
	}

	return courses, nil
}

func parseUintSlice(strs []string) []uint {
	ids := make([]uint, len(strs))
	for i, s := range strs {
		id, _ := strconv.ParseUint(s, 10, 64) // "10" → 10
		ids[i] = uint(id)
	}
	return ids
}

func buildList[T any](ids []uint, items map[uint]*T) []*T {
	itemsList := make([]*T, 0, len(ids))
	for _, id := range ids {
		if item, found := items[id]; found {
			itemsList = append(itemsList, item)
		}
	}
	return itemsList
}

// Helper for map building (can be generic!)
func buildMap[T any, K comparable](items []*T, keyFunc func(*T) K) map[K]*T {
	m := make(map[K]*T, len(items))
	for _, item := range items {
		m[keyFunc(item)] = item
	}
	return m
}

func (c *Cache) SetExpirationCourse(ctx context.Context, courseID uint) error {
	return c.redis.Expire(ctx, FormatKey(KeyCourse, strconv.Itoa(int(courseID))), TTLCourse).Err()
}

func (c *Cache) AddCourseAssignment(ctx context.Context, courseID uint, assignmentID uint) error {
	cacheKey := FormatKey(KeyCourseAssignments, strconv.Itoa(int(courseID)))
	c.redis.SAdd(ctx, cacheKey, assignmentID).Err()
	return c.SetExpirationCourseAssignments(ctx, courseID)
}

func (c *Cache) RemoveCourseAssignment(ctx context.Context, courseID uint, assignmentID uint) error {
	return c.redis.SRem(ctx, FormatKey(KeyCourseAssignments, strconv.Itoa(int(courseID))), strconv.Itoa(int(assignmentID))).Err()
}

func (c *Cache) DeleteCourseAssignments(ctx context.Context, courseID uint) error {
	return c.redis.Del(ctx, FormatKey(KeyCourseAssignments, strconv.Itoa(int(courseID)))).Err()
}

func (c *Cache) SetExpirationCourseAssignments(ctx context.Context, courseID uint) error {
	return c.redis.Expire(ctx, FormatKey(KeyCourseAssignments, strconv.Itoa(int(courseID))), TTLCourseAssignments).Err()
}

func (c *Cache) AddCourseNote(ctx context.Context, courseID uint, noteID uint) error {
	return c.redis.SAdd(ctx, FormatKey(KeyCourseNotes, strconv.Itoa(int(courseID))), strconv.Itoa(int(noteID))).Err()
}

func (c *Cache) RemoveCourseNote(ctx context.Context, courseID uint, noteID uint) error {
	return c.redis.SRem(ctx, FormatKey(KeyCourseNotes, strconv.Itoa(int(courseID))), strconv.Itoa(int(noteID))).Err()
}

func (c *Cache) DeleteCourseNotes(ctx context.Context, courseID uint) error {
	return c.redis.Del(ctx, FormatKey(KeyCourseNotes, strconv.Itoa(int(courseID)))).Err()
}

func (c *Cache) SetExpirationCourseNotes(ctx context.Context, courseID uint) error {
	return c.redis.Expire(ctx, FormatKey(KeyCourseNotes, strconv.Itoa(int(courseID))), TTLCourseNotes).Err()
}
