package server

import (
	"context"
	"encoding/json"
	Errors "errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	notif "unipilot/internal/models/notifications"
	"unipilot/internal/server/sse/grpc/notifications"

	//"unipilot/internal/models/document"

	"unipilot/internal/errors"
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

	db := c.Locals("db").(*gorm.DB)

	c.Locals("message", "Courses retrieved successfully")

	var userID uint
	if id := c.Query("id"); id != "" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Invalid course ID", fiber.StatusBadRequest)

		}
		userID = uint(idInt)
	} else {
		currentUser := c.Locals("user").(user.User)
		userID = currentUser.ID
	}

	// Step 2: Query database for user's courses using parameterized query for security
	var courses []course.Course
	if err := db.Where("user_id = ?", userID).Find(&courses).Error; err != nil {

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Courses not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting courses from database", fiber.StatusInternalServerError)
	}

	// Step 4: Send successful response with course data
	return c.JSON(courses)
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
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	c.Locals("message", "Course created successfully")

	var input course.Course
	// Parse JSON request body into input struct
	if err := c.BodyParser(&input); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	// Step 3: Validate business-critical required fields
	if input.Code == "" || input.Semester == "" || input.Instructor == "" {
		err := Errors.New("missing required fields")
		return errors.WrapServer(
			err,
			errors.ReqParamMissing,
			"Missing required fields",
			fiber.StatusBadRequest,
		)
	}

	// Add user id
	input.UserID = userID

	// Step 7: Persist course to database within transaction
	if result := db.Create(&input); result.Error != nil {
		if Errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return errors.WrapServer(
				result.Error,
				errors.DBConstraintViolation,
				"Course already exists",
				fiber.StatusConflict,
			)
		}
		return errors.WrapServer(
			result.Error,
			errors.DBQueryFailed,
			"Error creating course in database",
			fiber.StatusConflict,
		)
	}

	// Step 11: Send successful response with created course data
	return c.JSON(fiber.Map{
		"remote_id": input.ID,
	})
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
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	c.Locals("message", "Course updated successfully")

	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("course ID required"), errors.ReqParamMissing, "Course ID required", fiber.StatusBadRequest)
	}
	int_id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting course ID to int", fiber.StatusBadRequest)
	}
	courseID := uint(int_id)

	// Step 3: Define and parse assignment update request structure
	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err = c.BodyParser(&updateData)
	if err != nil {
		return errors.WrapServer(err, errors.ReqBodyInvalid, "Invalid request body", fiber.StatusBadRequest)
	}

	// Step 6: Execute raw SQL update with automatic timestamp tracking
	if err := db.Model(&course.Course{}).Where("id = ?", courseID).Update(updateData.Column, updateData.Value).Error; err != nil {

		if Errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.WrapServer(err, errors.DBConstraintViolation, "Course already exists", fiber.StatusConflict)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error updating course in database", fiber.StatusInternalServerError)
	}

	// Step 8: Course update completed (logged by middleware)
	return nil
}

func DeleteCourseHandler(c *fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}

	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	c.Locals("message", "Course deleted successfully")

	var courseID uint
	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("course ID required"), errors.ReqParamMissing, "Course ID required", fiber.StatusBadRequest)
	}
	int_id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Error converting course ID to int",
			fiber.StatusBadRequest,
		)
	}
	courseID = uint(int_id)

	courseObj, err := course.Get_Course_byId(courseID, db)
	if err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(
				err,
				errors.DBRecordNotFound,
				"Course not found",
				fiber.StatusNotFound,
			)
		}
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting course from database",
			fiber.StatusInternalServerError,
		)
	}

	assignments, err := assignment.GetAssignmentsbyCourse(courseObj.Code, userID, db)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting course assignments from database",
			fiber.StatusInternalServerError,
		)
	}

	if err := db.Transaction(func(tx *gorm.DB) error {

		for _, assign := range assignments {
			if err := assignment.DeleteAssignment(assign.ID, tx); err != nil {
				return errors.WrapServer(
					err,
					errors.DBQueryFailed,
					"Error deleting course assignment from database",
					fiber.StatusInternalServerError,
				)
			}
		}

		if err := tx.Delete(&courseObj).Error; err != nil {

			if Errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.WrapServer(
					err,
					errors.DBRecordNotFound,
					"Course not found",
					fiber.StatusNotFound,
				)
			}
			if Errors.Is(err, gorm.ErrForeignKeyViolated) {
				return errors.WrapServer(
					err,
					errors.DBConstraintViolation,
					"Course has dependencies",
					fiber.StatusConflict,
				)
			}
			return errors.WrapServer(
				err,
				errors.DBQueryFailed,
				"Error deleting course from database",
				fiber.StatusInternalServerError,
			)
		}
		return nil
	}); err != nil {
		return errors.WrapServer(
			err,
			errors.DBTransactionFailed,
			"Error transaction ; failed to delete course from database",
			fiber.StatusInternalServerError,
		)
	}
	return nil
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
	c.Locals("message", "Course link request processed")

	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID

	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("course ID required"), errors.ReqParamMissing, "Course ID required", fiber.StatusBadRequest)
	}
	int_id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting course ID to int", fiber.StatusBadRequest)
	}
	courseID := uint(int_id)

	var linkRequestData struct {
		UsersID []uint `json:"users_id"`
	}

	if err := c.BodyParser(&linkRequestData); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	// 1. Get send course informations
	courseObj, err := course.Get_Course_byId(courseID, db)
	if err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(
				err,
				errors.DBRecordNotFound,
				"Course not found",
				fiber.StatusNotFound,
			)
		}
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting course by code",
			fiber.StatusBadRequest,
		)
	}

	//2. Create an uuid for the link

	var linkId uuid.UUID
	if courseObj.LinkID == uuid.Nil {
		linkId = uuid.New()
		courseObj.LinkID = linkId
		if err = db.Save(&courseObj).Error; err != nil {
			return errors.WrapServer(
				err,
				errors.DBQueryFailed,
				"Error saving link identifier",
				fiber.StatusBadRequest,
			)
		}
	} else {
		linkId = courseObj.LinkID
	}

	cJson, err := json.Marshal(courseObj)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ProcJSONMarshalFailed,
			"Failed to marshal notification payload",
			fiber.StatusInternalServerError,
		)
	}

	if GrpcClient != nil {

		// 2. Send link info to users via SSE (field data)
		for _, sendeeID := range linkRequestData.UsersID {
			_, err := (*GrpcClient).SendNotification(context.Background(),
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
			if err != nil {
				server.LogWarn(
					context.Background(),
					errors.WrapServer(
						err,
						errors.GRPCNotificationFailed,
						"Failed to send notification",
						fiber.StatusInternalServerError,
					),
				)
			}
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
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
	c.Locals("message", "Course link accepted")
	// Step 1: Extract context values set by middleware (start_time, request_id, user, db)
	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID

	// Step 2: Decode the course data from the request body
	// The course object contains metadata about the course being linked (from the original course owner)
	var courseObj course.Course
	if err := c.BodyParser(&courseObj); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	// Step 3: Retrieve all assignments for the course being linked
	// Query uses courseObj.UserID (original course owner) and courseObj.Code (course code) to find assignments
	// Ordered by creation date to maintain chronological order
	var courseAssignments []assignment.Assignment
	if err := db.Where("user_id = ? AND course_code = ?", courseObj.UserID, courseObj.Code).Order("created_at").Find(&courseAssignments).Error; err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting course assignments",
			fiber.StatusBadRequest,
		)
	}

	// Step 4: Enrich each assignment with its associated documents
	// This allows the client to receive complete assignment data including all document references
	var responseAssignments []assignment.Assignment
	for _, assign := range courseAssignments {
		// Fetch documents for this assignment - needed for complete course sync
		assignmentDocuments, err := assignment.GetDocuments(assign.ID, db)

		if err != nil {
			return errors.WrapServer(
				err,
				errors.DBQueryFailed,
				"Error getting assignment documents",
				fiber.StatusBadRequest,
			)
		}

		// Attach documents to assignment before adding to response
		assign.Documents = assignmentDocuments
		responseAssignments = append(responseAssignments, assign)
	}

	rawData := map[string]string{
		"course_code": courseObj.Code,
		"link_id":     courseObj.LinkID.String(),
	}

	notificationData, err := json.Marshal(rawData)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ProcJSONMarshalFailed,
			"Failed to marshal notification payload",
			fiber.StatusInternalServerError,
		)
	}

	// Step 6: Notify the original course owner that their course link was accepted
	// This provides real-time feedback via SSE that the course sharing was successful
	if GrpcClient != nil {
		// Send notification to the original course owner (courseObj.UserID) that currentUser accepted the link
		// NotificationSync type indicates this is a synchronization event, not a new entity creation
		_, err := (*GrpcClient).SendNotification(context.Background(),
			&notifications.Notification{
				UserId:   uint32(courseObj.UserID), // Original course owner receives the notification
				SenderId: uint32(userID),           // Current user (accepter) is the sender
				Entity:   string(models.EntityCourse),
				EntityId: uint32(courseObj.ID),
				Type:     string(notif.NotificationSync),
				Title:    courseObj.Name,
				Message:  fmt.Sprintf("%s is now linked to your course : %s", currentUser.Username, courseObj.Code),
				Action:   string(notif.NotificationLink),
				Data:     string(notificationData), // No additional data needed - notification is informational
			},
		)
		if err != nil {
			return errors.WrapServer(
				err,
				errors.GRPCNotificationFailed,
				"Failed to send notification",
				fiber.StatusInternalServerError,
			)

		}
	}

	// Step 5: Return the enriched assignments to the client
	// The client will use this data to sync the course and assignments locally
	return c.JSON(fiber.Map{
		"assignments": responseAssignments,
	})
}
