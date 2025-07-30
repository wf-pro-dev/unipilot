package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"unipilot/internal/models/course"

	"gorm.io/gorm"
)

func GetCourseHandler(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	var courses []course.Course
	if err := db.Where("user_id = ?", userID).Find(&courses).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting assignment for user id = %d : %s", userID, err))
		return
	}

	var coursesMap []map[string]string
	for _, a := range courses {
		coursesMap = append(coursesMap, a.ToMap())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User's Assignments retrieved successfully",
		"courses": coursesMap,
	})
}

func UpdateCourseHandler(w http.ResponseWriter, r *http.Request) {

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	/*userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}*/

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var updateData struct {
		ID     string `json:"id"`
		Value  string `json:"value"`
		Column string `json:"column`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	int_id, err := strconv.Atoi(updateData.ID)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to convert assignment ID to int: %s", err))
		return
	}

	a, err := course.Get_Course_from_Local(uint(int_id), tx)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting course: %s", err))
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE courses SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), a.ID).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error updating assignment in database: %s", err))
		return
	}

	/*value := updateData.Value

	//PrintLog(fmt.Sprintf("column : %s, value :%s, user id:%s", updateData.Column, value, dbVal.(string) ))
	if updateData.Column == "course_code" {
		//PrintLog(fmt.Sprintf("column : %s, value :%s, user id:%s", updateData.Column, value, dbVal.(string) ))

		c, err := course.Get_Course_byCode(value, strconv.Itoa(int(userID)), tx)
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting new course: %s", err))
			return
		}
		PrintLog(fmt.Sprintf("Course %s", c.ToMap()))

		value = c.NotionID
		PrintLog(fmt.Sprintf("Course notion id: %s", value))
	}

	var obj map[string]string

	if updateData.Column == "status_name" {
		var status = models.Get_AssignmentStatus_byName(value, tx)
		obj = status.ToMap()
	} else if updateData.Column == "type_name" {
		var t = models.Get_AssignmentType_byName(value, tx)
		obj = t.ToMap()
	}

	err = a.Update_Notion(updateData.Column, value, obj)
	if err != nil {
		tx.Rollback()
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error updating assignment in notion", err))
		return
	}*/

	PrintLog(fmt.Sprintf("user_id %s column %s value %s",
		userIDVal, updateData.Column, updateData.Value))

	tx.Commit()

}
