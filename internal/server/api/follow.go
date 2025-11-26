package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
	"unipilot/internal/server"

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
func HandleFollow(w http.ResponseWriter, r *http.Request) {
	// Step 1: Enforce POST-only endpoint for follow operations
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 2: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 3: Parse follow request from JSON body
	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
		return
	}

	// Step 4: Validate followed_id parameter
	if req.FollowedID == 0 {
		err := fmt.Errorf("invalid followed_id")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid followed_id",
			"tags", []string{"FOLLOWS", "VALIDATION"},
		)
		return
	}

	// Step 5: Prevent self-following for social integrity
	if userID == req.FollowedID {
		err := fmt.Errorf("cannot follow yourself")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Cannot follow yourself",
			"tags", []string{"FOLLOWS", "VALIDATION"},
		)
		return
	}

	// Step 6: Check existing follow relationship status
	// Check if already following
	isFollowing, err := user.IsFollowing(userID, req.FollowedID, db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error checking follow status",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}

	// Step 7: Handle follow/unfollow operation based on current status
	var response FollowResponse
	if isFollowing {
		// Step 7a: Unfollow operation - remove existing relationship
		if err := user.RemoveFollow(userID, req.FollowedID, db); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error removing follow",
				"tags", []string{"FOLLOWS", "DB"},
			)
			return
		}

		// Update follow statistics for both users after unfollow
		if err := user.UpdateFollowStats(userID, db); err != nil {
			server.LogWarn(r.Context(),
				"Error updating follow stats", err,
				"tags", []string{"FOLLOWS", "DB"},
			)
		}
		if err := user.UpdateFollowStats(req.FollowedID, db); err != nil {
			server.LogWarn(r.Context(),
				"Error updating follow stats", err,
				"tags", []string{"FOLLOWS", "DB"},
			)
		}

		response = FollowResponse{
			Success: true,
			Message: "Unfollowed successfully",
		}
	} else {
		// Step 7b: Follow operation - create new relationship
		if err := user.CreateFollow(userID, req.FollowedID, db); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error creating follow",
				"tags", []string{"FOLLOWS", "DB"},
			)
			return
		}

		// Update follow statistics for both users after follow
		if err := user.UpdateFollowStats(userID, db); err != nil {
			server.LogWarn(r.Context(),
				"Error updating follow stats", err,
				"tags", []string{"FOLLOWS", "DB"},
			)
		}
		if err := user.UpdateFollowStats(req.FollowedID, db); err != nil {
			server.LogWarn(r.Context(),
				"Error updating follow stats", err,
				"tags", []string{"FOLLOWS", "DB"},
			)
		}

		response = FollowResponse{
			Success: true,
			Message: "Followed successfully",
		}

		// Step 8: Prepare notification data with social context
		var followedUser user.User
		if err := db.First(&followedUser, req.FollowedID).Error; err != nil {
			server.LogWarn(r.Context(),
				"Error getting followed user", err,
				"tags", []string{"FOLLOWS", "DB"},
			)
		}

		// Calculate shared courses for enhanced social context in notifications
		var sharedCoursesCount int64
		if err := db.Model(&course.Course{}).
			Where("user_id = ? AND code IN (SELECT code FROM courses WHERE user_id = ?)", userID, req.FollowedID).
			Count(&sharedCoursesCount).Error; err != nil {
			server.LogWarn(r.Context(),
				"Error counting shared courses", err,
				"tags", []string{"FOLLOWS", "DB"},
			)
		}

		/*if sseServer != nil {
			sseServer.SendNotification(
				req.FollowedID,
				userID,
				models.EntityFollow,
				req.FollowedID,
				notifications.NotificationFollow,
				currentUser.Username,
				fmt.Sprintf("%s followed you. You share %d courses with this user", currentUser.Username ,sharedCoursesCount),
				"create",
				"",
			)
		}*/
	}

	// Step 9: Send successful response with follow operation result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Step 10: Log follow action for social analytics and audit trail
	server.LogInfo(r.Context(), "Follow action completed successfully",
		"followed_id", req.FollowedID,
		"action", map[bool]string{true: "unfollow", false: "follow"}[isFollowing],
		"tags", []string{"FOLLOWS", "WRITE"},
	)
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
func HandleGetFollowers(w http.ResponseWriter, r *http.Request) {
	// Step 1: Enforce GET-only endpoint for follower retrieval
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 2: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	_ = currentUser.ID // Available but not used for this endpoint

	// Step 3: Extract and validate user_id parameter
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "User ID parameter required",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
		return
	}

	// Step 4: Convert user ID string to integer for database operations
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting user ID",
			"tags", []string{"FOLLOWS", "INVALID_USER_ID"},
		)
		return
	}

	// Step 5: Parse pagination parameters with defaults
	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Step 6: Attempt to retrieve followers from Redis cache first (performance optimization)
	var cacheKey = fmt.Sprintf("followers:%d", userID)
	followersHash, err := RedisClient.HGetAll(context.Background(), cacheKey).Result()
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting followers from redis",
			"tags", []string{"FOLLOWS", "REDIS"},
		)
		return
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		server.LogInfo(r.Context(), "Followers retrieved from cache",
			"count", len(cachedFollowers),
			"tags", []string{"FOLLOWS", "REDIS", "HIT"},
		)
		return
	}

	// Step 8: Cache miss - Query followers from database and populate cache
	followers, err := user.GetFollowers(uint(userID), limit, offset, db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting followers from database",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}
	total := len(followers)

	// Step 9: Cache the followers in redis
	for _, follower := range followers {
		followerJSON, err := json.Marshal(follower)
		if err != nil {
			server.LogWarn(r.Context(),
				"Error marshalling follower to json", err,
				"tags", []string{"FOLLOWS", "MARSHALLING"},
			)
			continue
		}
		if err := RedisClient.HSet(context.Background(), cacheKey, strconv.Itoa(int(follower.ID)), followerJSON).Err(); err != nil {
			server.LogWarn(r.Context(),
				"Error caching follower in redis", err,
				"tags", []string{"FOLLOWS", "REDIS"},
			)
		}
	}

	// Step 10: Set cache expiration to 10 minutes for optimal balance of freshness and performance
	if err := RedisClient.Expire(context.Background(), cacheKey, 10*time.Minute).Err(); err != nil {
		server.LogWarn(r.Context(),
			"Error expiring followers in redis", err,
			"tags", []string{"FOLLOWS", "REDIS"},
		)
	}

	// Step 11: Send successful response with followers list and total count
	response := FollowersResponse{
		Followers: followers,
		Total:     total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Step 12: Log successful retrieval for audit trail and monitoring
	server.LogInfo(r.Context(), "Followers retrieved successfully",
		"count", total,
		"tags", []string{"FOLLOWS", "READ"},
	)
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
func HandleGetFollowing(w http.ResponseWriter, r *http.Request) {
	// Step 1: Enforce GET-only endpoint for following list retrieval
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 2: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	_ = currentUser.ID // Available but not used for this endpoint

	// Step 3: Extract and validate user_id parameter
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "User ID parameter required",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
		return
	}

	// Step 4: Convert user ID string to integer for database operations
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting user ID",
			"tags", []string{"FOLLOWS", "INVALID_USER_ID"},
		)
		return
	}

	// Step 5: Parse pagination parameters with defaults
	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Step 6: Attempt to retrieve following list from Redis cache first (performance optimization)
	var cacheKey = fmt.Sprintf("following:%d", userID)
	followingHash, err := RedisClient.HGetAll(context.Background(), cacheKey).Result()
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting following from redis",
			"tags", []string{"FOLLOWS", "REDIS"},
		)
		return
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		server.LogInfo(r.Context(), "Following retrieved from cache",
			"count", len(cachedFollowing),
			"tags", []string{"FOLLOWS", "REDIS", "HIT"},
		)
		return
	}

	// Step 8: Cache miss - Query following list from database and populate cache
	following, err := user.GetFollowing(uint(userID), limit, offset, db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting following from database",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}
	total := len(following)

	// Step 9: Cache the following list in Redis for future requests
	for _, followed := range following {
		followedJSON, err := json.Marshal(followed)
		if err != nil {
			server.LogWarn(r.Context(),
				"Error marshalling followed to json", err,
				"tags", []string{"FOLLOWS", "MARSHALLING"},
			)
			continue
		}
		if err := RedisClient.HSet(context.Background(), cacheKey, strconv.Itoa(int(followed.ID)), followedJSON).Err(); err != nil {
			server.LogWarn(r.Context(),
				"Error caching followed in redis", err,
				"tags", []string{"FOLLOWS", "REDIS"},
			)
		}
	}

	// Step 10: Set cache expiration to 10 minutes for optimal balance of freshness and performance
	if err := RedisClient.Expire(context.Background(), cacheKey, 10*time.Minute).Err(); err != nil {
		server.LogWarn(r.Context(),
			"Error expiring following in redis", err,
			"tags", []string{"FOLLOWS", "REDIS"},
		)
	}

	// Step 11: Send successful response with following list and total count
	response := FollowingResponse{
		Following: following,
		Total:     total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Step 12: Log successful retrieval for audit trail and monitoring
	server.LogInfo(r.Context(), "Following retrieved successfully",
		"count", total,
		"tags", []string{"FOLLOWS", "READ"},
	)
}

// HandleGetFollowStatus retrieves follow relationship status and statistics between users.
// Provides comprehensive information for displaying follow buttons and user profile statistics.
// Returns follow relationship status along with follower/following counts for the target user.
//
// HTTP Method: GET
// Content-Type: Not required (query parameters used)
//
// Query Parameters:
//   - user_id: ID of the target user to check follow status against (string, required)
//
// Response (200 OK):
//   - is_following: Whether current user follows the target user (boolean)
//   - followers_count: Number of users following the target user (int)
//   - following_count: Number of users the target user follows (int)
//
// Authentication: Required (AuthMiddleware) - extracts current user from JWT token
//
// Database Operations:
//   - Checks follow relationship between current user and target user
//   - Retrieves follower count for target user
//   - Retrieves following count for target user
//   - No caching implemented (real-time status for UI accuracy)
//
// Use Cases:
//   - Displaying follow/unfollow buttons in user interfaces
//   - Showing user profile statistics (follower/following counts)
//   - Social feed relationship indicators
//   - User discovery and recommendation systems
//
// Error Responses:
//   - 400 Bad Request: Missing user_id parameter or invalid format
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 405 Method Not Allowed: Non-GET request
//   - 500 Internal Server Error: Database query failures
//
// Side Effects:
//   - Logs follow status checks for social analytics
//   - No database modifications (read-only operation)
//   - No caching (ensures real-time accuracy for UI state)
func HandleGetFollowStatus(w http.ResponseWriter, r *http.Request) {
	// Step 1: Enforce GET-only endpoint for follow status retrieval
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 2: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	currentUserID := currentUser.ID

	// Step 3: Extract and validate target user_id parameter
	targetUserIDStr := r.URL.Query().Get("user_id")
	if targetUserIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "User ID parameter required",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
		return
	}

	// Step 4: Convert target user ID string to integer for database operations
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting user ID",
			"tags", []string{"FOLLOWS", "INVALID_USER_ID"},
		)
		return
	}

	// Step 5: Check if current user follows the target user
	isFollowing, err := user.IsFollowing(currentUserID, uint(targetUserID), db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error checking follow status",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}

	// Step 6: Get follower count for the target user
	followersCount, err := user.GetFollowersCount(uint(targetUserID), db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting followers count",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}

	// Step 7: Get following count for the target user
	followingCount, err := user.GetFollowingCount(uint(targetUserID), db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting following count",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}

	// Step 8: Construct comprehensive follow status response
	response := FollowStatusResponse{
		IsFollowing:    isFollowing,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
	}

	// Step 9: Send successful response with follow status and statistics
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Step 10: Log successful status retrieval for social analytics and monitoring
	server.LogInfo(r.Context(), "Follow status retrieved successfully",
		"target_user_id", targetUserID,
		"is_following", isFollowing,
		"followers_count", followersCount,
		"following_count", followingCount,
		"tags", []string{"FOLLOWS", "READ"},
	)
}
