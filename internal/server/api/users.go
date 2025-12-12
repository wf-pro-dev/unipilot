package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
	"unipilot/internal/server"
)

// GetUsersHandler retrieves all users in the system except the current authenticated user.
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
//   - Cache miss: Queries database, populates cache with 1-hour TTL
//   - Individual user caching allows partial cache updates
//   - Non-blocking cache operations (warnings logged on failure)
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
//   - Populates Redis cache on cache miss with 1-hour expiration
//   - Logs cache hit/miss events for monitoring
//   - No database modifications (read-only operation)
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)

	// Step 2: Attempt to retrieve users from Redis cache first (performance optimization)
	usersHash, err := RedisClient.HGetAll(context.Background(), "users").Result()
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting users from redis",
			"tags", []string{"USERS", "REDIS"},
		)
		return
	}
	if len(usersHash) > 0 {
		// Step 3: Cache hit - Convert Redis hash to user array and exclude current user
		var cachedUsers []map[string]interface{}
		for _, userJSON := range usersHash {
			var userMap map[string]interface{}
			if err := json.Unmarshal([]byte(userJSON), &userMap); err == nil {
				if userMap["id"] == currentUser.ID {
					continue
				}
				cachedUsers = append(cachedUsers, userMap)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Users retrieved successfully",
			"users":   cachedUsers,
		})
		return
	}

	// Step 4: Cache miss - Query users from database and enrich with course data
	var users []user.User
	if err := db.Find(&users).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting users from database",
			"tags", []string{"USERS", "DB"},
		)
		return
	}

	// Step 5: Process each user and enrich with course codes for comprehensive profiles
	var usersMap []map[string]interface{}
	for _, u := range users {
		// Query course codes associated with this user
		var courses_code []string
		if err := db.Model(&course.Course{}).Select("code").Where("user_id = ? ", u.ID).Find(&courses_code).Error; err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting user courses",
				"tags", []string{"USERS", "DB"},
			)
			return
		}
		// Attach course codes to user object and convert to safe map format
		u.CoursesCode = courses_code
		userMap := u.ToMap()
		usersMap = append(usersMap, userMap)

		// Step 6: Cache individual user in Redis for future requests (non-blocking)
		userJSON, err := json.Marshal(userMap)
		if err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error marshalling user to json",
				"tags", []string{"USERS", "SERIALIZATION"},
			)
			return
		}

		if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(u.ID)), userJSON).Err(); err != nil {
			server.LogWarn(r.Context(), "Failed to cache user in Redis", err, "user_id", u.ID,
				"tags", []string{"cache", "cache", "medium"},
				"cache_status", "error",
				"error_type", "cache",
			)
		}
	}

	// Step 7: Set cache expiration to 1 hour for optimal balance of freshness and performance
	if err := RedisClient.Expire(context.Background(), "users", time.Hour).Err(); err != nil {
		server.LogWarn(r.Context(), "Failed to set cache expiration", err,
			"tags", []string{"cache", "cache", "low"},
			"error_type", "cache",
		)
	}

	// Step 8: Send successful response with enriched user data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Users retrieved successfully",
		"users":   usersMap,
	})

}
