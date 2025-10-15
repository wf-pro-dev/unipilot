package server

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	// Query user
	var users []user.User
	if err := db.Not("id = ?", currentUser.ID).Find(&users).Error; err != nil {
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
		usersMap = append(usersMap, u.ToMap())
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Users retrieved successfully",
		"users":   usersMap,
	})
}
