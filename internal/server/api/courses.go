package server

import (
	"context"
	"encoding/json"
	Errors "errors"
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
	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID
	c.Locals("message", "Course created successfully")

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
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	// Step 3: Validate business-critical required fields
	if input.LocalID == "" || input.Code == "" || input.Semester == "" || input.Instructor == "" || input.StartDate == "" || input.EndDate == "" {
		err := Errors.New("missing required fields")
		return errors.WrapServer(
			err,
			errors.ReqParamMissing,
			"Missing required fields",
			fiber.StatusBadRequest,
		)
	}

	// Step 4: Parse and validate date formats (expects YYYY-MM-DD format)
	startDate, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Invalid start date format",
			fiber.StatusBadRequest,
		)
	}

	endDate, err := time.Parse(time.DateOnly, input.EndDate)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Invalid end date format",
			fiber.StatusBadRequest,
		)
	}

	// Step 5: Convert string numeric fields to proper integer types
	credits, err := strconv.Atoi(input.Credits)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Error formatting credits",
			fiber.StatusBadRequest,
		)
	}
	localID, err := strconv.Atoi(input.LocalID)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Error formatting local_id",
			fiber.StatusBadRequest,
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

	// Step 8: Retrieve the created course to get auto-generated fields (ID, timestamps)
	courseObj, err := course.Get_Course_byId(cVal.ID, tx)
	if err != nil {
		tx.Rollback()
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

	// Step 9: Convert course struct to map for consistent JSON response format
	courseMap := courseObj.ToMap()
	if courseMap == nil {
		tx.Rollback()
		err := Errors.New("failed to process course data")
		return errors.WrapServer(
			err,
			errors.ProcJSONMarshalFailed,
			"Error processing course data",
			fiber.StatusInternalServerError,
		)
	}

	// Step 10: Commit transaction after all operations succeed
	if err := tx.Commit().Error; err != nil {
		return errors.WrapServer(err, errors.DBTransactionFailed, "Error committing course creation transaction", fiber.StatusInternalServerError)
	}

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

	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	c.Locals("message", "Course updated successfully")

	// Extract course ID from path parameter
	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(
			fmt.Errorf("course ID required"),
			errors.ReqParamMissing,
			"Course ID required",
			fiber.StatusBadRequest,
		)
	}

	intID, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Error converting course ID to int",
			fiber.StatusBadRequest,
		)
	}

	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err = c.BodyParser(&updateData)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	courseObj, err := course.Get_Course_byId(uint(intID), db)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting course from database",
			fiber.StatusInternalServerError,
		)
	}

	if err := db.Exec(fmt.Sprintf("UPDATE courses SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), courseObj.ID).Error; err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error updating course in database",
			fiber.StatusInternalServerError,
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

	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, "message", "Course link request processed")

	var linkRequestData struct {
		CourseCode string `json:"course_code"`
		UsersID    []uint `json:"users_id"`
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
	courseObj, err := course.Get_Course_byCode(linkRequestData.CourseCode, userID, db)
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
					ctx,
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

	server.LogInfo(context.Background(), "Course link request processed", "course_id", courseObj.ID, "recipients_count", len(linkRequestData.UsersID),
		"tags", []string{"course", "grpc", "medium"},
	)

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
	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, "message", "Course link accepted")

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
		assignmentDocuments, err := assignment.GetDocuments(assign.ID, userID, db)

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
				Action:   "sync",
				Data:     "", // No additional data needed - notification is informational
			},
		)
		if err != nil {
			server.LogWarn(
				ctx,
				errors.WrapServer(
					err,
					errors.GRPCNotificationFailed,
					"Failed to send notification",
					fiber.StatusInternalServerError,
				),
			)
		}
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

	db.Transaction(func(tx *gorm.DB) error {
		courseObj, err := course.Get_Course_byId(courseID, tx)
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

		assignments, err := assignment.GetAssignmentsbyCourse(courseObj.Code, tx)
		if err != nil {
			return errors.WrapServer(
				err,
				errors.DBQueryFailed,
				"Error getting course assignments from database",
				fiber.StatusInternalServerError,
			)
		}

		for _, assign := range assignments {
			if err := assignment.DeleteAssignment(assign, tx); err != nil {
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
					"Course not found",
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
	})
	return c.JSON(fiber.Map{"message": "Course deleted successfully"})
}
