package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
	"unipilot/internal/server"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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
	Followers []user.User `json:"followers"` // Array of users who follow the target user
	Total     int         `json:"total"`     // Total number of followers (for pagination)
}

// FollowingResponse represents the response for retrieving a user's following list.
// Contains paginated list of users that the specified user follows with total count.
// Supports pagination through limit/offset parameters for large following lists.
type FollowingResponse struct {
	Following []user.User `json:"following"` // Array of users that the target user follows
	Total     int         `json:"total"`     // Total number of users being followed (for pagination)
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
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 3: Extract user ID from path parameter
	var followedID uint
	var int_followedID int
	var err error
	idStr := c.Params("id")
	if idStr == "" {
		return server.ResponseError(c, fmt.Errorf("user ID required"), fiber.StatusBadRequest, "User ID required",
			"tags", []string{"FOLLOWS", "INVALID_FOLLOWED_ID"},
		)
	}
	if int_followedID, err = strconv.Atoi(idStr); err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error converting followed ID to uint",
			"tags", []string{"FOLLOWS", "INVALID_FOLLOWED_ID"},
		)
	}
	followedID = uint(int_followedID)

	// Step 4: Validate followed_id parameter
	if followedID == 0 {
		err := fmt.Errorf("invalid followed_id")
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid followed_id",
			"tags", []string{"FOLLOWS", "VALIDATION"},
		)
	}

	// Step 5: Prevent self-following for social integrity
	if userID == followedID {
		err := fmt.Errorf("cannot follow yourself")
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Cannot follow yourself",
			"tags", []string{"FOLLOWS", "VALIDATION"},
		)
	}

	// Step 6: Check existing follow relationship status
	// Check if already following
	isFollowing, err := user.IsFollowing(userID, followedID, db)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error checking follow status",
			"tags", []string{"FOLLOWS", "DB"},
		)
	}

	// Step 7: Handle follow/unfollow operation based on current status
	var response FollowResponse
	if isFollowing {
		// Step 7a: Unfollow operation - remove existing relationship
		if err := user.RemoveFollow(userID, followedID, db); err != nil {
			return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error removing follow",
				"tags", []string{"FOLLOWS", "DB"},
			)
		}

		// Update follow statistics for both users after unfollow
		if err := user.UpdateFollowStats(userID, db); err != nil {
			server.LogWarn(context.Background(), "Failed to update follow stats", err,
				"tags", []string{"follow", "db", "medium"},
				"error_type", "database",
			)
		}
		if err := user.UpdateFollowStats(followedID, db); err != nil {
			server.LogWarn(context.Background(), "Failed to update follow stats", err,
				"tags", []string{"follow", "db", "medium"},
				"error_type", "database",
			)
		}

		response = FollowResponse{
			Success: true,
			Message: "Unfollowed successfully",
		}
	} else {
		// Step 7b: Follow operation - create new relationship
		if err := user.CreateFollow(userID, followedID, db); err != nil {
			return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error creating follow",
				"tags", []string{"FOLLOWS", "DB"},
			)
		}

		// Update follow statistics for both users after follow
		if err := user.UpdateFollowStats(userID, db); err != nil {
			server.LogWarn(context.Background(), "Failed to update follow stats", err,
				"tags", []string{"follow", "db", "medium"},
				"error_type", "database",
			)
		}
		if err := user.UpdateFollowStats(followedID, db); err != nil {
			server.LogWarn(context.Background(), "Failed to update follow stats", err,
				"tags", []string{"follow", "db", "medium"},
				"error_type", "database",
			)
		}

		response = FollowResponse{
			Success: true,
			Message: "Followed successfully",
		}

		// Step 8: Prepare notification data with social context
		var followedUser user.User
		if err := db.First(&followedUser, followedID).Error; err != nil {
			server.LogWarn(context.Background(), "Failed to get followed user", err,
				"tags", []string{"follow", "db", "low"},
				"error_type", "database",
			)
		}

		// Calculate shared courses for enhanced social context in notifications
		var sharedCoursesCount int64
		if err := db.Model(&course.Course{}).
			Where("user_id = ? AND code IN (SELECT code FROM courses WHERE user_id = ?)", userID, followedID).
			Count(&sharedCoursesCount).Error; err != nil {
			server.LogWarn(context.Background(), "Failed to count shared courses", err,
				"tags", []string{"course", "db", "low"},
				"error_type", "database",
			)
		}

		/*if sseServer != nil {
			sseServer.SendNotification(
				followedID,
				userID,
				models.EntityFollow,
				followedID,
				notifications.NotificationFollow,
				currentUser.Username,
				fmt.Sprintf("%s followed you. You share %d courses with this user", currentUser.Username ,sharedCoursesCount),
				"create",
				"",
			)
		}*/
	}

	// Step 10: Log follow action for social analytics and audit trail
	action := "follow"
	if isFollowing {
		action = "unfollow"
	}
	server.LogInfo(context.Background(), "Follow action completed", "followed_id", followedID, "action", action,
		"tags", []string{"follow", "db", "medium", "update"})

	// Step 9: Send successful response with follow operation result
	return c.JSON(response)
}

// HandleGetFollowers retrieves a paginated list of users who follow the specified user.
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
//   - TTL: 10 minutes for social feed freshness
//   - Cache hit: Returns cached followers directly
//   - Cache miss: Queries database, populates cache, returns fresh data
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
//   - Populates Redis cache on cache miss with 10-minute TTL
//   - Logs cache hit/miss events for performance monitoring
//   - No database modifications (read-only operation)
func HandleGetFollowers(c *fiber.Ctx) error {
	// Step 2: Extract context values from middleware (user and database connection)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	_ = currentUser.ID // Available but not used for this endpoint

	// Step 3: Extract user ID from path parameter
	userIDStr := c.Params("id")
	if userIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		return server.ResponseError(c, err, fiber.StatusBadRequest, "User ID parameter required",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
	}

	// Step 4: Convert user ID string to integer for database operations
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error converting user ID",
			"tags", []string{"FOLLOWS", "INVALID_USER_ID"},
		)
	}

	// Step 5: Parse pagination parameters with defaults
	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Step 6: Attempt to retrieve followers from Redis cache first (performance optimization)
	var cacheKey = fmt.Sprintf("followers:%d", userID)
	followersHash, err := RedisClient.HGetAll(context.Background(), cacheKey).Result()
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting followers from redis",
			"tags", []string{"FOLLOWS", "REDIS"},
		)
	}

	// Step 7: Cache hit - Convert Redis hash to user array and return cached followers
	if len(followersHash) > 0 {
		var cachedFollowers []user.User
		for _, followerJSON := range followersHash {
			var follower user.User
			if err := json.Unmarshal([]byte(followerJSON), &follower); err == nil {
				cachedFollowers = append(cachedFollowers, follower)
			}
		}
		response := FollowersResponse{
			Followers: cachedFollowers,
			Total:     len(cachedFollowers),
		}
		return c.JSON(response)
	}

	// Step 8: Cache miss - Query followers from database and populate cache
	followers, err := user.GetFollowers(uint(userID), limit, offset, db)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting followers from database",
			"tags", []string{"FOLLOWS", "DB"},
		)
	}
	total := len(followers)

	// Step 9: Cache the followers in redis
	for _, follower := range followers {
		followerJSON, err := json.Marshal(follower)
		if err != nil {
			server.LogWarn(context.Background(), "Failed to marshal follower to JSON", err,
				"tags", []string{"follow", "io", "low"},
				"error_type", "internal",
			)
			continue
		}
		if err := RedisClient.HSet(context.Background(), cacheKey, strconv.Itoa(int(follower.ID)), followerJSON).Err(); err != nil {
			server.LogWarn(context.Background(), "Failed to cache follower in Redis", err,
				"tags", []string{"cache", "cache", "medium"},
				"cache_status", "error",
				"error_type", "cache",
			)
		}
	}

	// Step 10: Set cache expiration to 10 minutes for optimal balance of freshness and performance
	if err := RedisClient.Expire(context.Background(), cacheKey, 10*time.Minute).Err(); err != nil {
		server.LogWarn(context.Background(), "Failed to set cache expiration for followers", err,
			"tags", []string{"cache", "cache", "low"},
			"error_type", "cache",
		)
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
//   - TTL: 10 minutes for social feed freshness
//   - Cache hit: Returns cached following list directly
//   - Cache miss: Queries database, populates cache, returns fresh data
//
// Error Responses:
//   - 400 Bad Request: Missing user_id parameter or invalid format
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 405 Method Not Allowed: Non-GET request
//   - 500 Internal Server Error: Database query or Redis failures
//
// Side Effects:
//   - Populates Redis cache on cache miss with 10-minute TTL
//   - Logs cache hit/miss events for performance monitoring
//   - No database modifications (read-only operation)
func HandleGetFollowing(c *fiber.Ctx) error {
	// Step 2: Extract context values from middleware (user and database connection)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	_ = currentUser.ID // Available but not used for this endpoint

	// Step 3: Extract user ID from path parameter
	userIDStr := c.Params("id")
	if userIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		return server.ResponseError(c, err, fiber.StatusBadRequest, "User ID parameter required",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
	}

	// Step 4: Convert user ID string to integer for database operations
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error converting user ID",
			"tags", []string{"FOLLOWS", "INVALID_USER_ID"},
		)
	}

	// Step 5: Parse pagination parameters with defaults
	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Step 6: Attempt to retrieve following list from Redis cache first (performance optimization)
	var cacheKey = fmt.Sprintf("following:%d", userID)
	followingHash, err := RedisClient.HGetAll(context.Background(), cacheKey).Result()
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting following from redis",
			"tags", []string{"FOLLOWS", "REDIS"},
		)
	}

	// Step 7: Cache hit - Convert Redis hash to user array and return cached following list
	if len(followingHash) > 0 {
		var cachedFollowing []user.User
		for _, followingJSON := range followingHash {
			var following user.User
			if err := json.Unmarshal([]byte(followingJSON), &following); err == nil {
				cachedFollowing = append(cachedFollowing, following)
			}
		}
		response := FollowingResponse{
			Following: cachedFollowing,
			Total:     len(cachedFollowing),
		}
		return c.JSON(response)
	}

	// Step 8: Cache miss - Query following list from database and populate cache
	following, err := user.GetFollowing(uint(userID), limit, offset, db)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting following from database",
			"tags", []string{"FOLLOWS", "DB"},
		)
	}
	total := len(following)

	// Step 9: Cache the following list in Redis for future requests
	for _, followed := range following {
		followedJSON, err := json.Marshal(followed)
		if err != nil {
			server.LogWarn(context.Background(), "Failed to marshal followed to JSON", err,
				"tags", []string{"follow", "io", "low"},
				"error_type", "internal",
			)
			continue
		}
		if err := RedisClient.HSet(context.Background(), cacheKey, strconv.Itoa(int(followed.ID)), followedJSON).Err(); err != nil {
			server.LogWarn(context.Background(), "Failed to cache followed in Redis", err,
				"tags", []string{"cache", "cache", "medium"},
				"cache_status", "error",
				"error_type", "cache",
			)
		}
	}

	// Step 10: Set cache expiration to 10 minutes for optimal balance of freshness and performance
	if err := RedisClient.Expire(context.Background(), cacheKey, 10*time.Minute).Err(); err != nil {
		server.LogWarn(context.Background(), "Failed to set cache expiration for following", err,
			"tags", []string{"cache", "cache", "low"},
			"error_type", "cache",
		)
	}

	// Step 11: Send successful response with following list and total count
	response := FollowingResponse{
		Following: following,
		Total:     total,
	}

	return c.JSON(response)
}
