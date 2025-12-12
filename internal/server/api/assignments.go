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
func GetAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values from middleware (timing, user, database connection)
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Query user's assignments from database
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

	// Step 3: Convert assignments to safe map format for JSON response
	var assignmentsMap []map[string]string
	for _, a := range assignments {
		assignmentsMap = append(assignmentsMap, a.ToMap())
	}

	// Step 4: Send successful response with assignment data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
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
func CreateAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values from middleware (timing, user, database connection)
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Begin database transaction for atomic assignment creation
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"ASSIGNMENTS", "REQUEST"},
		)
		return
	}

	// Step 4: Validate all required fields for assignment creation
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

	// Step 5: Parse and validate deadline format (YYYY-MM-DD)
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

	// Step 6: Convert and validate local_id format (string to integer)
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

	// Step 7: Handle optional parent_id for sub-assignment relationships
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

	// Step 10: Retrieve complete assignment data with relationships for response
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

	// Step 11: Convert assignment to safe map format for JSON response
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

	// Step 12: Commit transaction after successful assignment creation
	tx.Commit()

	// Step 13: Send real-time notifications to all users linked to the course
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

	// Serialize assignment data for notification payload
	aJson, err := json.Marshal(newA)
	if err != nil {
		server.LogWarn(r.Context(), "Failed to marshal notification", err,
			"tags", []string{"notification", "network", "low"},
			"error_type", "internal",
		)
	}

	// Get all users linked to this course for notification distribution
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

	// Step 14: Send SSE notifications to linked users via gRPC (if available)
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

	// Step 15: Send successful response with created assignment data
	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Assignment created successfully",
		"assignment": assignmentMap,
	})

	// Step 16: Assignment creation completed (logged by middleware)
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
func UpdateAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values from middleware (timing, user, database connection)
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Begin database transaction for atomic assignment update
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Step 3: Define and parse assignment update request structure
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

	// Step 4: Convert assignment ID from string to integer for database query
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

	// Step 5: Validate assignment exists and user has ownership permissions
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

	// Step 6: Execute raw SQL update with automatic timestamp tracking
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

	// Step 7: Commit transaction after successful update
	tx.Commit()

	// Step 8: Assignment update completed (logged by middleware)
}
