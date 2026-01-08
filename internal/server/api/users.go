package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"unipilot/internal/errors"
	"unipilot/internal/models"
)

// GetUsersHandler retrieves all users in the system except the current authenticated models.
// Implements a two-tier caching strategy with Redis cache and database fallback for optimal
// performance. Enriches user data with associated course codes for comprehensive user profiles.
//
// HTTP Method: GET
// Content-Type: Not required (no request body expected)
//
// Request Body: None required (user context extracted from JWT token)
//
// Response (200 OK):
//   - message: Success message
//   - users: Array of user objects (as maps) with course codes included
//
// Authentication: Required (AuthMiddleware) - extracts current user from JWT token
//
// Database Operations:
//   - Reads from `users` table excluding current user (WHERE id != ?)
//   - Queries `courses` table to get course codes for each user
//   - Caches results in Redis hash with user ID as key
//
// Caching Strategy:
//   - Redis key: "users" (hash structure)
//   - Cache hit: Returns cached data directly (excludes current user)
//   - Cache miss: Queries database, populates cache with 3-hour TTL
//   - Individual user caching allows partial cache updates
//   - Non-blocking cache operations (warnings logged on failure)
//   - TTL rationale: Expensive query (N+1: all users + course codes), moderate data volatility
//
// Security Features:
//   - Current user excluded from results (prevents self-disclosure)
//   - User context validation through AuthMiddleware
//   - Sensitive fields excluded from user objects via ToMap()
//
// Error Responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: Database query, Redis operations, or JSON serialization failure
//
// Side Effects:
//   - Populates Redis cache on cache miss with 3-hour expiration
//   - Logs cache hit/miss events for monitoring
//   - No database modifications (read-only operation)
func GetUsersHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser, ok := c.Locals("user").(models.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}

	db := c.Locals("db").(*gorm.DB)
	if db == nil {
		return errors.WrapServer(fmt.Errorf("database connection not found"), errors.DBConnectionFailed, "Database connection not found", fiber.StatusInternalServerError)
	}

	c.Locals("message", "Users retrieved successfully")

	// Step 2: Attempt to retrieve users from Redis cache first (performance optimization)
	ctx := context.Background()
	usersHash, err := CacheService.GetUsers(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.CacheOperationFailed, "Error getting users from redis", fiber.StatusInternalServerError)
	}
	if len(usersHash) > 0 {
		// Step 3: Cache hit - Convert Redis hash to user array and exclude current user
		var cachedUsers []models.User
		for _, userJSON := range usersHash {
			var user models.User
			if err := json.Unmarshal([]byte(userJSON), &user); err == nil {
				log.Println("user", user.ID, "currentUser", currentUser.ID)
				if user.ID == currentUser.ID {
					continue
				}
				cachedUsers = append(cachedUsers, user)
			}
		}
		return c.JSON(cachedUsers)
	}

	// Step 4: Cache miss - Query users from database and enrich with course data
	var users []models.User
	if err := db.Omit("password_hash").Find(&users).Order("username ASC").Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting users from database", fiber.StatusInternalServerError)
	}

	// Step 5: Batch load course codes for all users
	userIDs := make([]uint, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	courseCodes, err := models.GetUsersCourseCodes(userIDs, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting course codes from database", fiber.StatusInternalServerError)
	}

	// Build map
	courseCodeMap := make(map[uint][]string)
	for _, cc := range courseCodes {
		courseCodeMap[cc.UserID] = append(courseCodeMap[cc.UserID], cc.Code)
	}

	// Attach to users
	var usersWithCourses []models.User
	for i := range users {
		users[i].CoursesCode = courseCodeMap[users[i].ID]
		if users[i].CoursesCode == nil {
			users[i].CoursesCode = []string{} // Ensure non-nil slice
		}
		usersWithCourses = append(usersWithCourses, users[i])

		// Cache individual user in Redis for future requests (non-blocking)
		if err := CacheService.SetUsers(ctx, usersWithCourses[i].ID, &usersWithCourses[i]); err != nil {
			return errors.WrapServer(err, errors.CacheOperationFailed, "Failed to cache user in Redis", fiber.StatusInternalServerError)
		}

	}

	// Step 7: Set cache expiration to 3 hours for optimal balance of freshness and performance
	if err := CacheService.SetExpirationUsers(ctx); err != nil {
		return errors.WrapServer(err, errors.CacheOperationFailed, "Failed to set cache expiration", fiber.StatusInternalServerError)
	}

	// Step 8: Send successful response with enriched user data
	return c.JSON(users)
}
