package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	var courses []course.Course
	if err := db.Where("user_id = ?", userID).Find(&courses).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting assignment for user id = %d : %s", userID, err))
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

func CreateCourseHandler(w http.ResponseWriter, r *http.Request) {

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
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate all required fields
	if input.LocalID == "" || input.Code == "" || input.Semester == "" || input.Instructor == "" || input.StartDate == "" || input.EndDate == "" {
		server.PrintERROR(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	start_date, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, "Invalid start date format")
		return
	}

	end_date, err := time.Parse(time.DateOnly, input.EndDate)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, "Invalid start date format")
		return
	}

	credits, err := strconv.Atoi(input.Credits)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error formating credits : %s", err))

		return
	}
	local_id, err := strconv.Atoi(input.LocalID)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error formating local_id : %s", err))

		return
	}

	cVal := course.Course{
		UserID:          userID,
		LocalID:         uint(local_id),
		Name:            input.Name,
		Code:            input.Code,
		Color:           input.Color,
		Semester:        input.Semester,
		Schedule:        input.Schedule,
		Credits:         credits,
		Location:        input.Location,
		StartDate:       start_date,
		EndDate:         end_date,
		Instructor:      input.Instructor,
		InstructorEmail: input.InstructorEmail,
	}

	result := tx.Create(&cVal)
	if result.Error != nil {
		server.PrintERROR(w, http.StatusConflict, fmt.Sprintf("Error creating assignment in database", err))
		return
	}

	cObj := &cVal

	c, err := course.Get_Course_byId(cObj.ID, tx)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting course: %s", err))
		return
	}

	// Convert to map safely
	courseMap := c.ToMap()
	if courseMap == nil {
		tx.Rollback()
		server.PrintERROR(w, http.StatusInternalServerError, "Failed to process course data")
		return
	}

	tx.Commit()

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Assignment created successfully",
		"course":  courseMap,
	})

	server.PrintLOG([]string{"SUCCESS", "CREATE", "COURSE"}, fmt.Sprintf("course created successfully : %v", courseMap))

}
func UpdateCourseHandler(w http.ResponseWriter, r *http.Request) {

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

	c, err := course.Get_Course_byId(uint(int_id), tx)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting course: %s", err))
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE courses SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), c.ID).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error updating assignment in database: %s", err))
		return
	}

	server.PrintLOG([]string{"SUCCESS", "UPDATE", "COURSE"}, fmt.Sprintf("user_id %s course %d column %s value %s",
		userIDVal, c.ID, updateData.Column, updateData.Value))

	tx.Commit()

}

func LinkRequestCourseHandler(w http.ResponseWriter, r *http.Request) {

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

	var currentUser user.User
	if err := db.First(&currentUser, userID).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database error")
		return
	}

	var linkRequestData struct {
		CourseCode string `json:"course_code"`
		UsersID    []uint `json:"users_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&linkRequestData)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	// 1. Get send course informations
	c, err := course.Get_Course_byCode(linkRequestData.CourseCode, userID, db)

	//2. Create an uuid for the link

	var linkId uuid.UUID
	if c.LinkID == uuid.Nil {
		linkId = uuid.New()
		c.LinkID = linkId
		if err = db.Save(&c).Error; err != nil {
			server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Could not save uuid %s", err))
			return

		}
	} else {
		linkId = c.LinkID
	}

	cJson, err := json.Marshal(c)
	if err != nil {
		log.Printf("[Error] error marshalling notification : %v ", err)

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
	// Infos : Course All except user_id, Sender (name)

	server.PrintLOG([]string{"SUCCESS", "LINK", "COURSE"}, fmt.Sprintf("Course ID : %v, Users length : %v, Link ID : %v", linkRequestData.CourseCode, len(linkRequestData.UsersID), linkId))

}

func AcceptLinkCourseHandler(w http.ResponseWriter, r *http.Request) {

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

	var currentUser user.User
	if err := db.First(&currentUser, userID).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database error")
		return
	}

	var c course.Course
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	//1. Get Course assignments
	var courseAssignments []assignment.Assignment
	err = db.Where("user_id = ? AND course_code = ?", c.UserID, c.Code).Order("created_at").Find(&courseAssignments).Error
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error getting course assignments with course code : %v ", err))
		return
	}

	// 2. list assignments id
	var responseAssignments []assignment.Assignment
	for _, assignment := range courseAssignments {
		assignmentDocuments, err := assignment.GetDocuments(db.Debug())

		if err != nil {
			server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error getting assignment %v documents: %v ", assignment.ID, err))
			return
		}

		assignment.Documents = assignmentDocuments
		responseAssignments = append(responseAssignments, assignment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Error":       err,
		"assignments": responseAssignments,
		//"documents": assignmentDocuments,
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
				Data:     string(""),
			},
		)

	}

	server.PrintLOG([]string{"SUCCESS", "ACCEPT", "COURSE"}, fmt.Sprintf("Course ID : %v, From : %v, To: %v", c.Code, c.UserID, userID))

}
