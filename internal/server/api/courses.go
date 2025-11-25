package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	notif "unipilot/internal/models/notifications"
	"unipilot/internal/server/sse/grpc/notifications"

	//"unipilot/internal/models/document"

	"unipilot/internal/models/user"
	"unipilot/internal/server"

	"gorm.io/gorm"
)

func GetCourseHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	var courses []course.Course
	if err := db.Where("user_id = ?", userID).Find(&courses).Error; err != nil {
		server.ResponseError(w, err, http.StatusInternalServerError, "Error getting courses from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	var coursesMap []map[string]string
	for _, a := range courses {
		coursesMap = append(coursesMap, a.ToMap())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User's Courses retrieved successfully",
		"courses": coursesMap,
	})

	server.LogInfo("Courses retrieved successfully",
		"request_id", requestID,
		"user_id", userID,
		"count", len(coursesMap),
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"COURSES", "READ"},
	)
}

func CreateCourseHandler(w http.ResponseWriter, r *http.Request) {
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
		LocalID         string `json:"local_id"`
		Name            string `json:"name"`
		Code            string `json:"code"`
		Color           string `json:"color"`
		Semester        string `json:"semester"`
		Schedule        string `json:"schedule"`
		Credits         string `json:"credits"`
		Location        string `json:"location"`
		StartDate       string `json:"start_date"`
		EndDate         string `json:"end_date"`
		Instructor      string `json:"instructor"`
		InstructorEmail string `json:"instructor_email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
		return
	}

	// Validate all required fields
	if input.LocalID == "" || input.Code == "" || input.Semester == "" || input.Instructor == "" || input.StartDate == "" || input.EndDate == "" {
		err := errors.New("missing required fields")
		server.ResponseError(w, err, http.StatusBadRequest, "Missing required fields",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "MISSING_REQUIRED_FIELDS"},
		)
		return
	}

	startDate, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid start date format",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_START_DATE"},
		)
		return
	}

	endDate, err := time.Parse(time.DateOnly, input.EndDate)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid end date format",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_END_DATE"},
		)
		return
	}

	credits, err := strconv.Atoi(input.Credits)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Error formatting credits",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_CREDITS"},
		)
		return
	}
	localID, err := strconv.Atoi(input.LocalID)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Error formatting local_id",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_LOCAL_ID"},
		)
		return
	}

	cVal := course.Course{
		UserID:          userID,
		LocalID:         uint(localID),
		Name:            input.Name,
		Code:            input.Code,
		Color:           input.Color,
		Semester:        input.Semester,
		Schedule:        input.Schedule,
		Credits:         credits,
		Location:        input.Location,
		StartDate:       startDate,
		EndDate:         endDate,
		Instructor:      input.Instructor,
		InstructorEmail: input.InstructorEmail,
	}

	if result := tx.Create(&cVal); result.Error != nil {
		tx.Rollback()
		server.ResponseError(w, result.Error, http.StatusConflict, "Error creating course in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	c, err := course.Get_Course_byId(cVal.ID, tx)
	if err != nil {
		tx.Rollback()
		server.ResponseError(w, err, http.StatusInternalServerError, "Error getting course from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	// Convert to map safely
	courseMap := c.ToMap()
	if courseMap == nil {
		tx.Rollback()
		err := errors.New("failed to process course data")
		server.ResponseError(w, err, http.StatusInternalServerError, "Error processing course data",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "MARSHALLING"},
		)
		return
	}

	tx.Commit()

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Course created successfully",
		"course":  courseMap,
	})

	server.LogInfo("Course created successfully",
		"request_id", requestID,
		"user_id", userID,
		"course_id", cVal.ID,
		"code", cVal.Code,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"COURSES", "WRITE"},
	)
}
func UpdateCourseHandler(w http.ResponseWriter, r *http.Request) {
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
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
		return
	}

	intID, err := strconv.Atoi(updateData.ID)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Error converting course ID to int",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_COURSE_ID"},
		)
		return
	}

	c, err := course.Get_Course_byId(uint(intID), tx)
	if err != nil {
		tx.Rollback()
		server.ResponseError(w, err, http.StatusInternalServerError, "Error getting course from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE courses SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), c.ID).Error; err != nil {
		tx.Rollback()
		server.ResponseError(w, err, http.StatusInternalServerError, "Error updating course in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	tx.Commit()

	server.LogInfo("Course updated successfully",
		"request_id", requestID,
		"user_id", userID,
		"course_id", c.ID,
		"update", updateData,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"COURSES", "WRITE"},
	)
}

func LinkRequestCourseHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	var linkRequestData struct {
		CourseCode string `json:"course_code"`
		UsersID    []uint `json:"users_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&linkRequestData); err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
		return
	}

	// 1. Get send course informations
	c, err := course.Get_Course_byCode(linkRequestData.CourseCode, userID, db)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Error getting course by code",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	//2. Create an uuid for the link

	var linkId uuid.UUID
	if c.LinkID == uuid.Nil {
		linkId = uuid.New()
		c.LinkID = linkId
		if err = db.Save(&c).Error; err != nil {
			server.ResponseError(w, err, http.StatusBadRequest, "Error saving link identifier",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"COURSES", "DB"},
			)
			return

		}
	} else {
		linkId = c.LinkID
	}

	cJson, err := json.Marshal(c)
	if err != nil {
		server.LogWarn(
			"Error marshalling notification payload", err,
			"request_id", requestID,
			"user_id", userID,
			"course_id", c.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "MARSHALLING"},
		)
	}

	if GrpcClient != nil {

		// 2. Send link info to users via SSE (field data)
		for _, sendeeID := range linkRequestData.UsersID {
			GrpcClient.SendNotification(context.Background(),
				&notifications.Notification{
					UserId:   uint32(sendeeID),
					SenderId: uint32(userID),
					Entity:   string(models.EntityCourse),
					EntityId: uint32(c.ID),
					Type:     string(notif.NotificationSync),
					Title:    c.Name,
					Message:  fmt.Sprintf("%s shared a course with you : %s", currentUser.Username, c.Code),
					Action:   "sync",
					Data:     string(cJson),
				},
			)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Course link request processed",
		"course_id":  c.ID,
		"link_id":    linkId,
		"recipients": linkRequestData.UsersID,
	})

	server.LogInfo("Course link request processed",
		"request_id", requestID,
		"user_id", userID,
		"course_id", c.ID,
		"recipients_count", len(linkRequestData.UsersID),
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"COURSES", "LINK", "REQUEST"},
	)
}

func AcceptLinkCourseHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	var c course.Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
		return
	}

	//1. Get Course assignments
	var courseAssignments []assignment.Assignment
	if err := db.Where("user_id = ? AND course_code = ?", c.UserID, c.Code).Order("created_at").Find(&courseAssignments).Error; err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Error getting course assignments",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	// 2. list assignments id
	var responseAssignments []assignment.Assignment
	for _, assignment := range courseAssignments {
		assignmentDocuments, err := assignment.GetDocuments(db)

		if err != nil {
			server.ResponseError(w, err, http.StatusBadRequest, "Error getting assignment documents",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"COURSES", "DB"},
			)
			return
		}

		assignment.Documents = assignmentDocuments
		responseAssignments = append(responseAssignments, assignment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"assignments": responseAssignments,
	})

	if GrpcClient != nil {

		// 2. Send link info to users via SSE (field data)
		GrpcClient.SendNotification(context.Background(),
			&notifications.Notification{
				UserId:   uint32(c.UserID),
				SenderId: uint32(userID),
				Entity:   string(models.EntityCourse),
				EntityId: uint32(c.ID),
				Type:     string(notif.NotificationSync),
				Title:    c.Name,
				Message:  fmt.Sprintf("%s is now linked to your course : %s", currentUser.Username, c.Code),
				Action:   "sync",
				Data:     "",
			},
		)

	}

	server.LogInfo("Course link accepted",
		"request_id", requestID,
		"course_code", c.Code,
		"from_user", c.UserID,
		"to_user", userID,
		"assignments_synced", len(responseAssignments),
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"COURSES", "LINK", "ACCEPT"},
	)
}
