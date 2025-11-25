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

// FollowRequest represents a follow request
type FollowRequest struct {
	FollowedID uint `json:"followed_id"`
}

// FollowResponse represents a follow response
type FollowResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// FollowersResponse represents a followers list response
type FollowersResponse struct {
	Followers []user.User `json:"followers"`
	Total     int         `json:"total"`
}

// FollowingResponse represents a following list response
type FollowingResponse struct {
	Following []user.User `json:"following"`
	Total     int         `json:"total"`
}

// FollowStatusResponse represents follow status response
type FollowStatusResponse struct {
	IsFollowing    bool `json:"is_following"`
	FollowersCount int  `json:"followers_count"`
	FollowingCount int  `json:"following_count"`
}

// HandleFollow handles follow/unfollow requests
func HandleFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
		return
	}

	if req.FollowedID == 0 {
		err := fmt.Errorf("invalid followed_id")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid followed_id",
			"tags", []string{"FOLLOWS", "VALIDATION"},
		)
		return
	}

	if userID == req.FollowedID {
		err := fmt.Errorf("cannot follow yourself")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Cannot follow yourself",
			"tags", []string{"FOLLOWS", "VALIDATION"},
		)
		return
	}

	// Check if already following
	isFollowing, err := user.IsFollowing(userID, req.FollowedID, db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error checking follow status",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}

	var response FollowResponse
	if isFollowing {
		// Unfollow
		if err := user.RemoveFollow(userID, req.FollowedID, db); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error removing follow",
				"tags", []string{"FOLLOWS", "DB"},
			)
			return
		}

		// Update stats for both users
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
		// Follow
		if err := user.CreateFollow(userID, req.FollowedID, db); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error creating follow",
				"tags", []string{"FOLLOWS", "DB"},
			)
			return
		}

		// Update stats for both users
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

		var followedUser user.User
		if err := db.First(&followedUser, req.FollowedID).Error; err != nil {
			server.LogWarn(r.Context(),
				"Error getting followed user", err,
				"tags", []string{"FOLLOWS", "DB"},
			)
		}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	server.LogInfo(r.Context(), "Follow action completed successfully",
		"followed_id", req.FollowedID,
		"action", map[bool]string{true: "unfollow", false: "follow"}[isFollowing],
		"tags", []string{"FOLLOWS", "WRITE"},
	)
}

// HandleGetFollowers handles getting followers list
func HandleGetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	_ = currentUser.ID // Available but not used for this endpoint

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "User ID parameter required",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting user ID",
			"tags", []string{"FOLLOWS", "INVALID_USER_ID"},
		)
		return
	}

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

	// Try to get followers from Redis cache first
	var cacheKey = fmt.Sprintf("followers:%d", userID)
	followersHash, err := RedisClient.HGetAll(context.Background(), cacheKey).Result()
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting followers from redis",
			"tags", []string{"FOLLOWS", "REDIS"},
		)
		return
	}
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

	followers, err := user.GetFollowers(uint(userID), limit, offset, db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting followers from database",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}
	total := len(followers)

	// Cache the followers in redis
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

	if err := RedisClient.Expire(context.Background(), cacheKey, 10*time.Minute).Err(); err != nil {
		server.LogWarn(r.Context(),
			"Error expiring followers in redis", err,
			"tags", []string{"FOLLOWS", "REDIS"},
		)
	}

	response := FollowersResponse{
		Followers: followers,
		Total:     total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	server.LogInfo(r.Context(), "Followers retrieved successfully",
		"count", total,
		"tags", []string{"FOLLOWS", "READ"},
	)
}

// HandleGetFollowing handles getting following list
func HandleGetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	_ = currentUser.ID // Available but not used for this endpoint

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "User ID parameter required",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting user ID",
			"tags", []string{"FOLLOWS", "INVALID_USER_ID"},
		)
		return
	}

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

	// Try to get following from Redis cache first
	var cacheKey = fmt.Sprintf("following:%d", userID)
	followingHash, err := RedisClient.HGetAll(context.Background(), cacheKey).Result()
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting following from redis",
			"tags", []string{"FOLLOWS", "REDIS"},
		)
		return
	}
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

	following, err := user.GetFollowing(uint(userID), limit, offset, db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting following from database",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}
	total := len(following)

	// Cache the following in redis
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

	if err := RedisClient.Expire(context.Background(), cacheKey, 10*time.Minute).Err(); err != nil {
		server.LogWarn(r.Context(),
			"Error expiring following in redis", err,
			"tags", []string{"FOLLOWS", "REDIS"},
		)
	}

	response := FollowingResponse{
		Following: following,
		Total:     total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	server.LogInfo(r.Context(), "Following retrieved successfully",
		"count", total,
		"tags", []string{"FOLLOWS", "READ"},
	)
}

// HandleGetFollowStatus handles getting follow status between two users
func HandleGetFollowStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	currentUserID := currentUser.ID

	targetUserIDStr := r.URL.Query().Get("user_id")
	if targetUserIDStr == "" {
		err := fmt.Errorf("user_id parameter required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "User ID parameter required",
			"tags", []string{"FOLLOWS", "REQUEST"},
		)
		return
	}

	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting user ID",
			"tags", []string{"FOLLOWS", "INVALID_USER_ID"},
		)
		return
	}

	isFollowing, err := user.IsFollowing(currentUserID, uint(targetUserID), db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error checking follow status",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}

	followersCount, err := user.GetFollowersCount(uint(targetUserID), db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting followers count",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}

	followingCount, err := user.GetFollowingCount(uint(targetUserID), db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting following count",
			"tags", []string{"FOLLOWS", "DB"},
		)
		return
	}

	response := FollowStatusResponse{
		IsFollowing:    isFollowing,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	server.LogInfo(r.Context(), "Follow status retrieved successfully",
		"target_user_id", targetUserID,
		"is_following", isFollowing,
		"followers_count", followersCount,
		"following_count", followingCount,
		"tags", []string{"FOLLOWS", "READ"},
	)
}
