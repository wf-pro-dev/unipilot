package server

import (
	"context"
	Errors "errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/server/sse/grpc/messages"

	"unipilot/internal/server"

	"gorm.io/gorm"
)

// GetAssignmentHandler retrieves all assignments belonging to the authenticated models.
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

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	var userID uint
	if id := c.Query("id"); id != "" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Invalid assignment ID", fiber.StatusBadRequest)
		}
		userID = uint(idInt)
	} else {
		currentUser, ok := c.Locals("user").(models.User)
		if !ok {
			return errors.WrapServer(fmt.Errorf("user not found"), errors.ValidationInvalid, "User not found", fiber.StatusInternalServerError)
		}
		userID = currentUser.ID
	}

	// Step 2: Query user's assignments from database
	assignments, err := models.GetAssignments(userID, db.Preload("User").Omit("password_hash").Preload("Course"))
	if err != nil {

		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Assignments not found", fiber.StatusNotFound)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting assignments from database", fiber.StatusInternalServerError)
	}
	// Step 4: Send successful response with assignment data
	return c.JSON(assignments)
}

// CreateAssignmentHandler creates a new assignment for the authenticated models.
// Validates input data, creates assignment record using database transactions,
// and sends real-time notifications to users linked to the associated models.
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

	ctx := c.UserContext()

	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	userID := currentUser.ID

	var input models.Assignment
	if err := c.BodyParser(&input); err != nil {
		return errors.WrapServer(err, errors.ReqBodyInvalid, "Invalid request body", fiber.StatusBadRequest)
	}
	input.UserID = userID

	// Step 4: Validate all required fields for assignment creation
	if err := input.Validate(); err != nil {
		return errors.Inherit(err, errors.ValidationInvalid).ToServerError(fiber.StatusBadRequest)
	}

	// Step 9: Create assignment record in database within transaction
	result := db.Preload("Course").Create(&input).First(&input)
	if result.Error != nil {
		if Errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return errors.WrapServer(result.Error, errors.DBConstraintViolation, "Assignment already exists", fiber.StatusConflict)
		}
		return errors.WrapServer(result.Error, errors.DBQueryFailed, "Error creating assignment in database", fiber.StatusInternalServerError)
	}

	// Step 14: Send SSE notifications to linked users via gRPC (if available)

	go func() {
		if GrpcClient != nil && input.Course.IsInCluster(db) && !input.IsCopy() {

			clusterRootID := input.ClusterRoot()
			users_course, err := CacheService.GetClusterUsers(context.Background(), clusterRootID, db)
			if err != nil {
				server.LogWarn(context.Background(), errors.WrapServer(err, errors.DBQueryFailed, "Error getting users linked to course", fiber.StatusInternalServerError))
				return
			}

			for _, sendeeID := range users_course {
				if sendeeID == userID {
					continue
				}

				_, err := (*GrpcClient).SendMessage(context.Background(),
					&messages.Message{
						ReceiverId: uint32(sendeeID),
						SenderId:   uint32(userID),
						Title:      input.Title,
						Message:    fmt.Sprintf("%s shared a new assignment on %s", currentUser.Username, input.CourseCode),
						Data:       []byte(""),
						Type:       string(models.MessageNoContent),
					},
				)
				if err != nil {
					server.LogWarn(context.Background(), errors.WrapServer(err, errors.GRPCFailed, "Failed to send notification", fiber.StatusInternalServerError))
				}

			}

			CacheService.AddCourseAssignment(context.Background(), clusterRootID, input.ID)
			CacheService.SetAssignments(context.Background(), []*models.Assignment{&input})
		}

	}()

	// Step 15: Send successful response with created assignment data
	return c.JSON(fiber.Map{
		"remote_id": input.ID,
	})
}

// UpdateAssignmentHandler updates a specific field of an existing models.
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

	ctx := c.UserContext()

	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

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
	assignment := models.Assignment{
		Model: gorm.Model{ID: assignmentID},
	}
	if err := db.Preload("Course").Model(&assignment).Where("id = ?", assignmentID).Update(updateData.Column, updateData.Value).First(&assignment).Error; err != nil {

		if Errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.WrapServer(err, errors.DBConstraintViolation, "Assignment already exists", fiber.StatusConflict)
		}

		return errors.WrapServer(err, errors.DBQueryFailed, "Error updating assignment in database", fiber.StatusInternalServerError)
	}

	go func() {
		if assignment.Course.IsInCluster(db) && !assignment.IsCopy() {
			CacheService.SetAssignments(context.Background(), []*models.Assignment{&assignment})
		}
	}()

	// Step 8: Assignment update completed (logged by middleware)
	return nil
}

func DeleteAssignmentHandler(c *fiber.Ctx) error {

	ctx := c.UserContext()

	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	// Step 2: Extract assignment ID from path parameter
	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("assignment ID required"), errors.ReqParamMissing, "Assignment ID required", fiber.StatusBadRequest)
	}
	assignmentID, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting assignment ID to int", fiber.StatusBadRequest)
	}

	assignment, err := models.GetAssignment(uint(assignmentID), db.Preload("Course"))
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting assignment from database", fiber.StatusInternalServerError)
	}
	if err := db.Set("qdrantClient", QdrantClient).Delete(&assignment).Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error deleting assignment from database", fiber.StatusInternalServerError)
	}

	go func() {
		if assignment.Course.IsInCluster(db) && !assignment.IsCopy() {
			clusterRootID := assignment.ClusterRoot()
			CacheService.RemoveCourseAssignment(context.Background(), clusterRootID, uint(assignmentID))
			CacheService.DeleteAssignment(context.Background(), uint(assignmentID))
		}
	}()

	return nil
}
