package database

import (
	"context"
	"fmt"
	"path/filepath"

	"gorm.io/gorm"

	"unipilot/internal/errors"
	"unipilot/internal/models/aimessage"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"
	"unipilot/internal/models/note"
	"unipilot/internal/models/notifications"
	"unipilot/internal/models/user"
	"unipilot/internal/services/fileops"
	"unipilot/internal/services/utils"
)

// GetDB returns the database connection
func (h *Database) GetDB() *gorm.DB {
	return h.db
}

// GetUser retrieves a user by ID
func (h *Database) GetUser(id uint) (*user.User, error) {
	var u user.User
	err := h.db.First(&u, id).Error
	return &u, err
}

// GetAssignment retrieves an assignment by ID
func (h *Database) GetAssignment(id uint) (*assignment.LocalAssignment, error) {
	assignment, err := assignment.Get_Local_Assignment_byId(id, h.db)
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignment, nil
}

// GetAssignments retrieves all assignments for a user
func (h *Database) GetAssignments() ([]assignment.LocalAssignment, error) {
	var LocalAssignment []assignment.LocalAssignment
	err := h.db.Preload("Course").Preload("Type").Preload("Status").Order("deadline DESC").Order("created_at DESC").Find(&LocalAssignment).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return LocalAssignment, nil
}

// CreateAssignment creates a new assignment
func (h *Database) CreateAssignment(assignment *assignment.LocalAssignment) error {
	err := h.db.Create(assignment).Error
	if err != nil {
		return errors.HandleDBCreateError(err)
	}
	return nil
}

// UpdateAssignment updates an existing assignment
func (h *Database) UpdateAssignment(LocalAssignment *assignment.LocalAssignment, column, value string) error {
	// Only update the assignment fields, not the related course data
	err := h.db.Exec(fmt.Sprintf("UPDATE local_assignments SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalAssignment.ID)).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// DeleteAssignment deletes an assignment
func (h *Database) DeleteAssignment(assignment *assignment.LocalAssignment) error {
	err := h.db.Delete(assignment).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// GetCourse retrieves a course by ID
func (h *Database) GetCourse(id uint) (*course.Course, error) {
	if h == nil || h.db == nil {
		return nil, errors.NewAppError(errors.DBConnectionFailed, "Database connection failed", nil)
	}
	course, err := course.Get_Course_byId(id, h.db)
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return course, nil
}

// GetCourses retrieves all courses for a user
func (h *Database) GetCourses() ([]course.LocalCourse, error) {
	var LocalCourse []course.LocalCourse
	err := h.db.Order("start_date DESC").Find(&LocalCourse).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return LocalCourse, nil
}

// CreateCourse creates a new course
func (h *Database) CreateCourse(course *course.Course) error {

	err := h.db.Create(course).Error
	if err != nil {
		return errors.HandleDBCreateError(err)
	}
	return nil
}

// UpdateCourse updates an existing course
func (h *Database) UpdateCourse(LocalCourse *course.LocalCourse, column, value string) error {
	// Only update the assignment fields, not the related course data
	err := h.db.Exec(fmt.Sprintf("UPDATE local_courses SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalCourse.ID)).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// DeleteCourse deletes a course
func (h *Database) DeleteCourse(course *course.LocalCourse) error {
	err := h.db.Delete(course).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// GetNotes returns all notes for the current user
func (h *Database) GetNotes() ([]note.LocalNote, error) {
	var LocalNote []note.LocalNote
	err := h.db.Preload("Course").Find(&LocalNote).Order("created_at DESC").Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return LocalNote, nil
}

// CreateNote creates a new note
func (h *Database) CreateNote(note *note.LocalNote) error {
	err := h.db.Create(note).Error
	if err != nil {
		return errors.HandleDBCreateError(err)
	}
	return nil
}

// UpdateNote updates an existing note
func (h *Database) UpdateNote(LocalNote *note.LocalNote, column, value string) error {
	err := h.db.Exec(fmt.Sprintf("UPDATE local_notes SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalNote.ID)).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// DeleteNote deletes a note
func (h *Database) DeleteNote(note *note.LocalNote) error {
	err := h.db.Delete(note).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// GetNotifications returns all notifications for the current user
func (h *Database) GetNotifications() ([]notifications.LocalNotification, error) {
	var LocalNotification []notifications.LocalNotification
	err := h.db.
		Where("type != ?", notifications.NotificationAssignment).
		Where("type != ?", notifications.NotificationCourse).
		Find(&LocalNotification).
		Order("created_at DESC").Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return LocalNotification, nil
}

func (h *Database) DeleteNotification(notification *notifications.LocalNotification) error {
	err := h.db.Delete(notification).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

func (h *Database) CreateDocument(ctx context.Context, uploadReq fileops.FileUploadRequest, hasLocalFile bool) (*fileops.FileUploadResponse, error) {

	// Validate file type
	if err := fileops.ValidateFileType(uploadReq.FileName); err != nil {
		return nil, errors.NewAppError(errors.ValidationInvalid, "Unsupported file type", err)
	}

	// Validate file size
	if uploadReq.FileSize > document.MaxFileSize {

		return nil, errors.NewAppError(errors.ValidationInvalid, "File size exceeds limit", fmt.Errorf("file size exceeds limit of %d MB", document.MaxFileSize/(1024*1024)))
	}

	// Create LocalDocument record
	localDoc := document.LocalDocument{
		AssignmentID:       uploadReq.AssignmentID,
		RemoteAssignmentID: uploadReq.RemoteAssignmentID,
		UserID:             uploadReq.UserID,
		Type:               uploadReq.Type,
		FileName:           uploadReq.FileName,
		FileType:           fileops.GetMimeType(uploadReq.FileName),
		FileSize:           uploadReq.FileSize,
		StorageKey:         uploadReq.StorageKey,
		Version:            1,
		HasLocalFile:       hasLocalFile, // Will be set to true after successful file write
	}

	// Generate file path
	documentDir, err := utils.GetDocumentDir()
	if err != nil {
		return nil, errors.NewAppError(errors.ValidationInvalid, "Failed to get app data path", err)
	}

	// Create unique filename with assignment and user info
	fileName := fmt.Sprintf("doc_%d_%d_%s", uploadReq.AssignmentID, uploadReq.UserID, uploadReq.FileName)
	filePath := filepath.Join(documentDir, fileName)
	localDoc.FilePath = filePath

	//Check storage quota
	var totalSize int64
	h.db.Model(&document.LocalDocument{}).
		Where("user_id = ? AND has_local_file = ?", uploadReq.UserID, true).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize)

	if totalSize+uploadReq.FileSize > document.MaxUserQuota {
		return nil, errors.NewAppError(errors.ValidationInvalid, "Storage quota exceeded", fmt.Errorf("storage quota exceeded. Current: %d MB, Limit: %d MB",
			totalSize/(1024*1024), document.MaxUserQuota/(1024*1024)))
	}

	var response *fileops.FileUploadResponse

	if err := h.db.Create(&localDoc).Error; err != nil {
		return nil, errors.HandleDBCreateError(err)
	}

	if hasLocalFile {
		// Upload the document locally
		response, err = fileops.WriteDocument(&localDoc, uploadReq.FileContent, h.db)
		if err != nil {
			return nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to write document on disk")
		}
	} else {
		response = &fileops.FileUploadResponse{
			LocalDocument: &localDoc,
			Success:       true,
			Message:       "Upload successful",
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, errors.DBTransactionFailed, "Failed to create document")
	}

	return response, nil
}

// Saving a message from AI SDK
func (h *Database) SaveUIMessage(assignmentID uint, vercelMessage map[string]interface{}) error {
	message, err := aimessage.FromUIMessage(assignmentID, vercelMessage)
	if err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Failed to create message from UI message")
	}
	err = h.db.Create(message).Error
	if err != nil {
		return errors.HandleDBCreateError(err)
	}
	return nil
}

// Retrieving conversation history
func (h *Database) GetConversationHistory(assignmentID uint) ([]map[string]interface{}, error) {
	var dbMessages []aimessage.LocalAiMessage
	err := h.db.Where("assignment_id = ?", assignmentID).
		Order("created_at ASC").
		Find(&dbMessages).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}

	uiMessages := make([]map[string]interface{}, len(dbMessages))
	for i, msg := range dbMessages {
		uiMessages[i], err = msg.ToUIMessage()
		if err != nil {
			return nil, errors.Wrap(err, errors.ProcDataProcessingFailed, "Failed to convert message to UI message")
		}
	}

	return uiMessages, nil
}
