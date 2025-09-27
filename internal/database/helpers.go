package database

import (
	"context"
	"fmt"
	"path/filepath"

	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"
	"unipilot/internal/models/note"
	"unipilot/internal/models/notifications"
	"unipilot/internal/models/user"
	"unipilot/internal/services/fileops"
	"unipilot/internal/services/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"gorm.io/gorm"
)

// GetAssignment retrieves an assignment by ID
func (h *Database) GetAssignment(id uint) (*assignment.LocalAssignment, error) {
	return assignment.Get_Local_Assignment_byId(id, h.db)
}

// GetAssignments retrieves all assignments for a user
func (h *Database) GetAssignments() ([]assignment.LocalAssignment, error) {
	var LocalAssignment []assignment.LocalAssignment
	err := h.db.Preload("Course").Preload("Type").Preload("Status").Order("deadline DESC").Order("created_at DESC").Find(&LocalAssignment).Error
	return LocalAssignment, err
}

// GetCourse retrieves a course by ID
func (h *Database) GetCourse(id uint) (*course.Course, error) {
	return course.Get_Course_byId(id, h.db)
}

// GetCourses retrieves all courses for a user
func (h *Database) GetCourses() ([]course.LocalCourse, error) {
	var LocalCourse []course.LocalCourse
	err := h.db.Find(&LocalCourse).Error
	return LocalCourse, err
}

// GetUser retrieves a user by ID
func (h *Database) GetUser(id uint) (*user.User, error) {
	var u user.User
	err := h.db.First(&u, id).Error
	return &u, err
}

// GetDB returns the database connection
func (h *Database) GetDB() *gorm.DB {
	return h.db
}

// CreateAssignment creates a new assignment
func (h *Database) CreateAssignment(assignment *assignment.LocalAssignment) error {
	return h.db.Create(assignment).Error
}

// UpdateAssignment updates an existing assignment
func (h *Database) UpdateAssignment(LocalAssignment *assignment.LocalAssignment, column, value string) error {
	// Only update the assignment fields, not the related course data
	return h.db.Exec(fmt.Sprintf("UPDATE local_assignments SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalAssignment.ID)).Error
}

// DeleteAssignment deletes an assignment
func (h *Database) DeleteAssignment(assignment *assignment.LocalAssignment) error {
	return h.db.Delete(assignment).Error
}

// CreateCourse creates a new course
func (h *Database) CreateCourse(course *course.Course) error {

	return h.db.Create(course).Error
}

// UpdateCourse updates an existing course
func (h *Database) UpdateCourse(LocalCourse *course.LocalCourse, column, value string) error {
	// Only update the assignment fields, not the related course data
	return h.db.Exec(fmt.Sprintf("UPDATE local_courses SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalCourse.ID)).Error
}

// DeleteCourse deletes a course
func (h *Database) DeleteCourse(course *course.LocalCourse) error {
	return h.db.Delete(course).Error
}

// GetNotes returns all notes for the current user
func (h *Database) GetNotes() ([]note.LocalNote, error) {
	var LocalNote []note.LocalNote
	err := h.db.Preload("Course").Find(&LocalNote).Order("created_at DESC").Error
	return LocalNote, err
}

// CreateNote creates a new note
func (h *Database) CreateNote(note *note.LocalNote) error {
	return h.db.Create(note).Error
}

// UpdateNote updates an existing note
func (h *Database) UpdateNote(LocalNote *note.LocalNote, column, value string) error {
	return h.db.Exec(fmt.Sprintf("UPDATE local_notes SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalNote.ID)).Error
}

// DeleteNote deletes a note
func (h *Database) DeleteNote(note *note.LocalNote) error {
	return h.db.Delete(note).Error
}

// GetNotifications returns all notifications for the current user
func (h *Database) GetNotifications() ([]notifications.LocalNotification, error) {
	h.db = h.db.Debug()
	var LocalNotification []notifications.LocalNotification
	err := h.db.
		Where("type = ?", notifications.NotificationFollow).
		Or("type = ?", notifications.NotificationSync).
		Or("type = ?", notifications.NotificationAssignment).
		Find(&LocalNotification).
		Order("created_at DESC").Error
	return LocalNotification, err
}

func (h *Database) DeleteNotification(notification *notifications.LocalNotification) error {
	return h.db.Delete(notification).Error
}

func (h *Database) CreateDocument(ctx context.Context, uploadReq fileops.FileUploadRequest, hasLocalFile bool) (*fileops.FileUploadResponse, error) {

	runtime.LogInfof(ctx, "DB CreateDocument: %v", uploadReq.FileName)

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	runtime.LogInfo(ctx, "STEP 1")

	// Validate file type
	if err := fileops.ValidateFileType(uploadReq.FileName); err != nil {
		return nil, fmt.Errorf("unsupported file type")
	}

	runtime.LogInfo(ctx, "STEP 2")

	// Validate file size
	if uploadReq.FileSize > document.MaxFileSize {

		return nil, fmt.Errorf("file size exceeds limit of %d MB", document.MaxFileSize/(1024*1024))
	}

	runtime.LogInfo(ctx, "STEP 3")

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

	runtime.LogInfo(ctx, "STEP 4")

	// Generate file path
	documentDir, err := utils.GetDocumentDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get app data path")
	}

	// Create unique filename with assignment and user info
	fileName := fmt.Sprintf("doc_%d_%d_%s", uploadReq.AssignmentID, uploadReq.UserID, uploadReq.FileName)
	filePath := filepath.Join(documentDir, fileName)
	localDoc.FilePath = filePath

	runtime.LogInfo(ctx, "STEP 5")

	//Check storage quota
	var totalSize int64
	tx.Model(&document.LocalDocument{}).
		Where("user_id = ? AND has_local_file = ?", uploadReq.UserID, true).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize)

	if totalSize+uploadReq.FileSize > document.MaxUserQuota {
		return nil, fmt.Errorf("storage quota exceeded. Current: %d MB, Limit: %d MB",
			totalSize/(1024*1024), document.MaxUserQuota/(1024*1024))
	}

	runtime.LogInfo(ctx, "STEP 6")

	// Save to database first
	if err := tx.Create(&localDoc).Error; err != nil {
		return nil, fmt.Errorf("failed to save document record")
	}

	runtime.LogInfo(ctx, "STEP 7")

	runtime.LogInfof(ctx, "Document Creation: %s", localDoc.FileName)

	var response *fileops.FileUploadResponse

	runtime.LogInfo(ctx, "STEP 8")

	if hasLocalFile {
		runtime.LogInfof(ctx, "Writing document to disk")
		// Upload the document locally
		response, err = fileops.WriteDocument(&localDoc, uploadReq.FileContent, tx)
		if err != nil {
			return nil, fmt.Errorf("upload failed: %w", err)
		}
	} else {
		runtime.LogInfof(ctx, "NO Writing document to disk")
		response = &fileops.FileUploadResponse{
			LocalDocument: &localDoc,
			Success:       true,
			Message:       "Upload successful",
		}
	}

	runtime.LogInfo(ctx, "STEP 9")

	if !response.Success {
		return nil, fmt.Errorf("upload failed: %s", response.Message)
	}

	runtime.LogInfof(ctx, "local upload response: %v", response)

	// Also store metadata remotely for sharing

	tx.Commit()

	return response, nil
}
