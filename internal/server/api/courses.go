package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
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
func GetCoursesHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values set by middleware (start_time, request_id, user, db)
	startTime := c.Locals("start_time").(time.Time)
	requestID := c.Locals("request_id").(string)
	db := c.Locals("db").(*gorm.DB)

	var userID uint
	if id := c.Query("id"); id != "" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid assignment ID",
				"tags", []string{"courses"},
			)
		}
		userID = uint(idInt)
	} else {
		currentUser := c.Locals("user").(user.User)
		userID = currentUser.ID
	}

	// Step 2: Query database for user's courses using parameterized query for security
	var courses []course.Course
	if err := db.Where("user_id = ?", userID).Find(&courses).Error; err != nil {
		// Handle database error with structured logging and proper HTTP status
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting courses from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	// Step 3: Transform course structs to maps for consistent JSON serialization
	var coursesMap []map[string]string
	for _, a := range courses {
		coursesMap = append(coursesMap, a.ToMap())
	}

	// Step 4: Send successful response with course data
	return c.JSON(fiber.Map{
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
func CreateCourseHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values and initialize transaction for atomicity
	startTime := c.Locals("start_time").(time.Time)
	requestID := c.Locals("request_id").(string)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
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
	if err := c.BodyParser(&input); err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
	}

	// Step 3: Validate business-critical required fields
	if input.LocalID == "" || input.Code == "" || input.Semester == "" || input.Instructor == "" || input.StartDate == "" || input.EndDate == "" {
		err := errors.New("missing required fields")
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Missing required fields",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "MISSING_REQUIRED_FIELDS"},
		)
	}

	// Step 4: Parse and validate date formats (expects YYYY-MM-DD format)
	startDate, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid start date format",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_START_DATE"},
		)
	}

	endDate, err := time.Parse(time.DateOnly, input.EndDate)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid end date format",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_END_DATE"},
		)
	}

	// Step 5: Convert string numeric fields to proper integer types
	credits, err := strconv.Atoi(input.Credits)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error formatting credits",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_CREDITS"},
		)
	}
	localID, err := strconv.Atoi(input.LocalID)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error formatting local_id",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_LOCAL_ID"},
		)
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
		return server.ResponseError(c, result.Error, fiber.StatusConflict, "Error creating course in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	// Step 8: Retrieve the created course to get auto-generated fields (ID, timestamps)
	courseObj, err := course.Get_Course_byId(cVal.ID, tx)
	if err != nil {
		tx.Rollback()
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting course from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	// Step 9: Convert course struct to map for consistent JSON response format
	courseMap := courseObj.ToMap()
	if courseMap == nil {
		tx.Rollback()
		err := errors.New("failed to process course data")
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error processing course data",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "MARSHALLING"},
		)
	}

	// Step 10: Commit transaction after all operations succeed
	tx.Commit()

	// Step 11: Send successful response with created course data
	return c.JSON(fiber.Map{
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
func UpdateCourseHandler(c *fiber.Ctx) error {
	startTime := c.Locals("start_time").(time.Time)
	requestID := c.Locals("request_id").(string)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	userID := currentUser.ID

	// Extract course ID from path parameter
	idStr := c.Params("id")
	if idStr == "" {
		return server.ResponseError(c, fmt.Errorf("course ID required"), fiber.StatusBadRequest, "Course ID required",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_COURSE_ID"},
		)
	}

	intID, err := strconv.Atoi(idStr)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error converting course ID to int",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_COURSE_ID"},
		)
	}

	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err = c.BodyParser(&updateData)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
	}

	courseObj, err := course.Get_Course_byId(uint(intID), db)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting course from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	if err := db.Exec(fmt.Sprintf("UPDATE courses SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), courseObj.ID).Error; err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error updating course in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	// Course update completed (logged by middleware)
	return c.JSON(fiber.Map{"message": "Course updated successfully"})
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
func LinkRequestCourseHandler(c *fiber.Ctx) error {
	startTime := c.Locals("start_time").(time.Time)
	requestID := c.Locals("request_id").(string)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	userID := currentUser.ID

	var linkRequestData struct {
		CourseCode string `json:"course_code"`
		UsersID    []uint `json:"users_id"`
	}

	if err := c.BodyParser(&linkRequestData); err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
	}

	// 1. Get send course informations
	courseObj, err := course.Get_Course_byCode(linkRequestData.CourseCode, userID, db)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error getting course by code",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	//2. Create an uuid for the link

	var linkId uuid.UUID
	if courseObj.LinkID == uuid.Nil {
		linkId = uuid.New()
		courseObj.LinkID = linkId
		if err = db.Save(&courseObj).Error; err != nil {
			return server.ResponseError(c, err, fiber.StatusBadRequest, "Error saving link identifier",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"COURSES", "DB"},
			)
		}
	} else {
		linkId = courseObj.LinkID
	}

	cJson, err := json.Marshal(courseObj)
	if err != nil {
		server.LogWarn(context.Background(), "Failed to marshal notification payload", err,
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
					EntityId: uint32(courseObj.ID),
					Type:     string(notif.NotificationSync),
					Title:    courseObj.Name,
					Message:  fmt.Sprintf("%s shared a course with you : %s", currentUser.Username, courseObj.Code),
					Action:   "sync",
					Data:     string(cJson),
				},
			)
		}
	}

	server.LogInfo(context.Background(), "Course link request processed", "course_id", courseObj.ID, "recipients_count", len(linkRequestData.UsersID),
		"tags", []string{"course", "network", "medium"},
		"external_service", "grpc")

	return c.JSON(fiber.Map{
		"message":    "Course link request processed",
		"course_id":  courseObj.ID,
		"link_id":    linkId,
		"recipients": linkRequestData.UsersID,
	})
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
func AcceptLinkCourseHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values set by middleware (start_time, request_id, user, db)
	startTime := c.Locals("start_time").(time.Time)
	requestID := c.Locals("request_id").(string)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Decode the course data from the request body
	// The course object contains metadata about the course being linked (from the original course owner)
	var courseObj course.Course
	if err := c.BodyParser(&courseObj); err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "REQUEST"},
		)
	}

	// Step 3: Retrieve all assignments for the course being linked
	// Query uses courseObj.UserID (original course owner) and courseObj.Code (course code) to find assignments
	// Ordered by creation date to maintain chronological order
	var courseAssignments []assignment.Assignment
	if err := db.Where("user_id = ? AND course_code = ?", courseObj.UserID, courseObj.Code).Order("created_at").Find(&courseAssignments).Error; err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error getting course assignments",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	// Step 4: Enrich each assignment with its associated documents
	// This allows the client to receive complete assignment data including all document references
	var responseAssignments []assignment.Assignment
	for _, assign := range courseAssignments {
		// Fetch documents for this assignment - needed for complete course sync
		assignmentDocuments, err := assignment.GetDocuments(assign.ID, userID, db)

		if err != nil {
			return server.ResponseError(c, err, fiber.StatusBadRequest, "Error getting assignment documents",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"COURSES", "DB"},
			)
		}

		// Attach documents to assignment before adding to response
		assign.Documents = assignmentDocuments
		responseAssignments = append(responseAssignments, assign)
	}

	// Step 6: Notify the original course owner that their course link was accepted
	// This provides real-time feedback via SSE that the course sharing was successful
	if GrpcClient != nil {
		// Send notification to the original course owner (courseObj.UserID) that currentUser accepted the link
		// NotificationSync type indicates this is a synchronization event, not a new entity creation
		GrpcClient.SendNotification(context.Background(),
			&notifications.Notification{
				UserId:   uint32(courseObj.UserID), // Original course owner receives the notification
				SenderId: uint32(userID),           // Current user (accepter) is the sender
				Entity:   string(models.EntityCourse),
				EntityId: uint32(courseObj.ID),
				Type:     string(notif.NotificationSync),
				Title:    courseObj.Name,
				Message:  fmt.Sprintf("%s is now linked to your course : %s", currentUser.Username, courseObj.Code),
				Action:   "sync",
				Data:     "", // No additional data needed - notification is informational
			},
		)
	}

	// Step 7: Log the successful link acceptance for audit and debugging
	server.LogInfo(context.Background(), "Course link accepted", "course_code", courseObj.Code, "from_user_id", courseObj.UserID, "assignments_synced", len(responseAssignments),
		"tags", []string{"course", "db", "high"},
		"data_size", len(responseAssignments))

	// Step 5: Return the enriched assignments to the client
	// The client will use this data to sync the course and assignments locally
	return c.JSON(fiber.Map{
		"assignments": responseAssignments,
	})
}

func DeleteCourseHandler(c *fiber.Ctx) error {
	startTime := c.Locals("start_time").(time.Time)
	requestID := c.Locals("request_id").(string)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	userID := currentUser.ID

	var courseID uint
	idStr := c.Params("id")
	if idStr == "" {
		return server.ResponseError(c, fmt.Errorf("course ID required"), fiber.StatusBadRequest, "Course ID required",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_COURSE_ID"},
		)
	}
	int_id, err := strconv.Atoi(idStr)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Error converting course ID to int",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "INVALID_COURSE_ID"},
		)
	}
	courseID = uint(int_id)

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	courseObj, err := course.Get_Course_byId(courseID, tx)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting course from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	assignments, err := assignment.GetAssignmentsbyCourse(courseObj.Code, tx)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting assignments from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	for _, assign := range assignments {
		if err := assignment.DeleteAssignment(assign, tx); err != nil {
			return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error deleting assignment from database",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"COURSES", "DB"},
			)
		}
	}

	if err := tx.Delete(&courseObj).Error; err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error deleting course from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"COURSES", "DB"},
		)
	}

	tx.Commit()
	return c.JSON(fiber.Map{"message": "Course deleted successfully"})
}
