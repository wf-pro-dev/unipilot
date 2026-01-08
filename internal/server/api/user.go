package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"unipilot/internal/errors"
	"unipilot/internal/models"
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
	currentUser := c.Locals("user").(models.User)

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
	c.Locals("message", "User updated successfully")
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser, ok := c.Locals("user").(models.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.AuthUnauthorized, "User not found", fiber.StatusUnauthorized)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("database connection not found"), errors.DBConnectionFailed, "Database connection not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID

	// Step 2: Define and parse update request structure
	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err := c.BodyParser(&updateData)
	if err != nil {
		return errors.WrapServer(err, errors.ReqBodyInvalid, "Invalid request body", fiber.StatusBadRequest)
	}

	// Step 3: Execute raw SQL update with automatic timestamp tracking
	if err := db.Exec(fmt.Sprintf("UPDATE users SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), userID).Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error updating user in database", fiber.StatusInternalServerError)
	}

	// Step 4: Retrieve updated user record to ensure consistency and get fresh data
	var userObj models.User
	if err := db.First(&userObj, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.WrapServer(err, errors.DBRecordNotFound, "User not found", fiber.StatusNotFound)
		}
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting user from database", fiber.StatusInternalServerError)
	}

	// Step 5: Convert updated user to safe map format (excludes sensitive fields)
	userMap := userObj.ToMap()
	if userMap == nil {
		err := fmt.Errorf("failed to process user data")
		return errors.WrapServer(err, errors.ProcDataProcessingFailed, "Error processing user data", fiber.StatusInternalServerError)
	}

	// Step 6: Update Redis cache with new user data for performance optimization
	userJSON, err := json.Marshal(userMap)
	if err != nil {
		return errors.WrapServer(err, errors.ProcJSONMarshalFailed, "Error marshalling user to json", fiber.StatusInternalServerError)
	}
	// Cache update is non-blocking - failure is logged but doesn't stop the response
	if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(userID)), userJSON).Err(); err != nil {
		server.LogWarn(c.Context(), errors.WrapServer(err, errors.CacheOperationFailed, "Failed to cache user in Redis", fiber.StatusInternalServerError))
	}

	// Step 7: Send successful response with updated user data
	return c.JSON(fiber.Map{
		"message": "User updated successfully",
		"user":    userMap,
	})
}

func UpdateProfilePictureHandler(c *fiber.Ctx) error {
	c.Locals("message", "Profile picture updated successfully")
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser, ok := c.Locals("user").(models.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.AuthUnauthorized, "User not found", fiber.StatusUnauthorized)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("database connection not found"), errors.DBConnectionFailed, "Database connection not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID

	// Step 2: Extract file from multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return errors.WrapServer(err, errors.ReqBodyInvalid, "Unable to get file from form", fiber.StatusBadRequest)
	}

	// Step 3: Generate unique storage paths and file names for cloud storage
	// Organize files by user ID to enable easy cleanup and access control
	profilePictureDir := fmt.Sprintf("users_data/user_%d", userID)

	// Use Unix timestamp to ensure uniqueness and prevent filename collisions
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), "profile_picture.jpg")
	newKey := fmt.Sprintf("%s/%s", profilePictureDir, uniqueFileName)

	// Step 4: Extract file from multipart form and write to local disk
	// Local storage is required as intermediate step before S3 upload
	filePath, _, err := WriteMultipartFile(newKey, file, c)
	if err != nil {
		return errors.WrapServer(err, errors.FSWriteFailed, "Error writing file to disk", fiber.StatusInternalServerError)
	}

	// Step 5: Upload file from local disk to AWS S3 cloud storage
	// This persists the file permanently and makes it accessible via CDN
	var publicURL string
	if publicURL, err = cloudstorage.UploadProfilePicture(filePath, uniqueFileName, newKey); err != nil {
		return errors.WrapServer(err, errors.StorageUploadFailed, "Error uploading file to S3", fiber.StatusInternalServerError)
	}

	// Step 6: Clean up local temporary file after successful S3 upload
	// Prevents disk space accumulation from temporary upload files
	os.Remove(filePath)

	// Step 7: Update user record in database with new avatar S3 key
	// Store the S3 path so the application can generate CDN URLs for the image
	currentUser.Avatar = publicURL
	if err := db.Save(&currentUser).Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error updating user profile picture in database", fiber.StatusInternalServerError)
	}

	// Step 8: Update Redis cache with new user data including updated avatar (non-blocking)
	userMap := currentUser.ToMap()
	if userMap != nil {
		userJSON, err := json.Marshal(userMap)
		if err == nil {
			if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(userID)), userJSON).Err(); err != nil {
				server.LogWarn(c.Context(), errors.WrapServer(err, errors.CacheOperationFailed, "Failed to cache user in Redis after profile picture update", fiber.StatusInternalServerError))
			}
		}
	}

	// Step 9: Send successful response with confirmation message
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile picture updated successfully",
	})
}

func GetUserCourseInvitationsHandler(c *fiber.Ctx) error {
	c.Locals("message", "User course invitations retrieved successfully")

	currentUser, ok := c.Locals("user").(models.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}

	invitations, err := models.GetUserCourseInvitations(currentUser.ID, db.Where("status = ?", models.InvitationPending))
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting pending invitations", fiber.StatusInternalServerError)
	}
	return c.JSON(invitations)
}
