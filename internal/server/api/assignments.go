package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"unipilot/internal/models/assignment"

	"unipilot/internal/server"

	"gorm.io/gorm"
)

func GetAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
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

	var assignments []assignment.Assignment
	if err := db.Where("user_id = ?", userID).Find(&assignments).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting assignment for user id = %d : %s", userID, err))
		return
	}

	var assignmentsMap []map[string]string
	for _, a := range assignments {
		assignmentsMap = append(assignmentsMap, a.ToMap())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "User's Assignments retrieved successfully",
		"assignments": assignmentsMap,
	})

}
func CreateAssignmentHandler(w http.ResponseWriter, r *http.Request) {

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
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

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var input struct {
		LocalID    string `json:"local_id"`
		Title      string `json:"title"`
		Todo       string `json:"todo"`
		Deadline   string `json:"deadline"`
		CourseCode string `json:"course_code"`
		TypeName   string `json:"type"`
		StatusName string `json:"status"`
		Priority   string `json:"priority"`
		Link       string `json:"link"`
		ParentID   string `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	// Validate all required fields
	if input.LocalID == "" || input.CourseCode == "" || input.Title == "" || input.TypeName == "" || input.Deadline == "" {
		server.PrintERROR(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	deadline, err := time.Parse(time.DateOnly, input.Deadline)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, "Invalid deadline format")
		return
	}

	local_id, err := strconv.Atoi(input.LocalID)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error formating local_id : %s", err))
		return
	}

	var parent_id = 0
	if input.ParentID != "" {

		parent_id, err = strconv.Atoi(input.ParentID)
		if err != nil {
			server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error formating parent_id : %s", err))
			return
		}
	}

	aVal := assignment.Assignment{
		Title:      input.Title,
		UserID:     userID,
		LocalID:    uint(local_id),
		Todo:       input.Todo,
		Deadline:   deadline,
		CourseCode: input.CourseCode,
		TypeName:   input.TypeName,
		StatusName: input.StatusName,
		Priority:   input.Priority,
		Link:       input.Link,
		ParentID:   uint(parent_id),
	}

	result := tx.Create(&aVal)
	if result.Error != nil {
		server.PrintERROR(w, http.StatusConflict, fmt.Sprintf("Error creating assignment in database", err))
		return
	}

	aObj := &aVal

	a, err := assignment.Get_Assignment_byID(aObj.ID, userID, tx)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting assignment: %s", err))
		return
	}

	// Convert to map safely
	assignmentMap := a.ToMap()
	if assignmentMap == nil {
		tx.Rollback()
		server.PrintERROR(w, http.StatusInternalServerError, "Failed to process assignment data")
		return
	}

	tx.Commit()

	// Send a notification to all the users linked

	newA, err := assignment.Get_Assignment_byID(aObj.ID, userID, db)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting assignment: %s", err))
		return
	}

	// gRPC -> SSE logic : TO BE MOVE IN DOCKER

	// aJson, err := json.Marshal(newA)
	_, err = json.Marshal(newA)
	if err != nil {
		log.Printf("[Error] error marshalling notification : %v ", err)
	}

	// link_users, err := newA.Course.GetLinkUsers(db)
	_, err = newA.Course.GetLinkUsers(db)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting users link to course assignment: %s", err))
		return
	}

	// PrintLog(fmt.Sprintf("link users : %v, sseServer : %v", link_users, sseServer != nil))
	// if sseServer != nil {
	// 	// 2. Send link info to users via SSE (field data)
	// 	for _,sendeeID := range link_users {
	// 		if sendeeID != userID {
	// 			PrintLog(fmt.Sprintf("sending to : %d ",sendeeID))
	// 			sseServer.SendNotification(
	// 				uint(sendeeID),
	// 				userID,
	// 				models.EntityAssignment,
	// 				newA.Course.ID,
	// 				notifications.NotificationAssignmentUpdate,
	// 				newA.Title,
	// 				fmt.Sprintf("%s shared a new assignment on %s", newA.User.Username, newA.CourseCode),
	// 				"assignment",
	// 				string(aJson),

	// 			)
	// 		}
	// 	}
	// }

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Assignment created successfully",
		"assignment": assignmentMap,
	})

}
func UpdateAssignmentHandler(w http.ResponseWriter, r *http.Request) {

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

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

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
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	int_id, err := strconv.Atoi(updateData.ID)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to convert assignment ID to int: %s", err))
		return
	}

	a, err := assignment.Get_Assignment_byID(uint(int_id), userID, tx)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting assignment: %s", err))
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE assignments SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), a.ID).Error; err != nil {

		server.PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error updating assignment in database: %s", err))
		return
	}

	server.PrintLOG([]string{"SUCCESS", "UPDATE", "ASSIGNMENT"}, fmt.Sprintf("user_id %d assignment %d column %s value %s",
		userIDVal, a.ID, updateData.Column, updateData.Value))

	tx.Commit()

}
