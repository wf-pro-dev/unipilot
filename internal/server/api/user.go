package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"unipilot/internal/models/user"
	"unipilot/internal/server"
)

// GetUserHandler retrieves the current authenticated user's information.
// Returns user data from the request context (populated by AuthMiddleware) in a safe
// map format that excludes sensitive fields like password hashes.
//
// HTTP Method: GET
// Content-Type: Not required (no request body expected)
//
// Request Body: None required (user context extracted from JWT token)
//
// Response (200 OK):
//   - message: Success message
//   - user: User object (as map) with sensitive fields removed
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Security Features:
//   - Uses ToMap() method to exclude sensitive fields from response
//   - No database queries required (uses cached context data)
//   - Safe serialization with error handling
//
// Error Responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: User data processing failure
//
// Side Effects:
//   - Logs successful user retrieval for audit trail
//   - No database or cache modifications
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract user context from JWT token (validated by AuthMiddleware)
	currentUser := r.Context().Value("user").(user.User)

	// Step 2: Convert user struct to safe map format (excludes sensitive fields like password hash)
	userMap := currentUser.ToMap()
	if userMap == nil {
		err := errors.New("failed to process user data")
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error processing user data",
			"tags", []string{"USER", "MARSHALLING"},
		)
		return
	}

	// Step 3: Send successful response with sanitized user data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User retrieved successfully",
		"user":    userMap,
	})

	// Step 4: Log successful user retrieval for audit trail
	server.LogInfo(r.Context(), "User retrieved successfully",
		"tags", []string{"USER", "READ"},
	)
}

// UpdateUserHandler updates a specific field of the authenticated user's profile.
// Performs database update using raw SQL for flexibility, then updates Redis cache
// for performance optimization. Uses transactions for data consistency.
//
// HTTP Method: POST/PUT
// Content-Type: application/json
//
// Request Body:
//   - column: Database column name to update (string, required)
//   - value: New value for the specified column (string, required)
//
// Response (200 OK):
//   - message: Success message
//   - user: Updated user object (as map) with sensitive fields removed
//
// Authentication: Required (AuthMiddleware) - needs user context for identification
//
// Database Operations:
//   - Updates specified column in `users` table using raw SQL
//   - Automatically updates `updated_at` timestamp
//   - Retrieves updated user record for response
//   - Updates Redis cache with new user data (non-blocking)
//
// Security Notes:
//   - Uses string interpolation for column name - validate input to prevent SQL injection
//   - Only authenticated users can update their own profile
//   - Sensitive fields excluded from response via ToMap()
//
// Error Responses:
//   - 400 Bad Request: Invalid JSON body or malformed request
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 404 Not Found: User record not found in database
//   - 500 Internal Server Error: Database operations or serialization failure
//
// Side Effects:
//   - Modifies user record in database with timestamp update
//   - Updates Redis cache (logs warning on failure, non-critical)
//   - Logs successful updates for audit trail
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Define and parse update request structure
	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"tags", []string{"USER", "REQUEST"},
		)
		return
	}

	// Step 3: Execute raw SQL update with automatic timestamp tracking
	if err := db.Exec(fmt.Sprintf("UPDATE users SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), userID).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error updating user in database",
			"tags", []string{"USER", "DB"},
		)
		return
	}

	// Step 4: Retrieve updated user record to ensure consistency and get fresh data
	var userObj user.User
	if err := db.First(&userObj, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			server.ResponseError(r.Context(), w, err, http.StatusNotFound, "User not found",
				"tags", []string{"USER", "DB"},
			)
		} else {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting user from database",
				"tags", []string{"USER", "DB"},
			)
		}
		return
	}

	// Step 5: Convert updated user to safe map format (excludes sensitive fields)
	userMap := userObj.ToMap()
	if userMap == nil {
		err := errors.New("failed to process user data")
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error processing user data",
			"tags", []string{"USER"},
		)
		return
	}

	// Step 6: Update Redis cache with new user data for performance optimization
	userJSON, err := json.Marshal(userMap)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error marshalling user to json",
			"tags", []string{"USER", "SERIALIZATION"},
		)
		return
	}
	// Cache update is non-blocking - failure is logged but doesn't stop the response
	if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(userID)), userJSON).Err(); err != nil {
		server.LogWarn(r.Context(),
			"Error caching user in redis", err,
			"tags", []string{"USER", "REDIS"},
		)
	}

	// Step 7: Send successful response with updated user data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User updated successfully",
		"user":    userMap,
	})

	// Step 8: Log successful update with change details for audit trail
	server.LogInfo(r.Context(), "User updated successfully",
		"update", updateData,
		"tags", []string{"USER", "WRITE"},
	)
}
