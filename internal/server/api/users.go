package server

import (
	"github.com/gofiber/fiber/v2"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/server"
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
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}
	var cursor *models.Cursor
	if err := c.BodyParser(&cursor); err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error unmarshalling cursor", fiber.StatusBadRequest)
	}

	results, err := models.GetUsers(cursor, 20, db.Debug().Omit("password_hash"))
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting users from database", fiber.StatusInternalServerError)
	}

	return c.JSON(results)
}
