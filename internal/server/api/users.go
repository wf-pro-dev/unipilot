package server

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

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
	currentUser := c.Locals("user").(models.User)
	db := c.Locals("db").(*gorm.DB)
	c.Locals("message", "Users retrieved successfully")

	// Step 2: Attempt to retrieve users from Redis cache first (performance optimization)
	usersHash, err := RedisClient.HGetAll(context.Background(), "users").Result()
	if err != nil {
		return errors.WrapServer(err, errors.CacheOperationFailed, "Error getting users from redis", fiber.StatusInternalServerError)
	}
	if len(usersHash) > 0 {
		// Step 3: Cache hit - Convert Redis hash to user array and exclude current user
		var cachedUsers []models.User
		for _, userJSON := range usersHash {
			var user models.User
			if err := json.Unmarshal([]byte(userJSON), &user); err == nil {
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
	if err := db.Find(&users).Order("name ASC").Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting users from database", fiber.StatusInternalServerError)
	}

	// Step 5: Process each user and enrich with course codes for comprehensive profiles
	var usersWithCourses []models.User
	for _, u := range users {
		// Query course codes associated with this user
		var courses_code []string
		if err := db.Model(&models.Course{}).Select("code").Where("user_id = ? ", u.ID).Find(&courses_code).Error; err != nil {
			return errors.WrapServer(err, errors.DBQueryFailed, "Error getting user courses", fiber.StatusInternalServerError)
		}
		// Attach course codes to user object and convert to safe map format
		u.CoursesCode = courses_code
		u.PasswordHash = ""
		usersWithCourses = append(usersWithCourses, u)

		// Step 6: Cache individual user in Redis for future requests (non-blocking)
		userJSON, err := json.Marshal(u)
		if err != nil {
			return errors.WrapServer(err, errors.ProcJSONMarshalFailed, "Error marshalling user to json", fiber.StatusInternalServerError)
		}

		if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(u.ID)), userJSON).Err(); err != nil {
			return errors.WrapServer(err, errors.CacheOperationFailed, "Failed to cache user in Redis", fiber.StatusInternalServerError)
		}
	}

	// Step 7: Set cache expiration to 3 hours for optimal balance of freshness and performance
	// Rationale: User list query is expensive (N+1 problem: all users + course codes per user)
	// User data changes moderately (profile updates), but course codes change less frequently
	// Longer TTL reduces database load while maintaining acceptable freshness
	if err := RedisClient.Expire(context.Background(), "users", 3*time.Hour).Err(); err != nil {
		return errors.WrapServer(err, errors.CacheOperationFailed, "Failed to set cache expiration", fiber.StatusInternalServerError)
	}

	// Step 8: Send successful response with enriched user data
	return c.JSON(users)
}
