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

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)

	// Convert to map safely
	userMap := currentUser.ToMap()
	if userMap == nil {
		err := errors.New("failed to process user data")
		server.ResponseError(w, err, http.StatusInternalServerError, "Error processing user data",
			"request_id", requestID,
			"user_id", currentUser.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USER", "MARSHALLING"},
		)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User retrieved successfully",
		"user":    userMap,
	})

	server.LogInfo("User retrieved successfully",
		"request_id", requestID,
		"user_id", currentUser.ID,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"USER", "READ"},
	)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USER", "REQUEST"},
		)
		return
	}

	if err := db.Exec(fmt.Sprintf("UPDATE users SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), userID).Error; err != nil {
		server.ResponseError(w, err, http.StatusInternalServerError, "Error updating user in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USER", "DB"},
		)
		return
	}

	// Query user
	var userObj user.User
	if err := db.First(&userObj, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			server.ResponseError(w, err, http.StatusNotFound, "User not found",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"USER", "DB"},
			)
		} else {
			server.ResponseError(w, err, http.StatusInternalServerError, "Error getting user from database",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"USER", "DB"},
			)
		}
		return
	}

	// Convert to map safely
	userMap := userObj.ToMap()
	if userMap == nil {
		err := errors.New("failed to process user data")
		server.ResponseError(w, err, http.StatusInternalServerError, "Error processing user data",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USER"},
		)
		return
	}

	// Update the user in redis
	userJSON, err := json.Marshal(userMap)
	if err != nil {
		server.ResponseError(w, err, http.StatusInternalServerError, "Error marshalling user to json",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USER", "SERIALIZATION"},
		)
		return
	}
	if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(userID)), userJSON).Err(); err != nil {
		server.LogWarn(
			"Error caching user in redis", err,
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"USER", "REDIS"},
		)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User updated successfully",
		"user":    userMap,
	})

	server.LogInfo("User updated successfully",
		"request_id", requestID,
		"user_id", userID,
		"update", updateData,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"USER", "WRITE"},
	)
}
