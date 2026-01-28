package server

import (
	"context"
	"encoding/json"
	Errors "errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/models"
	"unipilot/internal/server"
	"unipilot/internal/server/sse/grpc/messages"

	"unipilot/internal/errors"

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

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	var userID uint
	if id := c.Query("id"); id != "" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Invalid course ID", fiber.StatusBadRequest)

		}
		userID = uint(idInt)
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

	if input.ParentID != 0 {
		go CacheService.AddClusterCourse(context.Background(), input.ParentID, input.ID)
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

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

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

	// If parent ID is changed, add course to cluster cache
	if updateData.Column == "parent_id" && updateData.Value != "0" {
		int_value, err := strconv.Atoi(updateData.Value)
		if err != nil {
			return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting parent ID to int", fiber.StatusBadRequest)
		}
		parentID := uint(int_value)
		go CacheService.AddClusterCourse(context.Background(), parentID, courseID)
	}

	// Step 8: Course update completed (logged by middleware)
	return nil
}

func DeleteCourseHandler(c *fiber.Ctx) error {

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

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
	go func() {

		course, err := models.GetCourse(courseID, db)
		if err != nil {
			server.LogWarn(context.Background(), errors.WrapServer(err, errors.DBQueryFailed, "Failed deletion course cache operation", fiber.StatusInternalServerError))
			return
		}

		// If course in cluster
		if course.ParentID != 0 {
			go CacheService.RemoveClusterCourse(context.Background(), course.ParentID, courseID)
			go CacheService.RemoveClusterUser(context.Background(), course.ParentID, course.UserID)
		}
	}()

	return nil
}

// ClusterShareCourseHandler shares a course with a cluster of users
func ClusterShareHandler(c *fiber.Ctx) error {

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}
	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

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

	course, err := models.GetCourse(courseID, db.Preload("Assignments", "parent_id is NULL").Preload("Notes", "parent_id is NULL"))
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting course", fiber.StatusInternalServerError)
	}

	for _, userID := range linkRequestData.UsersID {

		newInvitation := models.CourseInvitation{
			OwnerID:    currentUser.ID,
			ReceiverID: userID,
			SenderID:   currentUser.ID,
			CourseID:   courseID,
			CourseCode: course.Code,
			Status:     models.InvitationPending,
		}

		if err := newInvitation.Validate(); err != nil {
			return errors.WrapServer(err, errors.ValidationInvalid, "Invalid invitation", fiber.StatusBadRequest)
		}

		if err := db.Create(&newInvitation).Error; err != nil {
			return errors.WrapServer(err, errors.DBQueryFailed, "Error creating invitation", fiber.StatusInternalServerError)
		}

		if GrpcClient != nil {
			_, err = (*GrpcClient).SendMessage(context.Background(),
				&messages.Message{
					ReceiverId: uint32(userID),
					SenderId:   uint32(currentUser.ID),
					Title:      fmt.Sprintf("New Course Invitation: %s", course.Code),
					Message:    fmt.Sprintf("%s invited you to join their course", currentUser.Username),
					Data:       []byte(""),
					Type:       string(models.MessageNoContent),
				},
			)
			if err != nil {
				server.LogWarn(context.Background(), errors.WrapServer(err, errors.GRPCFailed, "Failed to send notification", fiber.StatusInternalServerError))
			}
		}
	}

	// Warm course cache async
	go CacheService.SetCourses(context.Background(), []*models.Course{course})

	return c.SendStatus(fiber.StatusNoContent)
}

func ClusterRequestHandler(c *fiber.Ctx) error {

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}
	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("course ID required"), errors.ReqParamMissing, "Course ID required", fiber.StatusBadRequest)
	}

	int_id, err := strconv.Atoi(idStr)

	clusterID := uint(int_id)

	cluster, err := models.GetCourse(clusterID, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting course", fiber.StatusInternalServerError)
	}

	newInvitation := models.CourseInvitation{
		OwnerID:    cluster.UserID,
		ReceiverID: currentUser.ID,
		SenderID:   currentUser.ID,
		CourseID:   clusterID,
		CourseCode: cluster.Code,
		Status:     models.InvitationPending,
	}

	if err := newInvitation.Validate(); err != nil {
		return errors.WrapServer(err, errors.ValidationInvalid, "Invalid invitation", fiber.StatusBadRequest)
	}

	if err := db.Create(&newInvitation).Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error creating invitation", fiber.StatusInternalServerError)
	}

	// Warm cluster cache async
	go CacheService.SetCourses(context.Background(), []*models.Course{cluster})

	return c.SendStatus(fiber.StatusNoContent)
}

// AcceptInvitationHandler accepts a course invitation
func AcceptInvitationHandler(c *fiber.Ctx) error {

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}
	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("invitation ID required"), errors.ReqParamMissing, "Invitation ID required", fiber.StatusBadRequest)
	}
	int_id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting invitation ID to int", fiber.StatusBadRequest)
	}
	invitationID := uint(int_id)

	// Get invitation from database
	invitation, err := models.GetCourseInvitation(invitationID, db.Preload("Course"))
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting invitation", fiber.StatusInternalServerError)
	}

	// Set invitation status to accepted
	invitation.Status = models.InvitationAccepted
	if err := db.Model(&models.CourseInvitation{}).Where("id = ?", invitationID).Update("status", models.InvitationAccepted).Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error updating invitation", fiber.StatusInternalServerError)
	}

	// Warm course cache async
	go func() {
		// Add course to owner cluster
		CacheService.AddUserCluster(context.Background(), invitation.OwnerID, invitation.CourseID)
		CacheService.AddClusterUser(context.Background(), invitation.CourseID, invitation.OwnerID)
		// Add course to receiver cluster
		CacheService.AddUserCluster(context.Background(), invitation.ReceiverID, invitation.CourseID)
		CacheService.AddClusterUser(context.Background(), invitation.CourseID, invitation.ReceiverID)
		// Add course to cluster cache
		CacheService.AddClusterCourse(context.Background(), invitation.CourseID, invitation.CourseID)
		// Get courses from cache
		CacheService.SetCourses(context.Background(), []*models.Course{invitation.Course})
	}()

	courseJSON, err := json.Marshal(invitation.Course)
	if err != nil {
		return errors.WrapServer(err, errors.ProcJSONMarshalFailed, "Failed to marshal course", fiber.StatusInternalServerError)
	}

	isShare := invitation.OwnerID == invitation.SenderID
	if GrpcClient != nil {
		if isShare {
			ownerMessage := fmt.Sprintf("%s is now linked to your course : %s", currentUser.Username, invitation.Course.Code)
			// Send message to the course owner
			_, err = (*GrpcClient).SendMessage(context.Background(),
				&messages.Message{
					ReceiverId: uint32(invitation.SenderID),   // Original course owner receives the notification
					SenderId:   uint32(invitation.ReceiverID), // Current user (accepter) is the sender
					Title:      invitation.Course.Name,
					Message:    ownerMessage,
					Data:       []byte(courseJSON),
					Type:       string(models.MessageCourse),
				},
			)
		} else {
			receiverMessage := fmt.Sprintf("%s has accepted your course request for %s", currentUser.Username, invitation.Course.Code)
			// Send message to the course owner
			_, err = (*GrpcClient).SendMessage(context.Background(),
				&messages.Message{
					ReceiverId: uint32(invitation.SenderID),   // Original course owner receives the notification
					SenderId:   uint32(invitation.ReceiverID), // Current user (accepter) is the sender
					Title:      invitation.Course.Name,
					Message:    receiverMessage,
					Data:       []byte(courseJSON),
					Type:       string(models.MessageCourse),
				},
			)
		}

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
	return c.SendStatus(fiber.StatusNoContent)
}

func DeclineInvitationHandler(c *fiber.Ctx) error {

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	idStr := c.Params("id")
	if idStr == "" {
		return errors.WrapServer(fmt.Errorf("invitation ID required"), errors.ReqParamMissing, "Invitation ID required", fiber.StatusBadRequest)
	}
	int_id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error converting invitation ID to int", fiber.StatusBadRequest)
	}
	invitationID := uint(int_id)

	invitation, err := models.GetCourseInvitation(invitationID, db.Preload("Course"))
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting invitation", fiber.StatusInternalServerError)
	}

	if err := db.Delete(&invitation).Error; err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error updating invitation", fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func GetCoursesLinkedHandler(c *fiber.Ctx) error {

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}
	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	var courseClusters []uint

	courseClusters, err = CacheService.GetUserClusterIDs(ctx, currentUser.ID, db)
	if err != nil {
		return errors.WrapServer(err, errors.CacheOperationFailed, "Error getting user courses linked from redis", fiber.StatusInternalServerError)
	}

	var courseIDs []uint
	for _, clusterID := range courseClusters {
		clusterCourseIDs, err := CacheService.GetClusterCourses(ctx, clusterID, db)
		if err != nil {
			return errors.WrapServer(err, errors.CacheOperationFailed, "Error getting cluster courses from redis", fiber.StatusInternalServerError)
		}
		courseIDs = append(courseIDs, clusterCourseIDs...)
	}

	// Get courses from cache
	courses, err := CacheService.GetCoursesByIDs(ctx, courseIDs, db)
	if err != nil {
		return errors.WrapServer(err, errors.CacheOperationFailed, "Error getting courses from redis", fiber.StatusInternalServerError)
	}

	return c.JSON(courses)
}
