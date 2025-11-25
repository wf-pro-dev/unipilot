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

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {

	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)

	db := r.Context().Value("db").(*gorm.DB)

	// Try to get users from Redis cache first
	usersHash, err := RedisClient.HGetAll(context.Background(), "users").Result()
	if err != nil {
		server.ResponseError(w, err, http.StatusInternalServerError, "Error getting users from redis",
			"request_id", requestID,
			"user_id", currentUser.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USERS", "REDIS"},
		)
		return
	}
	if len(usersHash) > 0 {
		// Cache hit - Convert hash to map
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
		server.LogInfo("Users retrieved from cache",
			"request_id", requestID,
			"user_id", currentUser.ID,
			"count", len(cachedUsers),
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USERS", "REDIS", "HIT"},
		)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Users retrieved successfully",
			"users":   cachedUsers,
		})
		return
	}

	// Query users from database
	var users []user.User
	if err := db.Where("id != ?", currentUser.ID).Find(&users).Error; err != nil {
		server.ResponseError(w, err, http.StatusInternalServerError, "Error getting users from database",
			"request_id", requestID,
			"user_id", currentUser.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USERS", "DB"},
		)
		return
	}

	server.FileLogger.Infow("No users data found in redis",
		"request_id", requestID,
		"user_id", currentUser.ID,
		"count", len(users),
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"USERS", "REDIS", "MISS"},
	)

	var usersMap []map[string]interface{}
	for _, u := range users {

		var courses_code []string
		if err := db.Model(&course.Course{}).Select("code").Where("user_id = ? ", u.ID).Find(&courses_code).Error; err != nil {
			server.ResponseError(w, err, http.StatusInternalServerError, "Error getting user courses",
				"request_id", requestID,
				"user_id", currentUser.ID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"USERS", "DB"},
			)
			return
		}
		u.CoursesCode = courses_code
		userMap := u.ToMap()
		usersMap = append(usersMap, userMap)

		// Cache the users in redis
		userJSON, err := json.Marshal(userMap)
		if err != nil {
			server.ResponseError(w, err, http.StatusInternalServerError, "Error marshalling user to json",
				"request_id", requestID,
				"user_id", currentUser.ID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"USERS", "SERIALIZATION"},
			)
			return
		}

		if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(u.ID)), userJSON).Err(); err != nil {
			server.LogWarn(
				"Error caching user in redis", err,
				"request_id", requestID,
				"user_id", currentUser.ID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"USERS", "REDIS"},
			)
		}
	}

	// Set TTL for users to 10 minutes
	if err := RedisClient.Expire(context.Background(), "users", time.Hour).Err(); err != nil {
		server.LogWarn(
			"Error expiring users in redis", err,
			"request_id", requestID,
			"user_id", currentUser.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USERS", "REDIS"},
		)
	}

	server.FileLogger.Infow("Users cached successfully",
		"request_id", requestID,
		"user_id", currentUser.ID,
		"count", len(usersMap),
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"USERS", "REDIS", "CACHED"},
	)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Users retrieved successfully",
		"users":   usersMap,
	})

	server.LogInfo("Users retrieved successfully",
		"request_id", requestID,
		"user_id", currentUser.ID,
		"count", len(usersMap),
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"USERS", "READ"},
	)
}
