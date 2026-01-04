package server

import (
	"context"
	"encoding/json"
	Errors "errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/models"
	"unipilot/internal/server/sse/grpc/notifications"

	"unipilot/internal/errors"
	"unipilot/internal/server"

	"gorm.io/gorm"
)

// GetCourseHandler retrieves all courses belonging to the authenticated models.
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
		currentUser := c.Locals("user").(models.User)
		userID = currentUser.ID
	}

	// Step 2: Query database for user's courses using parameterized query for security
	var courses []models.Course
	if err := db.Where("user_id = ?", userID).Find(&courses).Error; err != nil {

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Courses not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting courses from database", fiber.StatusInternalServerError)
	}

	// Step 4: Send successful response with course data
	return c.JSON(courses)
}

// CreateCourseHandler creates a new course for the authenticated models.
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

	var input models.Course
	// Parse JSON request body into input struct
	if err := c.BodyParser(&input); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}
	// Add user id to the input struct
	input.UserID = userID

	// Step 3: Validate business-critical required fields
	if err := input.Validate(); err != nil {
		return errors.Inherit(err, errors.ValidationInvalid).ToServerError(fiber.StatusBadRequest)
	}

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

// UpdateCourseHandler updates a specific field of a models.
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
	if err := db.Model(&models.Course{}).Where("id = ?", courseID).Update(updateData.Column, updateData.Value).Error; err != nil {

		if Errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.WrapServer(err, errors.DBConstraintViolation, "Course already exists", fiber.StatusConflict)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error updating course in database", fiber.StatusInternalServerError)
	}

	// Step 8: Course update completed (logged by middleware)
	return nil
}

func DeleteCourseHandler(c *fiber.Ctx) error {

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

	if err := db.Set("qdrantClient", QdrantClient).Delete(&models.Course{}, courseID).Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error deleting course from database", fiber.StatusInternalServerError)
	}

	return nil
}

// LinkRequestCourseHandler initiates a course sharing request by sending notifications
// to specified users. Generates or retrieves a link UUID for the course and sends
// course data via SSE models.
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

	currentUser, ok := c.Locals("user").(models.User)
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
	courseObj, err := models.GetCourse(courseID, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting course by ID", fiber.StatusInternalServerError)
	}

	//2. Create an uuid for the link

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
					Type:     string(models.NotificationSync),
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
//   - models.Course object containing course metadata (user_id, code, etc.)
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
	currentUser, ok := c.Locals("user").(models.User)
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
	var courseObj models.Course
	if err := c.BodyParser(&courseObj); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	// Step 3: Retrieve all assignments for the course being linked
	var responseAssignments []models.Assignment
	responseAssignments, err := courseObj.GetAssignmentsByCourse(db.Preload("Documents").Order("created_at"))
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting course assignments", fiber.StatusInternalServerError)
	}

	// Step 6: Notify the original course owner that their course link was accepted
	if GrpcClient != nil {
		// Send notification to the original course owner (courseObj.UserID) that currentUser accepted the link
		_, err := (*GrpcClient).SendNotification(context.Background(),
			&notifications.Notification{
				UserId:   uint32(courseObj.UserID), // Original course owner receives the notification
				SenderId: uint32(userID),           // Current user (accepter) is the sender
				Entity:   string(models.EntityCourse),
				EntityId: uint32(courseObj.ID),
				Type:     string(models.NotificationSync),
				Title:    courseObj.Name,
				Message:  fmt.Sprintf("%s is now linked to your course : %s", currentUser.Username, courseObj.Code),
				Action:   string(models.NotificationLink),
				Data:     "", // No additional data needed - notification is informational
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
	return c.JSON(responseAssignments)
}

func GetCoursesLinkedHandler(c *fiber.Ctx) error {
	c.Locals("message", "Courses linked retrieved successfully")

	currentUser, ok := c.Locals("user").(models.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}

	ctx := context.Background()
	coursesLinked, err := CacheService.GetCoursesLinked(ctx, currentUser.ID)
	if err == nil && coursesLinked != nil {
		return c.JSON(coursesLinked)
	}

	// Get all linked courses with their root assignments and documents and notes

	coursesLinked, err = models.GetCoursesLinked(
		currentUser.ID,
		db.Preload("Children.Assignments", "parent_id = 0").
			Preload("Children.Assignments.User").Preload("Children.Assignments.Documents").
			Preload("Children.Notes", "parent_id = 0").Preload("Children.Notes.User"),
	)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting linked courses from database", fiber.StatusInternalServerError)
	}

	// Cache the result for future requests
	if cacheErr := CacheService.SetCoursesLinked(ctx, currentUser.ID, coursesLinked); cacheErr != nil {
		server.LogWarn(ctx, errors.WrapServer(cacheErr, errors.CacheOperationFailed, "Failed to cache linked courses", fiber.StatusInternalServerError))
	}

	if err := CacheService.SetExpirationCoursesLinked(ctx, currentUser.ID); err != nil {
		server.LogWarn(ctx, errors.WrapServer(err, errors.CacheOperationFailed, "Failed to set cache expiration for linked courses", fiber.StatusInternalServerError))
	}

	return c.JSON(coursesLinked)
}
