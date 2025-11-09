package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
	"unipilot/internal/server"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Safely get context values
	db := r.Context().Value("db").(*gorm.DB)
	if db == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	currentUser := r.Context().Value("user").(user.User)
	if currentUser.ID == 0 {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Try to get users from Redis cache first
	usersHash, err := RedisClient.HGetAll(context.Background(), "users").Result()
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting users from redis: %v", err))
		return
	}
	if len(usersHash) > 0 {
		// Cache hit - Convert hash to map
		var cachedUsers []map[string]interface{}
		for _, userJSON := range usersHash {

			var userMap map[string]interface{}
			if err := json.Unmarshal([]byte(userJSON), &userMap); err == nil {
				cachedUsers = append(cachedUsers, userMap)
			}
		}
		server.PrintLOG([]string{"REDIS", "INFO"}, "Users retrieved from cache")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Users retrieved successfully",
			"users":   cachedUsers,
		})
		return
	}

	server.PrintLOG([]string{"INFO", "GET", "USERS", "REDIS"}, "No users data found in redis, querying database")

	// Query users from database
	var users []user.User
	if err := db.Find(&users).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting assignment for user id = %d : %s", currentUser.ID, err))
		return
	}

	var usersMap []map[string]interface{}
	for _, u := range users {

		var courses_code []string
		if err := db.Model(&course.Course{}).Select("code").Where("user_id = ? ", u.ID).Find(&courses_code).Error; err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting assignment for user id = %d : %s", currentUser.ID, err))
			return
		}
		//courses_code = []string{ "MATH-1414" }
		u.CoursesCode = courses_code
		userMap := u.ToMap()
		usersMap = append(usersMap, userMap)

		// Cache the users in redis
		userJSON, err := json.Marshal(userMap)
		if err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling user to json: %v", err))
			return
		}

		if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(u.ID)), userJSON).Err(); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error caching user in redis: %v", err))
			return
		}
	}

	// Set TTL for users to 10 minutes
	if err := RedisClient.Expire(context.Background(), "users", 10*time.Minute).Err(); err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error expiring users in redis: %v", err))
		return
	}

	server.PrintLOG([]string{"INFO", "USERS", "GET", "REDIS"}, fmt.Sprintf("Users cached successfully for %d users", len(usersMap)))

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Users retrieved successfully",
		"users":   usersMap,
	})

	server.PrintLOG([]string{"SUCCESS", "USERS", "GET"}, fmt.Sprintf("Users retrieved successfully for %d users", len(usersMap)))
}
