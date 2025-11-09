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

// getCurrentUserID gets the current user ID from the request context
func getCurrentUserID(r *http.Request) (uint, error) {
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		return 0, fmt.Errorf("user_id not found in context")
	}
	return userID, nil
}

// HandleFollow handles follow/unfollow requests
func HandleFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user ID from session
	userID, err := getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FollowedID == 0 {
		http.Error(w, "Invalid followed_id", http.StatusBadRequest)
		return
	}

	if userID == req.FollowedID {
		http.Error(w, "Cannot follow yourself", http.StatusBadRequest)
		return
	}

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		server.PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	// Check if already following
	isFollowing, err := user.IsFollowing(userID, req.FollowedID, db)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database error")
		return
	}

	var response FollowResponse
	if isFollowing {
		// Unfollow
		if err := user.RemoveFollow(userID, req.FollowedID, db); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
		}

		// Update stats for both users
		if err := user.UpdateFollowStats(userID, db); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
		}
		if err := user.UpdateFollowStats(req.FollowedID, db); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
		}

		response = FollowResponse{
			Success: true,
			Message: "Unfollowed successfully",
		}
	} else {
		// Follow
		if err := user.CreateFollow(userID, req.FollowedID, db); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
		}

		// Update stats for both users
		if err := user.UpdateFollowStats(userID, db); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
		}
		if err := user.UpdateFollowStats(req.FollowedID, db); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
		}

		response = FollowResponse{
			Success: true,
			Message: "Followed successfully",
		}

		var currentUser user.User
		if err := db.First(&currentUser, userID).Error; err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
		}
		var followedUser user.User
		if err := db.First(&followedUser, req.FollowedID).Error; err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
		}

		var sharedCoursesCount int64
		if err := db.Model(&course.Course{}).
			Where("user_id = ? AND code IN (SELECT code FROM courses WHERE user_id = ?)", userID, req.FollowedID).
			Count(&sharedCoursesCount).Error; err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
			return
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

	server.PrintLOG([]string{"SUCCESS", "FOLLOW"}, fmt.Sprintf("User ID : %v, Followed ID : %v", userID, req.FollowedID))
}

// HandleGetFollowers handles getting followers list
func HandleGetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id parameter required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
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

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		server.PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	// Try to get followers from Redis cache first
	var cacheKey = fmt.Sprintf("followers:%d", userID)
	followersHash, err := RedisClient.HGetAll(context.Background(), cacheKey).Result()
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting followers from redis: %v", err))
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
		server.PrintLOG([]string{"INFO", "FOLLOW", "GET", "REDIS"}, fmt.Sprintf("Followers retrieved from cache for user id: %d", userID))
		return
	}

	followers, err := user.GetFollowers(uint(userID), limit, offset, db)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	total := len(followers)

	// Cache the followers in redis
	for _, follower := range followers {
		followerJSON, err := json.Marshal(follower)
		if err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling follower to json: %v", err))
			return
		}
		if err := RedisClient.HSet(context.Background(), cacheKey, strconv.Itoa(int(follower.ID)), followerJSON).Err(); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error caching follower in redis: %v", err))
			return
		}
	}

	if err := RedisClient.Expire(context.Background(), cacheKey, 10*time.Minute).Err(); err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error expiring followers in redis: %v", err))
		return
	}

	server.PrintLOG([]string{"INFO", "FOLLOW", "GET", "REDIS"}, fmt.Sprintf("Followers cached successfully for user id: %d", userID))

	response := FollowersResponse{
		Followers: followers,
		Total:     len(followers),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	server.PrintLOG([]string{"SUCCESS", "GET", "FOLLOW"}, fmt.Sprintf("User ID : %d, Nb of followers : %d", userID, total))
}

// HandleGetFollowing handles getting following list
func HandleGetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id parameter required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
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

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		server.PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	// Try to get following from Redis cache first
	var cacheKey = fmt.Sprintf("following:%d", userID)
	followingHash, err := RedisClient.HGetAll(context.Background(), cacheKey).Result()
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting following from redis: %v", err))
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
		server.PrintLOG([]string{"INFO", "FOLLOW", "GET", "REDIS"}, fmt.Sprintf("Following retrieved from cache for user id: %d", userID))
		return
	}

	following, err := user.GetFollowing(uint(userID), limit, offset, db)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	total := len(following)

	// Cache the followers in redis
	for _, followed := range following {
		followedJSON, err := json.Marshal(followed)
		if err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling followed to json: %v", err))
			return
		}
		if err := RedisClient.HSet(context.Background(), cacheKey, strconv.Itoa(int(followed.ID)), followedJSON).Err(); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error caching followed in redis: %v", err))
			return
		}
	}

	if err := RedisClient.Expire(context.Background(), cacheKey, 10*time.Minute).Err(); err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error expiring following in redis: %v", err))
		return
	}

	server.PrintLOG([]string{"INFO", "FOLLOW", "GET", "REDIS"}, fmt.Sprintf("Followers cached successfully for user id: %d", userID))

	response := FollowingResponse{
		Following: following,
		Total:     len(following),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	server.PrintLOG([]string{"SUCCESS", "GET", "FOLLOW"}, fmt.Sprintf("User ID : %d, Nb of following : %d", userID, total))
}

// HandleGetFollowStatus handles getting follow status between two users
func HandleGetFollowStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user ID from session
	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetUserIDStr := r.URL.Query().Get("user_id")
	if targetUserIDStr == "" {
		http.Error(w, "user_id parameter required", http.StatusBadRequest)
		return
	}

	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}
	dbVal := r.Context().Value("db")
	if dbVal == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		server.PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	isFollowing, err := user.IsFollowing(currentUserID, uint(targetUserID), db)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	followersCount, err := user.GetFollowersCount(uint(targetUserID), db)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	followingCount, err := user.GetFollowingCount(uint(targetUserID), db)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := FollowStatusResponse{
		IsFollowing:    isFollowing,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	server.PrintLOG([]string{"SUCCESS", "GET", "FOLLOW"}, fmt.Sprintf("User ID : %d, Target User ID : %d, Is Following : %v, Followers Count : %d, Following Count : %d", currentUserID, targetUserID, isFollowing, followersCount, followingCount))
}
