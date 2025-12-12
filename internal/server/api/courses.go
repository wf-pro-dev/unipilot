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

// GetCourseHandler retrieves all courses belonging to the authenticated user.
// Returns a JSON response containing an array of course objects converted to maps.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request (must contain authenticated user context from AuthMiddleware)
//
// Response:
//   - 200 OK: JSON object with "message" and "courses" array
//   - 500 Internal Server Error: If database query fails
func GetCourseHandler(w http.ResponseWriter, r *http.Request) {

	// Step 1: Extract context values set by middleware (start_time, request_id, user, db)
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Query database for user's courses using parameterized query for security
	var courses []course.Course
	if err := db.Where("user_id = ?", userID).Find(&courses).Error; err != nil {
		// Handle database error with structured logging and proper HTTP status
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting courses from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	// Step 3: Transform course structs to maps for consistent JSON serialization
	var coursesMap []map[string]string
	for _, a := range courses {
		coursesMap = append(coursesMap, a.ToMap())
	}

	// Step 4: Send successful response with course data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User's Courses retrieved successfully",
		"courses": coursesMap,
	})

}

// CreateCourseHandler creates a new course for the authenticated user.
// Validates required fields, parses dates and numeric values, and stores the course in the database.
// Uses a database transaction to ensure atomicity.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request (must contain authenticated user context from AuthMiddleware)
//
// Request Body:
//   - local_id: Course local identifier (string, required)
//   - name: Course name (string)
//   - code: Course code (string, required)
//   - color: Course color (string)
//   - semester: Semester identifier (string, required)
//   - schedule: Course schedule (string)
//   - credits: Number of credits (string, will be converted to int)
//   - location: Course location (string)
//   - start_date: Start date in YYYY-MM-DD format (string, required)
//   - end_date: End date in YYYY-MM-DD format (string, required)
//   - instructor: Instructor name (string, required)
//   - instructor_email: Instructor email (string)
//
// Response:
//   - 200 OK: JSON object with "message" and created "course" object
//   - 400 Bad Request: If validation fails or date/numeric parsing fails
//   - 409 Conflict: If course creation fails (e.g., duplicate constraint)
//   - 500 Internal Server Error: If database operations fail
func CreateCourseHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values and initialize transaction for atomicity
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Begin database transaction to ensure all-or-nothing course creation
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Step 2: Define input structure for JSON unmarshaling with proper field mapping
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

	// Parse JSON request body into input struct
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
		return
	}

	// Step 3: Validate business-critical required fields
	if input.LocalID == "" || input.Code == "" || input.Semester == "" || input.Instructor == "" || input.StartDate == "" || input.EndDate == "" {
		err := errors.New("missing required fields")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Missing required fields",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "MISSING_REQUIRED_FIELDS"},
		)
		return
	}

	// Step 4: Parse and validate date formats (expects YYYY-MM-DD format)
	startDate, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid start date format",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_START_DATE"},
		)
		return
	}

	endDate, err := time.Parse(time.DateOnly, input.EndDate)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid end date format",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_END_DATE"},
		)
		return
	}

	// Step 5: Convert string numeric fields to proper integer types
	credits, err := strconv.Atoi(input.Credits)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error formatting credits",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_CREDITS"},
		)
		return
	}
	localID, err := strconv.Atoi(input.LocalID)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error formatting local_id",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_LOCAL_ID"},
		)
		return
	}

	// Step 6: Construct course object with validated and transformed data
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

	// Step 7: Persist course to database within transaction
	if result := tx.Create(&cVal); result.Error != nil {
		tx.Rollback()
		server.ResponseError(r.Context(), w, result.Error, http.StatusConflict, "Error creating course in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	// Step 8: Retrieve the created course to get auto-generated fields (ID, timestamps)
	c, err := course.Get_Course_byId(cVal.ID, tx)
	if err != nil {
		tx.Rollback()
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting course from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	// Step 9: Convert course struct to map for consistent JSON response format
	courseMap := c.ToMap()
	if courseMap == nil {
		tx.Rollback()
		err := errors.New("failed to process course data")
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error processing course data",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "MARSHALLING"},
		)
		return
	}

	// Step 10: Commit transaction after all operations succeed
	tx.Commit()

	// Step 11: Send successful response with created course data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Course created successfully",
		"course":  courseMap,
	})

	// Step 12: Course creation completed (logged by middleware)
}

// UpdateCourseHandler updates a specific field of a course.
// Uses a database transaction and executes a raw SQL UPDATE statement.
// Note: This function uses string interpolation for the column name, which could be a security risk
// if the column name is not properly validated.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request (must contain authenticated user context from AuthMiddleware)
//
// Request Body:
//   - id: Course ID to update (string, will be converted to int)
//   - column: Database column name to update (string)
//   - value: New value for the column (string)
//
// Response:
//   - 200 OK: Success (no body, logged)
//   - 400 Bad Request: If request body is invalid or course ID conversion fails
//   - 500 Internal Server Error: If database operations fail
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
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
		return
	}

	intID, err := strconv.Atoi(updateData.ID)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting course ID to int",
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
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting course from database",
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
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error updating course in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	tx.Commit()

	// Course update completed (logged by middleware)
}

// LinkRequestCourseHandler initiates a course sharing request by sending notifications
// to specified users. Generates or retrieves a link UUID for the course and sends
// course data via SSE notifications.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request (must contain authenticated user context from AuthMiddleware)
//
// Request Body:
//   - course_code: Code of the course to share (string, required)
//   - users_id: Array of user IDs to send the link request to ([]uint, required)
//
// Response:
//   - 200 OK: JSON object with "message", "course_id", "link_id", and "recipients"
//   - 400 Bad Request: If request body is invalid or course lookup fails
//
// Side Effects:
//   - Creates or updates course LinkID if not already set
//   - Sends SSE notifications to all specified recipients via gRPC
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
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
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
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error getting course by code",
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
			server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error saving link identifier",
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
		server.LogWarn(r.Context(), "Failed to marshal notification payload", err,
			"tags", []string{"notification", "network", "low"},
			"error_type", "internal",
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

	server.LogInfo(r.Context(), "Course link request processed", "course_id", c.ID, "recipients_count", len(linkRequestData.UsersID),
		"tags", []string{"course", "network", "medium"},
		"external_service", "grpc")
}

// AcceptLinkCourseHandler accepts a course link request and returns all assignments
// with their associated documents for synchronization. Notifies the original course
// owner that the link was accepted.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request (must contain authenticated user context from AuthMiddleware)
//
// Request Body:
//   - course.Course object containing course metadata (user_id, code, etc.)
//
// Response:
//   - 200 OK: JSON object with "assignments" array containing assignments with embedded documents
//   - 400 Bad Request: If request body is invalid or database queries fail
//
// Side Effects:
//   - Sends SSE notification to original course owner via gRPC
func AcceptLinkCourseHandler(w http.ResponseWriter, r *http.Request) {

	// Step 1: Extract context values set by middleware (start_time, request_id, user, db)
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Decode the course data from the request body
	// The course object contains metadata about the course being linked (from the original course owner)
	var c course.Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
		return
	}

	// Step 3: Retrieve all assignments for the course being linked
	// Query uses c.UserID (original course owner) and c.Code (course code) to find assignments
	// Ordered by creation date to maintain chronological order
	var courseAssignments []assignment.Assignment
	if err := db.Where("user_id = ? AND course_code = ?", c.UserID, c.Code).Order("created_at").Find(&courseAssignments).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error getting course assignments",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
		return
	}

	// Step 4: Enrich each assignment with its associated documents
	// This allows the client to receive complete assignment data including all document references
	var responseAssignments []assignment.Assignment
	for _, assignment := range courseAssignments {
		// Fetch documents for this assignment - needed for complete course sync
		assignmentDocuments, err := assignment.GetDocuments(db)

		if err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error getting assignment documents",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"COURSES", "DB"},
			)
			return
		}

		// Attach documents to assignment before adding to response
		assignment.Documents = assignmentDocuments
		responseAssignments = append(responseAssignments, assignment)
	}

	// Step 5: Return the enriched assignments to the client
	// The client will use this data to sync the course and assignments locally
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"assignments": responseAssignments,
	})

	// Step 6: Notify the original course owner that their course link was accepted
	// This provides real-time feedback via SSE that the course sharing was successful
	if GrpcClient != nil {
		// Send notification to the original course owner (c.UserID) that currentUser accepted the link
		// NotificationSync type indicates this is a synchronization event, not a new entity creation
		GrpcClient.SendNotification(context.Background(),
			&notifications.Notification{
				UserId:   uint32(c.UserID), // Original course owner receives the notification
				SenderId: uint32(userID),   // Current user (accepter) is the sender
				Entity:   string(models.EntityCourse),
				EntityId: uint32(c.ID),
				Type:     string(notif.NotificationSync),
				Title:    c.Name,
				Message:  fmt.Sprintf("%s is now linked to your course : %s", currentUser.Username, c.Code),
				Action:   "sync",
				Data:     "", // No additional data needed - notification is informational
			},
		)

	}

	// Step 7: Log the successful link acceptance for audit and debugging
	server.LogInfo(r.Context(), "Course link accepted", "course_code", c.Code, "from_user_id", c.UserID, "assignments_synced", len(responseAssignments),
		"tags", []string{"course", "db", "high"},
		"data_size", len(responseAssignments))
}
