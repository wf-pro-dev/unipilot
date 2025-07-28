package server

import (
	"encoding/json"
	"net/http"

	"unipilot/internal/models/user"

	"gorm.io/gorm"
)

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Safely get context values
	dbVal := r.Context().Value("db")
	if dbVal == nil {
		PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	// Query user
	var userObj user.User
	if err := db.First(&userObj, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			PrintERROR(w, http.StatusNotFound, "User not found")
		} else {
			PrintERROR(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Convert to map safely
	userMap := userObj.ToMap()
	if userMap == nil {
		PrintERROR(w, http.StatusInternalServerError, "Failed to process user data")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User retrieved successfully",
		"user":    userMap,
	})
}
