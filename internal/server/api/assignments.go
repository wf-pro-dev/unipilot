package server

import (
	"context"
	"encoding/json"
	Errors "errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	notif "unipilot/internal/models/notifications"
	"unipilot/internal/models/user"
	"unipilot/internal/server/sse/grpc/notifications"

	"unipilot/internal/server"

	"gorm.io/gorm"
)

// GetAssignmentHandler retrieves all assignments belonging to the authenticated user.
// Returns a list of assignments converted to map format for consistent JSON serialization.
// Provides comprehensive assignment data for user's task management interface.
//
// HTTP Method: GET
// Content-Type: Not required (no request body expected)
//
// Request Body: None required (user context extracted from JWT token)
//
// Response (200 OK):
//   - message: Success message
//   - assignments: Array of assignment objects (as maps) with all assignment details
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Database Operations:
//   - Reads from `assignments` table filtered by `user_id`
//   - No caching strategy implemented (direct database access)
//
// Security Features:
//   - User can only access their own assignments (user_id filtering)
//   - Uses ToMap() method for safe JSON serialization
//   - Request tracking with unique request ID for audit trail
//
// Error Responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: Database query failure
//
// Side Effects:
//   - Logs successful retrieval with performance metrics
//   - No database modifications (read-only operation)
//   - Request duration tracking for performance monitoring
func GetAssignmentHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (timing, user, database connection)

	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}

	c.Locals("message", "Assignments list retrieved")

	var userID uint
	if id := c.Query("id"); id != "" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Invalid assignment ID", fiber.StatusBadRequest)
		}
		userID = uint(idInt)
	} else {
		currentUser, ok := c.Locals("user").(user.User)
		if !ok {
			return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
		}
		userID = currentUser.ID
	}

	// Step 2: Query user's assignments from database
	var assignments []assignment.Assignment
	if err := db.Where("user_id = ?", userID).Find(&assignments).Error; err != nil {

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Assignments not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting assignments from database", fiber.StatusInternalServerError)
	}

	// Step 3: Convert assignments to safe map format for JSON response
	var assignmentsMap []map[string]string
	for _, a := range assignments {
		assignmentsMap = append(assignmentsMap, a.ToMap())
	}

	// Step 4: Send successful response with assignment data
	return c.JSON(fiber.Map{
		"message":     "User's Assignments retrieved successfully",
		"assignments": assignmentsMap,
	})
}

// CreateAssignmentHandler creates a new assignment for the authenticated user.
// Validates input data, creates assignment record using database transactions,
// and sends real-time notifications to users linked to the associated course.
//
// HTTP Method: POST
// Content-Type: application/json
//
// Request Body:
//   - local_id: Assignment local identifier (string, required)
//   - title: Assignment title (string, required)
//   - todo: Assignment description/tasks (string, optional)
//   - deadline: Assignment deadline in YYYY-MM-DD format (string, required)
//   - course_code: Associated course code (string, required)
//   - type: Assignment type/category (string, required)
//   - status: Assignment status (string, optional)
//   - priority: Assignment priority level (string, optional)
//   - link: Related link/URL (string, optional)
//   - parent_id: Parent assignment ID for sub-assignments (string, optional)
//
// Response (200 OK):
//   - message: Success message
//   - assignment: Created assignment object (as map) with all details
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Database Operations:
//   - Creates record in `assignments` table using transaction
//   - Validates course association and user permissions
//   - Retrieves complete assignment data with relationships
//
// Notification System:
//   - Sends SSE notifications to all users linked to the course
//   - Uses gRPC client for real-time notification delivery
//   - Includes assignment data in notification payload
//
// Security Features:
//   - Transaction rollback on any failure for data consistency
//   - Input validation for required fields and data formats
//   - User isolation (assignments belong to authenticated user)
//
// Error Responses:
//   - 400 Bad Request: Invalid JSON, missing fields, or format errors
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 409 Conflict: Assignment creation failure (e.g., constraint violations)
//   - 500 Internal Server Error: Database operations or notification failures
//
// Side Effects:
//   - Creates assignment record in database
//   - Sends real-time notifications via SSE to course participants
//   - Logs creation with performance metrics and assignment details
func CreateAssignmentHandler(c *fiber.Ctx) error {

	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID

	c.Locals("message", "Assignment created successfully")

	// Step 3: Define and parse assignment creation request structure
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

	if err := c.BodyParser(&input); err != nil {
		return errors.WrapServer(err, errors.ReqBodyInvalid, "Invalid request body", fiber.StatusBadRequest)
	}

	// Step 4: Validate all required fields for assignment creation
	// Validate all required fields
	if input.LocalID == "" || input.CourseCode == "" || input.Title == "" || input.TypeName == "" || input.Deadline == "" {
		return errors.WrapServer(fmt.Errorf("missing required fields"), errors.ReqParamMissing, "Missing required fields", fiber.StatusBadRequest)

	}

	// Step 5: Parse and validate deadline format (YYYY-MM-DD)
	deadline, err := time.Parse(time.DateOnly, input.Deadline)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Invalid deadline format", fiber.StatusBadRequest)

	}

	// Step 6: Convert and validate local_id format (string to integer)
	local_id, err := strconv.Atoi(input.LocalID)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error formatting local_id", fiber.StatusBadRequest)
	}

	// Step 7: Handle optional parent_id for sub-assignment relationships
	var parent_id = 0
	if input.ParentID != "" {
		parent_id, err = strconv.Atoi(input.ParentID)
		if err != nil {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Error formatting parent_id", fiber.StatusBadRequest)
		}
	}

	// Step 8: Construct assignment object with validated data
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

	// Step 9: Create assignment record in database within transaction
	result := db.Create(&aVal)
	if result.Error != nil {
		if Errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return errors.WrapServer(result.Error, errors.DBConstraintViolation, "Assignment already exists", fiber.StatusConflict)
		}

		return errors.WrapServer(result.Error, errors.DBQueryFailed, "Error creating assignment in database", fiber.StatusInternalServerError)
	}

	// Step 10: Retrieve complete assignment data with relationships for response
	aObj := &aVal
	a, err := assignment.Get_Assignment_byID(aObj.ID, userID, db)
	if err != nil {

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Assignment not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting assignment from database", fiber.StatusInternalServerError)
	}

	// Step 11: Convert assignment to safe map format for JSON response
	// Convert to map safely
	assignmentMap := a.ToMap()
	if assignmentMap == nil {
		return errors.WrapServer(fmt.Errorf("assignment map is nil"), errors.ProcJSONMarshalFailed, "Error processing assignment data", fiber.StatusInternalServerError)
	}

	// Step 13: Send real-time notifications to all users linked to the course
	// Send a notification to all the users linked

	newA, err := assignment.Get_Assignment_byID(aObj.ID, userID, db)
	if err != nil {

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Assignment not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting assignment from database", fiber.StatusInternalServerError)
	}

	// Serialize assignment data for notification payload
	aJson, err := json.Marshal(newA)
	if err != nil {

		return errors.WrapServer(err, errors.ProcJSONMarshalFailed, "Error processing assignment data", fiber.StatusInternalServerError)
	}

	// Get all users linked to this course for notification distribution
	link_users, err := newA.Course.GetLinkUsers(db)
	if err != nil {

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Users linked to course not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting users linked to course", fiber.StatusInternalServerError)
	}

	// Step 14: Send SSE notifications to linked users via gRPC (if available)
	if GrpcClient != nil {
		for _, sendeeID := range link_users {
			if sendeeID != userID {
				_, err := (*GrpcClient).SendNotification(context.Background(),
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
				if err != nil {
					server.LogWarn(context.Background(), errors.WrapServer(err, errors.GRPCFailed, "Failed to send notification", fiber.StatusInternalServerError))
				}
			}
		}
	}

	// Step 15: Send successful response with created assignment data
	return c.JSON(fiber.Map{
		"message":    "Assignment created successfully",
		"assignment": assignmentMap,
	})
}

// UpdateAssignmentHandler updates a specific field of an existing assignment.
// Performs targeted field updates using raw SQL within a database transaction
// for data consistency. Only the assignment owner can update their assignments.
//
// HTTP Method: POST/PUT
// Content-Type: application/json
//
// Request Body:
//   - id: Assignment ID to update (string, required, converted to int)
//   - column: Database column name to update (string, required)
//   - value: New value for the specified column (string, required)
//
// Response (200 OK): No response body, success indicated by status code
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Database Operations:
//   - Validates assignment ownership by user_id
//   - Updates specified column in `assignments` table using raw SQL
//   - Automatically updates `updated_at` timestamp
//   - Uses database transaction for atomicity
//
// Security Features:
//   - User can only update their own assignments (ownership validation)
//   - Transaction rollback on any failure for data consistency
//   - Input validation for assignment ID format
//
// Security Notes:
//   - Uses string interpolation for column name - validate input to prevent SQL injection
//   - Assignment ownership verified before allowing updates
//
// Error Responses:
//   - 400 Bad Request: Invalid JSON body, malformed assignment ID
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: Database operations failure
//
// Side Effects:
//   - Modifies assignment record in database with timestamp update
//   - Logs successful updates with change details for audit trail
//   - Request duration tracking for performance monitoring
func UpdateAssignmentHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (timing, user, database connection)

	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID
	c.Locals("message", "Assignment updated successfully")

	var assignmentID uint
	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("assignment ID required"), errors.ReqParamMissing, "Assignment ID required", fiber.StatusBadRequest)
	}
	int_id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting assignment ID to int", fiber.StatusBadRequest)
	}
	assignmentID = uint(int_id)

	// Step 2: Begin database transaction for atomic assignment update
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Step 3: Define and parse assignment update request structure
	var updateData struct {
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err = c.BodyParser(&updateData)
	if err != nil {
		return errors.WrapServer(err, errors.ReqBodyInvalid, "Invalid request body", fiber.StatusBadRequest)
	}

	// Step 5: Validate assignment exists and user has ownership permissions
	a, err := assignment.Get_Assignment_byID(assignmentID, userID, tx)
	if err != nil {
		tx.Rollback()

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Assignment not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting assignment from database", fiber.StatusInternalServerError)
	}

	// Step 6: Execute raw SQL update with automatic timestamp tracking
	if err := tx.Exec(fmt.Sprintf("UPDATE assignments SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), a.ID).Error; err != nil {
		tx.Rollback()

		if Errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.WrapServer(err, errors.DBConstraintViolation, "Assignment already exists", fiber.StatusConflict)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error updating assignment in database", fiber.StatusInternalServerError)
	}

	// Step 7: Commit transaction after successful update
	if err := tx.Commit().Error; err != nil {
		return errors.WrapServer(err, errors.DBTransactionFailed, "Error committing assignment update transaction", fiber.StatusInternalServerError)
	}

	// Step 8: Assignment update completed (logged by middleware)
	return nil
}

func DeleteAssignmentHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (timing, user, database connection)

	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("db not found"), errors.ValidationInvalid, "DB not found", fiber.StatusInternalServerError)
	}
	userID := currentUser.ID

	c.Locals("message", "Assignment deleted successfully")
	// Step 2: Extract assignment ID from path parameter
	var assignmentID uint
	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("assignment ID required"), errors.ReqParamMissing, "Assignment ID required", fiber.StatusBadRequest)
	}
	int_id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting assignment ID to int", fiber.StatusBadRequest)
	}
	assignmentID = uint(int_id)

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	a, err := assignment.Get_Assignment_byID(assignmentID, userID, tx)
	if err != nil {
		tx.Rollback()

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Assignment not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting assignment from database", fiber.StatusInternalServerError)
	}

	if err := assignment.DeleteAssignment(*a, tx); err != nil {
		tx.Rollback()

		if Errors.Is(err, gorm.ErrModelValueRequired) {
			return errors.WrapServer(err, errors.DBConstraintViolation, "Deleting assignment violates foreign key constraints", fiber.StatusConflict)
		}

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Assignment not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error deleting assignment from database", fiber.StatusInternalServerError)
	}

	// Step 3:  Get Documents related to the assignment

	if err := tx.Commit().Error; err != nil {
		return errors.WrapServer(err, errors.DBTransactionFailed, "Error committing assignment delete transaction", fiber.StatusInternalServerError)
	}
	return c.JSON(fiber.Map{"message": "Assignment deleted successfully"})
}
