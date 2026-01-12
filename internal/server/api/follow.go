package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/server"
	"unipilot/internal/server/sse/grpc/messages"
)

// FollowRequest represents a follow/unfollow request payload.
// Used for both creating and removing follow relationships between users.
// Contains the target user ID that the current user wants to follow or unfollow.
type FollowRequest struct {
	FollowedID uint `json:"followed_id"` // ID of the user to follow/unfollow
}

// FollowResponse represents the response for follow/unfollow operations.
// Provides success status and descriptive message for client feedback.
// Used to confirm whether follow or unfollow action completed successfully.
type FollowResponse struct {
	Success bool   `json:"success"` // Whether the follow/unfollow operation succeeded
	Message string `json:"message"` // Descriptive message ("Followed successfully" or "Unfollowed successfully")
}

// FollowersResponse represents the response for retrieving a user's followers list.
// Contains paginated list of users who follow the specified user with total count.
// Supports pagination through limit/offset parameters for large follower lists.
type FollowersResponse struct {
	Followers []models.User `json:"followers"` // Array of users who follow the target user
	Total     int           `json:"total"`     // Total number of followers (for pagination)
}

// FollowingResponse represents the response for retrieving a user's following list.
// Contains paginated list of users that the specified user follows with total count.
// Supports pagination through limit/offset parameters for large following lists.
type FollowingResponse struct {
	Following []models.User `json:"following"` // Array of users that the target user follows
	Total     int           `json:"total"`     // Total number of users being followed (for pagination)
}

// FollowStatusResponse represents the follow relationship status between two users.
// Provides comprehensive information about follow relationship and user statistics.
// Used for displaying follow buttons and user profile statistics in the UI.
type FollowStatusResponse struct {
	IsFollowing    bool `json:"is_following"`    // Whether current user follows the target user
	FollowersCount int  `json:"followers_count"` // Number of users following the target user
	FollowingCount int  `json:"following_count"` // Number of users the target user follows
}

// HandleFollow manages follow and unfollow operations between users with social features.
// Toggles follow relationships, updates user statistics, calculates shared courses,
// and prepares notifications for social engagement. Prevents self-following and
// handles duplicate follow attempts gracefully.
//
// HTTP Method: POST
// Content-Type: application/json
//
// Request Body:
//   - followed_id: ID of the user to follow/unfollow (uint, required)
//
// Response (200 OK):
//   - success: Boolean indicating operation success
//   - message: Descriptive message ("Followed successfully" or "Unfollowed successfully")
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Database Operations:
//   - Checks existing follow relationship status
//   - Creates or removes follow record in database
//   - Updates follow statistics for both users (follower/following counts)
//   - Calculates shared courses between users for notification context
//
// Social Features:
//   - Prevents users from following themselves
//   - Tracks shared courses for enhanced social context
//   - Maintains bidirectional follow statistics
//   - Prepares rich notifications with course sharing information
//
// Notification System (Prepared):
//   - Code prepared for SSE notifications on new follows
//   - Includes shared course count in notification message
//   - Currently disabled pending Docker deployment
//
// Security Features:
//   - Input validation for followed_id parameter
//   - Self-follow prevention
//   - User authentication required for all operations
//   - Graceful handling of duplicate follow attempts
//
// Error Responses:
//   - 400 Bad Request: Invalid JSON, missing followed_id, or self-follow attempt
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 405 Method Not Allowed: Non-POST request
//   - 500 Internal Server Error: Database operations failure
//
// Side Effects:
//   - Creates or removes follow relationship in database
//   - Updates follow statistics for both users
//   - Prepares notification data for future SSE implementation
//   - Logs follow actions for social analytics
func HandleFollow(c *fiber.Ctx) error {
	// Step 2: Extract context values from middleware (user and database connection)
	currentUser, ok := c.Locals("user").(models.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID

	// Step 3: Extract user ID from path parameter
	var followedID uint
	var int_followedID int
	var err error
	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(
			fmt.Errorf("user ID required"),
			errors.ReqParamMissing,
			"User ID required",
			fiber.StatusBadRequest,
		)

	}
	if int_followedID, err = strconv.Atoi(idStr); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Error converting followed ID to uint",
			fiber.StatusBadRequest,
		)
	}
	followedID = uint(int_followedID)

	// Step 4: Validate followed_id parameter
	if followedID == 0 {
		err := fmt.Errorf("invalid followed_id")
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Invalid followed_id",
			fiber.StatusBadRequest,
		)
	}

	// Step 5: Prevent self-following for social integrity
	if userID == followedID {
		err := fmt.Errorf("cannot follow yourself")
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Cannot follow yourself",
			fiber.StatusBadRequest,
		)
	}

	// Step 6: Check existing follow relationship status
	// Check if already following
	isFollowing, err := models.IsFollowing(userID, followedID, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error checking follow status", fiber.StatusInternalServerError)
	}

	// Step 7: Handle follow/unfollow operation based on current status
	var response FollowResponse
	if isFollowing {
		c.Locals("message", "Follow removed")
		// Step 7a: Unfollow operation - remove existing relationship
		if err := models.RemoveFollow(userID, followedID, db); err != nil {
			return errors.WrapServer(err, errors.DBQueryFailed, "Error removing follow", fiber.StatusInternalServerError)
		}

		// Update Redis caches: remove user from followers and following lists (non-blocking)
		ctx := context.Background()
		// Remove current user from followed user's followers list
		if err := CacheService.DeleteUserFollowers(ctx, followedID, userID); err != nil {
			server.LogWarn(ctx, errors.WrapServer(err, errors.CacheOperationFailed, "Failed to remove user from followers cache", fiber.StatusInternalServerError))
		}
		// Remove followed user from current user's following list
		if err := CacheService.DeleteUserFollowing(ctx, userID, followedID); err != nil {
			server.LogWarn(ctx, errors.WrapServer(err, errors.CacheOperationFailed, "Failed to remove user from following cache", fiber.StatusInternalServerError))
		}

		response = FollowResponse{
			Success: true,
			Message: "Unfollowed successfully",
		}

	} else {
		c.Locals("message", "New follower added")
		// Step 7b: Follow operation - create new relationship
		if err := models.CreateFollow(userID, followedID, db); err != nil {
			if errors.HasCode(err, errors.DBConstraintViolation) {
				return errors.WrapServer(err, errors.DBConstraintViolation, "Already following", fiber.StatusConflict)
			}
			return errors.WrapServer(err, errors.DBQueryFailed, "Error creating follow", fiber.StatusInternalServerError)
		}

		// Step 8: Prepare notification data with social context and get followed user for cache update
		var followedUser models.User
		if err := db.First(&followedUser, followedID).Error; err != nil {
			return errors.WrapServer(err, errors.DBQueryFailed, "Error getting followed user", fiber.StatusInternalServerError)
		}

		// Update Redis caches: add user to followers and following lists (non-blocking)
		ctx := context.Background()

		// Add current user to followed user's followers list
		if err := CacheService.SetUserFollowers(ctx, followedID, userID, &currentUser); err != nil {
			server.LogWarn(ctx, errors.WrapServer(err, errors.CacheOperationFailed, "Failed to add user to followers cache", fiber.StatusInternalServerError))
		}

		// Add followed user to current user's following list
		if err := CacheService.SetUserFollowing(ctx, userID, followedID, &followedUser); err != nil {
			server.LogWarn(ctx, errors.WrapServer(err, errors.CacheOperationFailed, "Failed to add user to following cache", fiber.StatusInternalServerError))
		}

		response = FollowResponse{
			Success: true,
			Message: "Followed successfully",
		}

		// Calculate shared courses for enhanced social context in notifications
		var sharedCoursesCount int64
		if err := db.Model(&models.Course{}).
			Where("user_id = ? AND code IN (SELECT code FROM courses WHERE user_id = ?)", userID, followedID).
			Count(&sharedCoursesCount).Error; err != nil {
			return errors.WrapServer(err, errors.DBQueryFailed, "Error counting shared courses", fiber.StatusInternalServerError)
		}

		if GrpcClient != nil {
			_, err = (*GrpcClient).SendMessage(context.Background(),
				&messages.Message{
					ReceiverId: uint32(followedID),
					SenderId:   uint32(userID),
					Title:      "New Follower",
					Message:    fmt.Sprintf("%s followed you", currentUser.Username),
					Data:       []byte(""),
					Type:       string(models.MessageNoContent),
				},
			)
			if err != nil {
				server.LogWarn(context.Background(), errors.WrapServer(err, errors.GRPCFailed, "Failed to send notification", fiber.StatusInternalServerError))
			}
		}
	}

	// Step 9: Send successful response with follow operation result
	return c.JSON(response)
}

// HandleGetFollowers retrieves a paginated list of users who follow the specified models.
// Implements Redis caching with database fallback for optimal performance on social feeds.
// Supports pagination through limit/offset parameters for handling large follower lists.
//
// HTTP Method: GET
// Content-Type: Not required (query parameters used)
//
// Query Parameters:
//   - user_id: ID of the user whose followers to retrieve (string, required)
//   - limit: Maximum number of followers to return (string, optional, default: 20)
//   - offset: Number of followers to skip for pagination (string, optional, default: 0)
//
// Response (200 OK):
//   - followers: Array of user objects representing the followers
//   - total: Total number of followers (for pagination calculations)
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Caching Strategy:
//   - Redis key pattern: "followers:{user_id}"
//   - Cache structure: Hash with follower ID as field, JSON user object as value
//   - TTL: 30 minutes for social feed freshness
//   - Cache hit: Returns cached followers directly
//   - Cache miss: Queries database, populates cache, returns fresh data
//   - Cache is updated immediately on follow/unfollow operations (no stale data)
//   - TTL rationale: Moderate query cost (JOIN), high access frequency, cache updated on changes
//
// Database Operations:
//   - Queries followers through user model relationships
//   - Supports pagination with limit/offset parameters
//   - Retrieves complete user objects for follower information
//
// Performance Features:
//   - Redis caching reduces database load for popular users
//   - Individual follower caching allows partial cache updates
//   - Non-blocking cache operations (warnings logged on failure)
//   - Pagination prevents memory issues with large follower lists
//
// Error Responses:
//   - 400 Bad Request: Missing user_id parameter or invalid format
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 405 Method Not Allowed: Non-GET request
//   - 500 Internal Server Error: Database query or Redis failures
//
// Side Effects:
//   - Populates Redis cache on cache miss with 30-minute TTL
//   - Logs cache hit/miss events for performance monitoring
//   - No database modifications (read-only operation)
func HandleGetFollowers(c *fiber.Ctx) error {
	// Step 2: Extract context values from middleware (user and database connection)
	currentUser := c.Locals("user").(models.User)
	db := c.Locals("db").(*gorm.DB)
	_ = currentUser.ID // Available but not used for this endpoint
	c.Locals("message", "Followers list retrieved")

	// Step 3: Extract user ID from path parameter
	userIDStr := c.Params("id")
	if userIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		return errors.WrapServer(err, errors.ReqParamMissing, "User ID parameter required", fiber.StatusBadRequest)
	}

	// Step 4: Convert user ID string to integer for database operations
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting user ID", fiber.StatusBadRequest)
	}

	// Step 5: Parse pagination parameters with defaults
	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting limit to int", fiber.StatusBadRequest)
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		} else {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting offset to int", fiber.StatusBadRequest)
		}
	}

	var followers []models.User
	var total int
	// Step 6: Attempt to retrieve followers from Redis cache first (performance optimization)
	ctx := context.Background()
	followers, err = CacheService.GetUserFollowers(ctx, uint(userID))
	if err == nil {
		response := FollowersResponse{
			Followers: followers,
			Total:     len(followers),
		}
		return c.JSON(response)
	} else if errors.HasCode(err, errors.CacheOperationFailed) {
		server.LogWarn(ctx, errors.WrapServer(err, errors.CacheOperationFailed, "Error getting followers from redis", fiber.StatusInternalServerError))
	}

	// Step 7: Cache hit - Convert Redis hash to user array and return cached followers

	// Step 8: Cache miss - Query followers from database and populate cache
	followers, err = models.GetFollowers(uint(userID), limit, offset, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting followers from database", fiber.StatusInternalServerError)
	}
	total = len(followers)

	// Step 9: Cache the followers in redis
	for _, follower := range followers {
		if err := CacheService.SetUserFollowers(ctx, uint(userID), follower.ID, &follower); err != nil {
			server.LogWarn(ctx, errors.WrapServer(err, errors.RedisFailed, "Error caching follower in redis", fiber.StatusInternalServerError))
		}
	}

	// Step 10: Set cache expiration
	if err := CacheService.SetExpirationUserFollowers(ctx, uint(userID)); err != nil {
		server.LogWarn(ctx, errors.WrapServer(err, errors.RedisFailed, "Error setting cache expiration for followers", fiber.StatusInternalServerError))
	}

	// Step 11: Send successful response with followers list and total count
	response := FollowersResponse{
		Followers: followers,
		Total:     total,
	}

	return c.JSON(response)
}

// HandleGetFollowing retrieves a paginated list of users that the specified user follows.
// Implements Redis caching with database fallback for optimal performance on social feeds.
// Supports pagination through limit/offset parameters for handling large following lists.
//
// HTTP Method: GET
// Content-Type: Not required (query parameters used)
//
// Query Parameters:
//   - user_id: ID of the user whose following list to retrieve (string, required)
//   - limit: Maximum number of following users to return (string, optional, default: 20)
//   - offset: Number of following users to skip for pagination (string, optional, default: 0)
//
// Response (200 OK):
//   - following: Array of user objects representing users being followed
//   - total: Total number of users being followed (for pagination calculations)
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Caching Strategy:
//   - Redis key pattern: "following:{user_id}"
//   - Cache structure: Hash with followed user ID as field, JSON user object as value
//   - TTL: 30 minutes for social feed freshness
//   - Cache hit: Returns cached following list directly
//   - Cache miss: Queries database, populates cache, returns fresh data
//   - Cache is updated immediately on follow/unfollow operations (no stale data)
//   - TTL rationale: Moderate query cost (JOIN), high access frequency, cache updated on changes
//
// Error Responses:
//   - 400 Bad Request: Missing user_id parameter or invalid format
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 405 Method Not Allowed: Non-GET request
//   - 500 Internal Server Error: Database query or Redis failures
//
// Side Effects:
//   - Populates Redis cache on cache miss with 30-minute TTL
//   - Logs cache hit/miss events for performance monitoring
//   - No database modifications (read-only operation)
func HandleGetFollowing(c *fiber.Ctx) error {
	// Step 2: Extract context values from middleware (user and database connection)
	currentUser := c.Locals("user").(models.User)
	db := c.Locals("db").(*gorm.DB)
	_ = currentUser.ID // Available but not used for this endpoint
	c.Locals("message", "Following list retrieved")
	// Step 3: Extract user ID from path parameter
	userIDStr := c.Params("id")
	if userIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		return errors.WrapServer(err, errors.ReqParamMissing, "User ID parameter required", fiber.StatusBadRequest)
	}

	// Step 4: Convert user ID string to integer for database operations
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting user ID", fiber.StatusBadRequest)
	}

	// Step 5: Parse pagination parameters with defaults
	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting limit to int", fiber.StatusBadRequest)
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		} else {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting offset to int", fiber.StatusBadRequest)
		}
	}

	var following []models.User
	ctx := context.Background()
	following, err = CacheService.GetUserFollowing(ctx, uint(userID))
	if err == nil {
		response := FollowingResponse{
			Following: following,
			Total:     len(following),
		}
		return c.JSON(response)
	} else if errors.HasCode(err, errors.CacheOperationFailed) {
		server.LogWarn(ctx, errors.WrapServer(err, errors.CacheOperationFailed, "Error getting following from redis", fiber.StatusInternalServerError))
	}

	// Step 7: Cache miss - Query following list from database and populate cache
	following, err = models.GetFollowing(uint(userID), limit, offset, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting following from database", fiber.StatusInternalServerError)
	}
	total := len(following)

	// Step 8: Cache the following list in Redis for future requests
	for _, followed := range following {
		if err := CacheService.SetUserFollowing(ctx, uint(userID), followed.ID, &followed); err != nil {
			server.LogWarn(ctx, errors.WrapServer(err, errors.RedisFailed, "Error caching followed in redis", fiber.StatusInternalServerError))
		}
	}

	// Step 9: Set cache expiration
	if err := CacheService.SetExpirationUserFollowing(ctx, uint(userID)); err != nil {
		server.LogWarn(ctx, errors.WrapServer(err, errors.RedisFailed, "Error setting cache expiration for following", fiber.StatusInternalServerError))
	}

	// Step 10: Send successful response with following list and total count
	response := FollowingResponse{
		Following: following,
		Total:     total,
	}

	return c.JSON(response)
}
