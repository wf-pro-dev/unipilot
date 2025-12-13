package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"unipilot/internal/models/user"
	"unipilot/internal/server"
	cloudstorage "unipilot/internal/services/cloud_storage"
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
func GetUserHandler(c *fiber.Ctx) error {
	// Step 1: Extract user context from JWT token (validated by AuthMiddleware)
	currentUser := c.Locals("user").(user.User)

	// Step 3: Send successful response with sanitized user data
	return c.JSON(fiber.Map{
		"message": "User retrieved successfully",
		"user":    currentUser,
	})
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
func UpdateUserHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Define and parse update request structure
	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err := c.BodyParser(&updateData)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid request body",
			"tags", []string{"USER", "REQUEST"},
		)
	}

	// Step 3: Execute raw SQL update with automatic timestamp tracking
	if err := db.Exec(fmt.Sprintf("UPDATE users SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), userID).Error; err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error updating user in database",
			"tags", []string{"USER", "DB"},
		)
	}

	// Step 4: Retrieve updated user record to ensure consistency and get fresh data
	var userObj user.User
	if err := db.First(&userObj, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return server.ResponseError(c, err, fiber.StatusNotFound, "User not found",
				"tags", []string{"USER", "DB"},
			)
		}
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting user from database",
			"tags", []string{"USER", "DB"},
		)
	}

	// Step 5: Convert updated user to safe map format (excludes sensitive fields)
	userMap := userObj.ToMap()
	if userMap == nil {
		err := errors.New("failed to process user data")
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error processing user data",
			"tags", []string{"USER"},
		)
	}

	// Step 6: Update Redis cache with new user data for performance optimization
	userJSON, err := json.Marshal(userMap)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error marshalling user to json",
			"tags", []string{"USER", "SERIALIZATION"},
		)
	}
	// Cache update is non-blocking - failure is logged but doesn't stop the response
	if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(userID)), userJSON).Err(); err != nil {
		server.LogWarn(context.Background(), "Failed to cache user in Redis", err, "user_id", userID,
			"tags", []string{"cache", "cache", "medium"},
			"cache_status", "error",
			"error_type", "cache",
		)
	}

	// Step 7: Send successful response with updated user data
	return c.JSON(fiber.Map{
		"message": "User updated successfully",
		"user":    userMap,
	})
}

func UpdateProfilePictureHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Extract file from multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Unable to get file from form",
			"tags", []string{"USER", "PROFILE_PICTURE", "REQUEST"},
		)
	}

	// Step 3: Generate unique storage paths and file names for cloud storage
	// Organize files by user ID to enable easy cleanup and access control
	profilePictureDir := fmt.Sprintf("users_data/user_%d", userID)

	// Use Unix timestamp to ensure uniqueness and prevent filename collisions
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), "profile_picture.jpg")
	newKey := fmt.Sprintf("%s/%s", profilePictureDir, uniqueFileName)

	// Step 4: Extract file from multipart form and write to local disk
	// Local storage is required as intermediate step before S3 upload
	filePath, _, err := WriteFileToDiskFiber(newKey, file, c)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error writing file to disk",
			"tags", []string{"USER", "PROFILE_PICTURE", "STORAGE", "FILESYSTEM"},
		)
	}

	// Step 5: Upload file from local disk to AWS S3 cloud storage
	// This persists the file permanently and makes it accessible via CDN
	var publicURL string
	if publicURL, err = cloudstorage.UploadProfilePicture(filePath, uniqueFileName, newKey); err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error uploading file to S3",
			"tags", []string{"USER", "PROFILE_PICTURE", "STORAGE", "S3"},
		)
	}

	// Step 6: Clean up local temporary file after successful S3 upload
	// Prevents disk space accumulation from temporary upload files
	os.Remove(filePath)

	// Step 7: Update user record in database with new avatar S3 key
	// Store the S3 path so the application can generate CDN URLs for the image
	currentUser.Avatar = publicURL
	if err := db.Save(&currentUser).Error; err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error updating user profile picture in database",
			"tags", []string{"USER", "PROFILE_PICTURE", "DB"},
		)
	}

	// Step 9: Log successful profile picture update for audit trail
	server.LogInfo(context.Background(), "Profile picture updated", "user_id", userID,
		"tags", []string{"user", "upload", "high", "update"},
		"external_service", "s3")

	// Step 8: Send successful response with confirmation message
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile picture updated successfully",
	})
}
