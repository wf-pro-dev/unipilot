package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	notif "unipilot/internal/models/notifications"
	"unipilot/internal/models/user"
	"unipilot/internal/server/sse/grpc/notifications"

	"unipilot/internal/server"

	"gorm.io/gorm"
)

func GetAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	var assignments []assignment.Assignment
	if err := db.Where("user_id = ?", userID).Find(&assignments).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting assignments from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "DB"},
		)
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

	server.LogInfo(r.Context(), "Assignments retrieved successfully",
		"request_id", requestID,
		"user_id", userID,
		"count", len(assignmentsMap),
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"ASSIGNMENTS", "READ"},
	)
}
func CreateAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

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
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "REQUEST"},
		)
		return
	}
	// Validate all required fields
	if input.LocalID == "" || input.CourseCode == "" || input.Title == "" || input.TypeName == "" || input.Deadline == "" {
		server.ResponseError(r.Context(), w, fmt.Errorf("missing required fields"), http.StatusBadRequest, "Missing required fields",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "MISSING_REQUIRED_FIELDS"},
		)
		return
	}

	deadline, err := time.Parse(time.DateOnly, input.Deadline)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid deadline format",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "INVALID_DEADLINE"},
		)
		return
	}

	local_id, err := strconv.Atoi(input.LocalID)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error formatting local_id",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "INVALID_LOCAL_ID"},
		)
		return
	}

	var parent_id = 0
	if input.ParentID != "" {
		parent_id, err = strconv.Atoi(input.ParentID)
		if err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error formatting parent_id",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"ASSIGNMENTS", "INVALID_PARENT_ID"},
			)
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
		tx.Rollback()
		server.ResponseError(r.Context(), w, result.Error, http.StatusConflict, "Error creating assignment in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "DB"},
		)
		return
	}

	aObj := &aVal

	a, err := assignment.Get_Assignment_byID(aObj.ID, userID, tx)
	if err != nil {
		tx.Rollback()
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting assignment from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "DB"},
		)
		return
	}

	// Convert to map safely
	assignmentMap := a.ToMap()
	if assignmentMap == nil {
		tx.Rollback()
		server.ResponseError(r.Context(), w, fmt.Errorf("failed to process assignment data"), http.StatusInternalServerError, "Error processing assignment data",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "MARSHALLING"},
		)
		return
	}

	tx.Commit()

	// Send a notification to all the users linked

	newA, err := assignment.Get_Assignment_byID(aObj.ID, userID, db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting assignment from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "DB"},
		)
		return
	}

	aJson, err := json.Marshal(newA)
	if err != nil {
		server.LogWarn(r.Context(), "Error marshalling notification", err,
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "MARSHALLING"},
		)
	}

	link_users, err := newA.Course.GetLinkUsers(db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting users linked to course",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "DB"},
		)
		return
	}

	if GrpcClient != nil {
		for _, sendeeID := range link_users {
			if sendeeID != userID {
				GrpcClient.SendNotification(context.Background(),
					&notifications.Notification{
						UserId:   uint32(sendeeID),
						SenderId: uint32(userID),
						Entity:   string(models.EntityAssignment),
						EntityId: uint32(newA.Course.ID),
						Type:     string(notif.NotificationAssignmentUpdate),
						Title:    newA.Title,
						Message:  fmt.Sprintf("%s shared a new assignment on %s", newA.User.Username, newA.CourseCode),
						Action:   "assignment",
						Data:     string(aJson),
					},
				)
			}
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Assignment created successfully",
		"assignment": assignmentMap,
	})

	server.LogInfo(r.Context(), "Assignment created successfully",
		"request_id", requestID,
		"user_id", userID,
		"assignment_id", aObj.ID,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"ASSIGNMENTS", "WRITE"},
	)
}
func UpdateAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var updateData struct {
		ID     string `json:"id"`
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "REQUEST"},
		)
		return
	}

	int_id, err := strconv.Atoi(updateData.ID)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting assignment ID to int",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "VALIDATION"},
		)
		return
	}

	a, err := assignment.Get_Assignment_byID(uint(int_id), userID, tx)
	if err != nil {
		tx.Rollback()
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting assignment from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "DB"},
		)
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE assignments SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), a.ID).Error; err != nil {
		tx.Rollback()
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error updating assignment in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "DB"},
		)
		return
	}

	tx.Commit()

	server.LogInfo(r.Context(), "Assignment updated successfully",
		"request_id", requestID,
		"user_id", userID,
		"assignment_id", a.ID,
		"update", updateData,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"ASSIGNMENTS", "WRITE"},
	)
}
