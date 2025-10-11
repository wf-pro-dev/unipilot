package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"

	"unipilot/internal/models/user"
	"unipilot/internal/server"
)

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Safely get context values
	dbVal := r.Context().Value("db")
	if dbVal == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		server.PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	// Query user
	var userObj user.User
	if err := db.First(&userObj, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			server.PrintERROR(w, http.StatusNotFound, "User not found")
		} else {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Convert to map safely
	userMap := userObj.ToMap()
	if userMap == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Failed to process user data")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User retrieved successfully",
		"user":    userMap,
	})
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Safely get context values
	dbVal := r.Context().Value("db")
	if dbVal == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		server.PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	if err := db.Exec(fmt.Sprintf("UPDATE users SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), userID).Error; err != nil {

		server.PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error updating assignment in database: %s", err))
		return
	}

	// Query user
	var userObj user.User
	if err := db.First(&userObj, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			server.PrintERROR(w, http.StatusNotFound, "User not found")
		} else {
			server.PrintERROR(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Convert to map safely
	userMap := userObj.ToMap()
	if userMap == nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Failed to process user data")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User retrieved successfully",
		"user":    userMap,
	})

	server.PrintLOG([]string{"SUCCESS", "UPDATE", "USER"}, fmt.Sprintf("user_id %d column %s value %s",
		userIDVal, updateData.Column, updateData.Value))
}
