package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unipilot/internal/app"
	"unipilot/internal/auth"
	"unipilot/internal/client"
	"unipilot/internal/events"
	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"
	"unipilot/internal/models/note"
	"unipilot/internal/models/notifications"
	"unipilot/internal/models/user"
	"unipilot/internal/network"
	"unipilot/internal/services/daemon"
	"unipilot/internal/services/fileops"
	"unipilot/internal/storage"
	"unipilot/internal/sync"

	"gorm.io/gorm"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx       context.Context
	Auth      *auth.Auth
	Events    *events.Events
	DB        *app.DatabaseHelper
	DaemonMgr *daemon.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		Auth:   auth.NewAuth(),
		Events: events.NewEvents(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize database helper
	dbHelper, err := app.NewDatabaseHelper()
	if err != nil {
		fmt.Printf("Warning: Could not initialize database helper: %v\n", err)
	} else {
		a.DB = dbHelper
	}

	// Initialize daemon manager
	if a.DB != nil {
		userID := a.DB.GetCurrentUserID()
		if userID > 0 {
			daemonMgr, err := daemon.NewManager(userID, ctx)
			if err != nil {
				log.Printf("Warning: Could not initialize daemon manager: %v", err)
			} else {
				a.DaemonMgr = daemonMgr

				// Auto-install daemon if not already installed
				if !daemonMgr.IsDaemonInstalled() {
					log.Printf("[App] Installing notification daemon...")
					if err := daemonMgr.InstallDaemon(); err != nil {
						log.Printf("Warning: Could not install notification daemon: %v", err)
						runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
							Type:    runtime.ErrorDialog,
							Title:   "Notification Setup",
							Message: "Failed to set up background notifications. Notifications will only work when the app is running.",
						})
					} else {
						log.Printf("[App] Notification daemon installed successfully")
					}
				} else {
					log.Printf("[App] Notification daemon already installed")
				}

				// Start daemon if not running
				if !daemonMgr.IsDaemonRunning() {
					log.Printf("[App] Starting notification daemon...")
					if err := daemonMgr.StartDaemon(); err != nil {
						log.Printf("Warning: Could not start notification daemon: %v", err)
					} else {
						log.Printf("[App] Notification daemon started successfully")
					}
				} else {
					log.Printf("[App] Notification daemon already running")
				}
			}
		}
	}

	// Check if user is already authenticated and initialize HTTP client if needed
	if _, err := a.Auth.IsAuthenticated(); err == nil {
		log.Println("[App] User already authenticated, initializing HTTP client...")
		if err := a.initializeAuthenticatedClient(); err != nil {
			log.Printf("[App] Failed to initialize authenticated client: %v", err)
		} else {
			if network.IsOnline() {
				// Start background sync manager
				syncManager := sync.NewSyncManager(a.DB.GetDB())
				go syncManager.BackgroundSync()

				// Process any pending syncs on startup
				go func() {
					if err := syncManager.ProcessPendingSyncs(); err != nil {
						log.Printf("[App] Startup sync error: %v", err)
					}
				}()
			}

		}
	}
}

// initializeAuthenticatedClient creates HTTP client from stored cookies when user is already authenticated
func (a *App) initializeAuthenticatedClient() error {
	httpClient, err := client.NewClientWithCookies()
	if err != nil {
		return fmt.Errorf("could not create http client from stored cookies: %w", err)
	}

	a.Auth.Client = httpClient
	log.Println("[App] HTTP client initialized from stored cookies")
	return nil
}

// ========================================
// CREATE OPERATIONS
// ========================================

// CreateAssignment creates a new assignment
func (a *App) CreateAssignment(assignmentData *assignment.LocalAssignment) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	tx := a.DB.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	localAssignment := &assignment.LocalAssignment{
		Title:      assignmentData.Title,
		Todo:       assignmentData.Todo,
		Deadline:   assignmentData.Deadline,
		CourseCode: assignmentData.CourseCode,
		TypeName:   assignmentData.TypeName,
		StatusName: assignmentData.StatusName,
		Priority:   assignmentData.Priority,
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return fmt.Errorf("user not authenticated")
	}

	// Create the assignment locally first
	if err := tx.Create(localAssignment).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Always try to sync with server
	remoteAssignment := &assignment.Assignment{
		LocalID:    localAssignment.ID,
		Title:      localAssignment.Title,
		Todo:       localAssignment.Todo,
		Deadline:   localAssignment.Deadline,
		CourseCode: localAssignment.CourseCode,
		TypeName:   localAssignment.TypeName,
		StatusName: localAssignment.StatusName,
		Priority:   localAssignment.Priority,
	}

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var responseAssignment map[string]interface{}
	var clientErr error
	if isOnline {
		responseAssignment, clientErr = client.CreateAssignment(remoteAssignment)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}

	if clientErr != nil {
		// Server operation failed, create sync log
		syncManager := sync.NewSyncManager(tx)
		if syncErr := syncManager.CreateSyncLog(
			models.EntityAssignment,
			localAssignment.ID,
			"create",
			"",
			"",
			clientErr,
		); syncErr != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create sync log: %w", syncErr)
		}

		// Commit the transaction with the sync log
		tx.Commit()

		return nil
	}

	// Server operation succeeded
	str_remote_id, ok := responseAssignment["id"].(string)
	if !ok {
		tx.Rollback()
		return fmt.Errorf("invalid remote assignment ID %v", responseAssignment["id"])
	}

	remote_id, err := strconv.Atoi(str_remote_id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("invalid remote assignment ID %v", responseAssignment["id"])
	}

	localAssignment.RemoteID = uint(remote_id)

	if err := tx.Save(localAssignment).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// CreateCourse creates a new course
func (a *App) CreateCourse(courseData *course.LocalCourse) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	tx := a.DB.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	localCourse := &course.LocalCourse{
		Name:            courseData.Name,
		Code:            courseData.Code,
		Color:           courseData.Color,
		Semester:        courseData.Semester,
		Schedule:        courseData.Schedule,
		Credits:         courseData.Credits,
		Location:        courseData.Location,
		Instructor:      courseData.Instructor,
		InstructorEmail: courseData.InstructorEmail,
		StartDate:       courseData.StartDate,
		EndDate:         courseData.EndDate,
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return fmt.Errorf("user not authenticated")
	}

	// Check if a soft-deleted course with the same code exists
	var existingCourse course.LocalCourse
	if err := tx.Unscoped().Where("code = ? AND deleted_at IS NOT NULL", localCourse.Code).First(&existingCourse).Error; err == nil {
		// A soft-deleted course with this code exists, permanently delete it first
		if err := tx.Unscoped().Delete(&existingCourse).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clean up soft-deleted course: %w", err)
		}
	}

	// Create the course within the transaction
	if err := tx.Create(localCourse).Error; err != nil {
		tx.Rollback()
		return err
	}

	remoteCourse := &course.Course{
		LocalID:         localCourse.ID,
		Name:            localCourse.Name,
		Code:            localCourse.Code,
		Color:           localCourse.Color,
		Semester:        localCourse.Semester,
		Schedule:        localCourse.Schedule,
		Credits:         localCourse.Credits,
		Location:        localCourse.Location,
		Instructor:      localCourse.Instructor,
		InstructorEmail: localCourse.InstructorEmail,
		StartDate:       localCourse.StartDate,
		EndDate:         localCourse.EndDate,
	}

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var responseCourse map[string]interface{}
	var clientErr error
	if isOnline {
		responseCourse, clientErr = client.CreateCourse(remoteCourse)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}

	if clientErr != nil {
		syncManager := sync.NewSyncManager(tx)
		if syncErr := syncManager.CreateSyncLog(
			models.EntityCourse,
			localCourse.ID,
			"create",
			"",
			"",
			clientErr,
		); syncErr != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create sync log: %w", syncErr)
		}
		tx.Commit()
		return nil
	}

	str_remote_id, ok := responseCourse["id"].(string)
	if !ok {
		tx.Rollback()
		return fmt.Errorf("invalid remote course ID %v", responseCourse["id"])
	}

	remote_id, err := strconv.Atoi(str_remote_id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("invalid remote course ID %v", responseCourse["id"])
	}

	log.Println("[App] Remote course ID:", remote_id)

	localCourse.RemoteID = uint(remote_id)

	if err := tx.Save(localCourse).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (a *App) CreateNote(noteData *note.LocalNote) error {

	if !network.IsOnline() {
		return fmt.Errorf("no network connection")
	}

	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	tx := a.DB.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	localNote := &note.LocalNote{
		Title:      noteData.Title,
		Subject:    noteData.Subject,
		CourseCode: noteData.CourseCode,
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return fmt.Errorf("user not authenticated")
	}

	// Create the note within the transaction
	if err := tx.Create(localNote).Error; err != nil {
		tx.Rollback()
		return err
	}

	remoteNote := &note.Note{
		LocalID:    localNote.ID,
		Title:      localNote.Title,
		Subject:    localNote.Subject,
		CourseCode: localNote.CourseCode,
	}

	responseNote, err := client.CreateNote(remoteNote)
	if err != nil {
		tx.Rollback()
		fmt.Println("Error creating remote note:", err)
		return err
	}
	localNote.RemoteID = responseNote["id"]
	localNote.Keywords = responseNote["keywords"]
	localNote.Content = responseNote["content"]

	if err := tx.Save(&localNote).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil

}

// UploadDocument opens a file dialog and uploads a document to an assignment
func (a *App) UploadDocument(assignmentID uint, documentType string) (*document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	// Validate document type
	if documentType != string(document.DocumentTypeSupport) && documentType != string(document.DocumentTypeSubmission) {
		return nil, fmt.Errorf("invalid document type: %s", documentType)
	}

	// Open file dialog
	filters := []runtime.FileFilter{
		{
			DisplayName: "Documents",
			Pattern:     "*.pdf;*.doc;*.docx;*.ppt;*.pptx;*.xls;*.xlsx;*.txt;*.md",
		},
		{
			DisplayName: "Images",
			Pattern:     "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.svg",
		},
		{
			DisplayName: "All Files",
			Pattern:     "*",
		},
	}

	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Select Document to Upload",
		Filters: filters,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open file dialog: %w", err)
	}

	if filePath == "" {
		return nil, fmt.Errorf("no file selected")
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get current user ID
	userID := a.DB.GetCurrentUserID()

	// Create upload request
	uploadReq := fileops.FileUploadRequest{
		AssignmentID: assignmentID,
		UserID:       userID,
		Type:         document.DocumentType(documentType),
		FileName:     filepath.Base(filePath),
		FileContent:  file,
		FileSize:     fileInfo.Size(),
	}

	// Upload the document locally
	response, err := fileops.UploadDocument(uploadReq, a.DB.GetDB())
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("upload failed: %s", response.Message)
	}

	// Also store metadata remotely for sharing
	if _, err := a.Auth.IsAuthenticated(); err == nil && a.Auth.Client != nil {
		metadataReq := map[string]interface{}{
			"assignment_id": assignmentID,
			"local_id":      response.LocalDocument.ID,
			"type":          documentType,
			"file_name":     filepath.Base(filePath),
			"file_type":     fileops.GetMimeType(filepath.Base(filePath)),
			"file_size":     fileInfo.Size(),
		}

		jsonData, _ := json.Marshal(metadataReq)
		resp, _ := a.Auth.Client.Post("https://newsroom.dedyn.io/acc-homework/document/metadata",
			"application/json", strings.NewReader(string(jsonData)))
		if resp.StatusCode == 200 {
			defer resp.Body.Close()
		}
		// We don't block on this - local file upload is the priority

	}

	return response.LocalDocument, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// UpdateAssignment updates an existing assignment
func (a *App) UpdateAssignment(LocalAssignment *assignment.LocalAssignment, column, value string) error {

	runtime.LogInfof(a.ctx, "[Backend] assignment %v, remote_id %v %v changed to %v", LocalAssignment.ID, LocalAssignment.RemoteID, column, value)

	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	db := a.DB.GetDB()

	if err := a.DB.UpdateAssignment(LocalAssignment, column, value); err != nil {
		return err
	}

	assignment_id_int := int(LocalAssignment.RemoteID)

	assignment_id := strconv.Itoa(assignment_id_int)

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var clientErr error
	if isOnline {
		clientErr = client.SendAssignmentUpdate(assignment_id, column, value)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}

	if clientErr != nil {

		runtime.LogErrorf(a.ctx, "[Backend] failed to send assignment update: %v", clientErr)

		sm := sync.NewSyncManager(db)
		syncLog, err := sm.GetSyncLog(models.EntityAssignment, LocalAssignment.ID, "update", column)

		// If no sync log is found, create a new one
		if err != nil {
			runtime.LogErrorf(a.ctx, "[Backend] failed to get sync log: %v", err)
			if syncErr := sm.CreateSyncLog(
				models.EntityAssignment,
				LocalAssignment.ID,
				"update",
				column,
				value,
				clientErr,
			); syncErr != nil {
				runtime.LogErrorf(a.ctx, "[Backend] failed to create sync log: %v", syncErr)
				return fmt.Errorf("failed to create sync log: %w", syncErr)
			}
			return nil
		}
		runtime.LogInfof(a.ctx, "[Backend] sync log: %v", syncLog)
		// If a sync log is found, update it
		syncLog.Value = value
		if err := db.Save(&syncLog).Error; err != nil {
			runtime.LogErrorf(a.ctx, "[Backend] failed to save sync log: %v", err)
			return fmt.Errorf("failed to save sync log: %w", err)
		}
		return nil
	}

	runtime.LogInfof(a.ctx, "[Backend] assignment update sent successfully")

	return nil
}

// UpdateCourse updates an existing course
func (a *App) UpdateCourse(course *course.LocalCourse, column, value string) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	db := a.DB.GetDB()

	if err := a.DB.UpdateCourse(course, column, value); err != nil {
		return err
	}

	course_id_int := int(course.RemoteID)

	course_id := strconv.Itoa(course_id_int)

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var clientErr error
	if isOnline {
		clientErr = client.SendCourseUpdate(course_id, column, value)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}

	if clientErr != nil {
		sm := sync.NewSyncManager(db)
		syncLog, err := sm.GetSyncLog(models.EntityCourse, course.ID, "update", column)
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityCourse,
				course.ID,
				"update",
				column,
				value,
				clientErr,
			); syncErr != nil {
				return fmt.Errorf("failed to create sync log: %w", syncErr)
			}
			return nil
		}
		syncLog.Value = value
		if err := db.Save(&syncLog).Error; err != nil {
			return fmt.Errorf("failed to save sync log: %w", err)
		}
		return nil
	}

	return nil
}

func (a *App) UpdateNote(LocalNote *note.LocalNote, column, value string) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	db := a.DB.GetDB()

	if err := a.DB.UpdateNote(LocalNote, column, value); err != nil {
		return err
	}

	note_id_int := int(LocalNote.ID)

	note_id := strconv.Itoa(note_id_int)

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var clientErr error
	if isOnline {
		clientErr = client.SendNoteUpdate(note_id, column, value)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}

	if clientErr != nil {
		sm := sync.NewSyncManager(db)
		syncLog, err := sm.GetSyncLog(models.EntityNote, LocalNote.ID, "update", column)
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityNote,
				LocalNote.ID,
				"update",
				column,
				value,
				clientErr,
			); syncErr != nil {
				return fmt.Errorf("failed to create sync log: %w", syncErr)
			}
			return nil
		}
		syncLog.Value = value
		if err := db.Save(&syncLog).Error; err != nil {
			return fmt.Errorf("failed to save sync log: %w", err)
		}
		return nil
	}

	return nil
}

// UploadNewDocumentVersion uploads a new version of an existing document
func (a *App) UploadNewDocumentVersion(existingDocumentID uint) (*document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	// Verify the existing document belongs to the user
	var existingDoc document.LocalDocument
	if err := a.DB.GetDB().Where("id = ? AND user_id = ?", existingDocumentID, userID).First(&existingDoc).Error; err != nil {
		return nil, fmt.Errorf("document not found or access denied")
	}

	// Open file dialog
	filters := []runtime.FileFilter{
		{
			DisplayName: "Documents",
			Pattern:     "*.pdf;*.doc;*.docx;*.ppt;*.pptx;*.xls;*.xlsx;*.txt;*.md",
		},
		{
			DisplayName: "Images",
			Pattern:     "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.svg",
		},
		{
			DisplayName: "All Files",
			Pattern:     "*",
		},
	}

	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Select New Version of Document",
		Filters: filters,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open file dialog: %w", err)
	}

	if filePath == "" {
		return nil, fmt.Errorf("no file selected")
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create new version request
	uploadReq := fileops.FileUploadRequest{
		AssignmentID: existingDoc.AssignmentID,
		UserID:       userID,
		Type:         existingDoc.Type,
		FileName:     filepath.Base(filePath),
		FileContent:  file,
		FileSize:     fileInfo.Size(),
	}

	// Upload new version locally
	response, err := fileops.UploadNewVersion(existingDocumentID, uploadReq, a.DB.GetDB())
	if err != nil {
		return nil, fmt.Errorf("version upload failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("version upload failed: %s", response.Message)
	}

	// Also update metadata remotely for sharing (async)
	if _, err := a.Auth.IsAuthenticated(); err == nil && a.Auth.Client != nil {
		metadataReq := map[string]interface{}{
			"assignment_id": existingDoc.AssignmentID,
			"local_id":      existingDoc.ID,
			"type":          string(existingDoc.Type),
			"file_name":     filepath.Base(filePath),
			"file_type":     fileops.GetMimeType(filepath.Base(filePath)),
			"file_size":     fileInfo.Size(),
			"version":       response.LocalDocument.Version,
		}

		go func() {
			jsonData, _ := json.Marshal(metadataReq)
			resp, err := a.Auth.Client.Post("https://newsroom.dedyn.io/acc-homework/documents/metadata",
				"application/json", strings.NewReader(string(jsonData)))
			if err == nil {
				defer resp.Body.Close()
			}
		}()
	}

	return response.LocalDocument, nil
}

func (a *App) UpdateUser(column, value string) (*user.User, error) {
	// Get the current user from storage
	u, err := storage.GetCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Update the specific field
	switch column {
	case "email":
		u.Email = value
	case "username":
		u.Username = value
	case "university":
		u.University = value
	case "semester":
		u.Semester = value
	case "year":
		u.Year = value
	case "language":
		u.Language = value
	case "avatar":
		u.Avatar = value
	default:
		return nil, fmt.Errorf("unknown column: %s", column)
	}

	// Update the timestamp
	u.UpdatedAt = time.Now()

	runtime.LogInfof(a.ctx, "Updated user: %v", u.ToMap())

	// Store the updated user in the credentials file
	if err := storage.StoreCredentials(*u); err != nil {
		return nil, fmt.Errorf("failed to store credentials: %w", err)
	}

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var clientErr error
	if isOnline {
		clientErr = client.SendUserUpdate(column, value)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}

	if clientErr != nil {
		db := a.DB.GetDB()
		sm := sync.NewSyncManager(db)
		syncLog, err := sm.GetSyncLog(models.EntityUser, u.ID, "update", column)
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityUser,
				u.ID,
				"update",
				column,
				value,
				clientErr,
			); syncErr != nil {
				return nil, fmt.Errorf("failed to create sync log: %w", syncErr)
			}
			return u, nil
		}
		syncLog.Value = value

		if err := db.Save(&syncLog).Error; err != nil {
			return nil, fmt.Errorf("failed to save sync log: %w", err)
		}
		return u, nil
	}

	return u, nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// DeleteAssignment deletes an assignment
func (a *App) DeleteAssignment(assignment *assignment.LocalAssignment) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	db := a.DB.GetDB()

	//Get all documents related to the assignment
	var documents []document.LocalDocument
	documents, err := a.GetAssignmentDocuments(assignment.ID)
	if err != nil {
		return err
	}

	// Delete all documents related to the assignment
	for _, document := range documents {
		if err := a.DeleteDocument(document.ID); err != nil {
			return err
		}
	}

	if err := a.DB.DeleteAssignment(assignment); err != nil {
		return err
	}

	assignment_id_str := strconv.Itoa(int(assignment.ID))

	deleted_at := time.Now().Format(time.RFC3339)

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var clientErr error
	if isOnline {
		clientErr = client.SendAssignmentUpdate(assignment_id_str, "deleted_at", deleted_at)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}

	if clientErr != nil {
		sm := sync.NewSyncManager(db)
		_, err := sm.GetSyncLog(models.EntityAssignment, assignment.ID, "create", "")
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityAssignment,
				assignment.ID,
				"delete",
				"deleted_at",
				deleted_at,
				clientErr,
			); syncErr != nil {
				return fmt.Errorf("failed to create sync log: %w", syncErr)
			}
			return nil
		}

		if err := sm.Undo(models.EntityAssignment, assignment.ID); err != nil {
			return fmt.Errorf("failed to delete sync log: %w", err)
		}

		return nil
	}

	return nil
}

// DeleteCourse deletes a course
func (a *App) DeleteCourse(course *course.LocalCourse) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	db := a.DB.GetDB()

	// Get all assignments related to the course
	assignments, err := a.GetCourseAssignments(course)
	if err != nil {
		return err
	}

	// Delete all assignments related to the course
	for _, assignment := range assignments {
		if err := a.DeleteAssignment(&assignment); err != nil {
			return err
		}
	}

	if err := a.DB.DeleteCourse(course); err != nil {
		return err
	}

	course_id_str := strconv.Itoa(int(course.ID))

	deleted_at := time.Now().Format(time.RFC3339)

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var clientErr error
	if isOnline {
		clientErr = client.SendCourseUpdate(course_id_str, "deleted_at", deleted_at)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}

	if clientErr != nil {

		sm := sync.NewSyncManager(db)
		_, err := sm.GetSyncLog(models.EntityCourse, course.ID, "create", "")
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityCourse,
				course.ID,
				"delete",
				"deleted_at",
				deleted_at,
				clientErr,
			); syncErr != nil {
				return fmt.Errorf("failed to create sync log: %w", syncErr)
			}
			return nil
		}

		if err := sm.Undo(models.EntityCourse, course.ID); err != nil {
			return fmt.Errorf("failed to save sync log: %w", err)
		}

		return nil
	}

	return nil
}

// DeleteDocument removes a document and its file
func (a *App) DeleteDocument(documentID uint) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	// Get local document record
	var doc document.LocalDocument
	if err := a.DB.GetDB().Where("id = ? AND user_id = ?", documentID, userID).First(&doc).Error; err != nil {
		return fmt.Errorf("document not found or access denied")
	}

	// Delete physical file if it exists
	if doc.HasLocalFile && doc.FilePath != "" {
		if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	db := a.DB.GetDB()
	// Delete database record
	if err := db.Delete(&doc).Error; err != nil {
		return fmt.Errorf("failed to delete document record: %w", err)
	}

	// Storage info is now calculated on-demand, no need to update cache

	// Also store metadata remotely for sharing
	if _, err := a.Auth.IsAuthenticated(); err == nil && a.Auth.Client != nil {

		resp, _ := a.Auth.Client.Post(fmt.Sprintf("https://newsroom.dedyn.io/acc-homework/document/metadata/delete?document_id=%d", documentID),
			"application/json", nil)
		if resp.StatusCode == 200 {
			defer resp.Body.Close()
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("failed to delete document metadata: %s", resp.Status)
		}

	}

	return nil
}

func (a *App) DeleteNote(note *note.LocalNote) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	db := a.DB.GetDB()

	if err := a.DB.DeleteNote(note); err != nil {
		return err
	}

	note_id_str := strconv.Itoa(int(note.ID))

	deleted_at := time.Now().Format(time.RFC3339)

	isOnline := network.IsOnline()
	runtime.LogInfof(a.ctx, "isOnline : %v", isOnline)
	var clientErr error
	if isOnline {
		clientErr = client.SendNoteUpdate(note_id_str, "deleted_at", deleted_at)
	} else {
		clientErr = fmt.Errorf("user is offline")
	}
	if clientErr != nil {
		sm := sync.NewSyncManager(db)
		_, err := sm.GetSyncLog(models.EntityNote, note.ID, "create", "")

		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityNote,
				note.ID,
				"delete",
				"deleted_at",
				deleted_at,
				clientErr,
			); syncErr != nil {
				return fmt.Errorf("failed to create sync log: %w", syncErr)
			}
			return nil
		}

		if err := sm.Undo(models.EntityNote, note.ID); err != nil {
			return fmt.Errorf("failed to save sync log: %w", err)
		}

		return nil
	}

	return nil
}

func (a *App) DeleteNotification(notification *notifications.LocalNotification) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.DeleteNotification(notification)
}

// ========================================
// OTHER OPERATIONS
// ========================================

// Register handles user registration
func (a *App) Register(username, email, password, university, language string) (*user.User, error) {
	user, err := a.Auth.Register(username, email, password, university, language)
	if err != nil {
		fmt.Println("Register error: ", err)
		return nil, err
	}

	// Reinitialize database helper after registration
	dbHelper, err := app.NewDatabaseHelper()
	if err != nil {
		fmt.Printf("Warning: Could not initialize database helper after registration: %v\n", err)
	} else {
		a.DB = dbHelper
	}

	return user, nil
}

// Login handles user authentication
func (a *App) Login(username, password string) (*user.User, error) {
	user, err := a.Auth.Login(username, password)
	if err != nil {
		return nil, err
	}

	// Reinitialize database helper after login
	dbHelper, err := app.NewDatabaseHelper()
	if err != nil {
		fmt.Printf("Warning: Could not initialize database helper after login: %v\n", err)
	} else {
		a.DB = dbHelper
	}

	return user, nil
}

// Logout handles user logout
func (a *App) Logout() error {
	if err := a.Auth.Logout(); err != nil {
		return err
	}

	a.DB = nil

	return nil
}

// IsAuthenticated checks if the user is currently authenticated
func (a *App) IsAuthenticated() (*user.User, error) {
	user, err := storage.GetCurrentUser()
	if err != nil {
		// When no credentials exist, return a LocalCredentials object with IsAuthenticated = false
		// instead of an error, so frontend can properly handle the unauthenticated state
		log.Printf("[App] No credentials found: %v", err)
		return nil, nil
	}

	return user, nil
}

// Sync performs synchronization of local changes with the remote server
func (a *App) Sync() error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Check if user is authenticated
	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return fmt.Errorf("user not authenticated")
	}

	// Check if we're online
	if !network.IsOnline() {
		return fmt.Errorf("not online, cannot sync")
	}

	log.Println("[App] Syncing local changes with the remote server")

	sm := sync.NewSyncManager(a.DB.GetDB())

	// Perform the sync
	return sm.ProcessPendingSyncs()
}

// GetAssignment returns an assignment by ID
func (a *App) GetAssignment(id uint) (*assignment.LocalAssignment, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetAssignment(id)
}

// GetCourse returns a course by ID
func (a *App) GetCourse(id uint) (*course.Course, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetCourse(id)
}

// GetUser returns a user by ID
func (a *App) GetUser(id uint) (*user.User, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetUser(id)
}

// GetAssignments returns all assignments for the current user
func (a *App) GetAssignments() ([]assignment.LocalAssignment, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetAssignments()
}

// GetCourses returns all courses for the current user
func (a *App) GetCourses() ([]course.LocalCourse, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetCourses()
}

// GetNotes returns all notes for the current user
func (a *App) GetNotes() ([]note.LocalNote, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetNotes()
}

// GetNotifications returns all notifications for the current user
func (a *App) GetNotifications() ([]notifications.LocalNotification, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetNotifications()
}

// Document Management Methods

// GetAssignmentDocuments retrieves all documents for an assignment
func (a *App) GetAssignmentDocuments(assignmentID uint) ([]document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	// Use LocalDocument and only return documents we have locally
	var documents []document.LocalDocument
	err := a.DB.GetDB().Where(
		"assignment_id = ? AND has_local_file = ?",
		assignmentID, true,
	).Order("created_at DESC").Find(&documents).Error

	return documents, err
}

// GetSupportDocuments retrieves only support documents for an assignment
func (a *App) GetSupportDocuments(assignmentID uint) ([]document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	var documents []document.LocalDocument
	err := a.DB.GetDB().Where(
		"assignment_id = ? AND type = ? AND has_local_file = ?",
		assignmentID, document.DocumentTypeSupport, true,
	).Order("created_at DESC").Find(&documents).Error

	return documents, err
}

// GetSubmissionDocuments retrieves only submission documents for an assignment
func (a *App) GetSubmissionDocuments(assignmentID uint) ([]document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	var documents []document.LocalDocument
	err := a.DB.GetDB().Where(
		"assignment_id = ? AND type = ? AND has_local_file = ?",
		assignmentID, document.DocumentTypeSubmission, true,
	).Order("created_at DESC").Find(&documents).Error

	return documents, err
}

// OpenDocument opens a document file with the system default application
func (a *App) OpenDocument(documentID uint) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return fmt.Errorf("user not authenticated")
	}

	// Get local document record
	var doc document.LocalDocument
	if err := a.DB.GetDB().Where("id = ?", documentID).First(&doc).Error; err != nil {
		return fmt.Errorf("document not found or access denied")
	}

	// Check if we have the file locally
	if !doc.HasLocalFile {
		return fmt.Errorf("file not available offline - please sync to download")
	}

	// Check if file actually exists on disk
	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		// Update database to reflect missing file
		a.DB.GetDB().Model(&doc).Update("has_local_file", false)
		return fmt.Errorf("file not found on disk")
	}

	// Open with system default application
	runtime.BrowserOpenURL(a.ctx, "file://"+doc.FilePath)
	return nil
}

// SaveDocumentAs opens a save dialog and copies the document to chosen location
func (a *App) SaveDocumentAs(documentID uint) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return fmt.Errorf("user not authenticated")
	}

	// Get local document record
	var doc document.LocalDocument
	if err := a.DB.GetDB().Where("id = ?", documentID).First(&doc).Error; err != nil {
		return fmt.Errorf("document not found or access denied")
	}

	// Check if we have the file locally
	if !doc.HasLocalFile {
		return fmt.Errorf("file not available offline - please sync to download")
	}

	// Check if file actually exists on disk
	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		// Update database to reflect missing file
		a.DB.GetDB().Model(&doc).Update("has_local_file", false)
		return fmt.Errorf("file not found on disk")
	}

	// Open save dialog
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           fmt.Sprintf("Save %s", doc.FileName),
		DefaultFilename: doc.FileName,
	})

	if err != nil {
		return fmt.Errorf("failed to open save dialog: %w", err)
	}

	if savePath == "" {
		return fmt.Errorf("no save location selected")
	}

	// Copy file
	sourceFile, err := os.Open(doc.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// GetUserStorageInfo returns storage statistics for the current user
func (a *App) GetUserStorageInfo() (*document.StorageInfo, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	// Calculate storage info on-demand
	storageInfo, err := document.GetUserStorageInfo(userID, a.DB.GetDB())
	if err != nil {
		return nil, fmt.Errorf("failed to get storage info: %w", err)
	}

	return storageInfo, nil
}

// GetRemoteDocumentMetadata retrieves document metadata from remote server (for shared assignments)
func (a *App) GetRemoteDocumentMetadata(assignmentID uint) ([]map[string]interface{}, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	if a.Auth.Client == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	// Make API call to get remote metadata
	url := fmt.Sprintf("https://newsroom.dedyn.io/acc-homework/documents?assignment_id=%d", assignmentID)
	resp, err := a.Auth.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Success   bool                     `json:"success"`
		Documents []map[string]interface{} `json:"documents"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("server request failed")
	}

	return result.Documents, nil
}

func (a *App) GetCourseAssignments(course *course.LocalCourse) ([]assignment.LocalAssignment, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if _, err := a.Auth.IsAuthenticated(); err != nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	var assignments []assignment.LocalAssignment
	err := a.DB.GetDB().Where("course_code = ?", course.Code).Find(&assignments).Order("created_at ASC").Error
	return assignments, err
}

// GetRemoteUsers returns all users from the remote server
func (a *App) GetRemoteUsers() ([]user.User, error) {
	return client.GetRemoteUsers()
}

// Follow a user
func (a *App) Follow(followedID uint) (bool, error) {
	return client.Follow(followedID)
}

// FollowResponse represents the response for follow-related queries
type FollowResponse struct {
	Users []user.User `json:"users"`
	Count int         `json:"count"`
}

// GetFollowers returns all followers for the current user
func (a *App) GetFollowers(userID uint) (*FollowResponse, error) {
	followers, count, err := client.GetFollowers(userID)
	if err != nil {
		return nil, err
	}
	return &FollowResponse{
		Users: followers,
		Count: count,
	}, nil
}

// GetFollowing returns all following for the current user
func (a *App) GetFollowing(userID uint) (*FollowResponse, error) {
	following, count, err := client.GetFollowing(userID)
	if err != nil {
		return nil, err
	}
	return &FollowResponse{
		Users: following,
		Count: count,
	}, nil
}

// GetNetworkStatus returns the current network connectivity status
func (a *App) GetNetworkStatus() map[string]interface{} {
	isOnline := network.IsOnline()
	return map[string]interface{}{
		"online":    isOnline,
		"timestamp": time.Now().Unix(),
	}
}

// Add daemon management methods for the UI:

// InstallNotificationDaemon installs the notification daemon
func (a *App) InstallNotificationDaemon() error {
	if a.DaemonMgr == nil {
		return fmt.Errorf("daemon manager not initialized")
	}
	return a.DaemonMgr.InstallDaemon()
}

// UninstallNotificationDaemon uninstalls the notification daemon
func (a *App) UninstallNotificationDaemon() error {
	if a.DaemonMgr == nil {
		return fmt.Errorf("daemon manager not initialized")
	}
	return a.DaemonMgr.UninstallDaemon()
}

// IsNotificationDaemonInstalled checks if the daemon is installed
func (a *App) IsNotificationDaemonInstalled() bool {
	if a.DaemonMgr == nil {
		return false
	}
	return a.DaemonMgr.IsDaemonInstalled()
}

// IsNotificationDaemonRunning checks if the daemon is running
func (a *App) IsNotificationDaemonRunning() bool {
	if a.DaemonMgr == nil {
		return false
	}
	return a.DaemonMgr.IsDaemonRunning()
}

// StartNotificationDaemon starts the daemon
func (a *App) StartNotificationDaemon() error {
	if a.DaemonMgr == nil {
		return fmt.Errorf("daemon manager not initialized")
	}
	return a.DaemonMgr.StartDaemon()
}

// StopNotificationDaemon stops the daemon
func (a *App) StopNotificationDaemon() error {
	if a.DaemonMgr == nil {
		return fmt.Errorf("daemon manager not initialized")
	}
	return a.DaemonMgr.StopDaemon()
}

// GetNotificationDaemonStatus returns the daemon status
func (a *App) GetNotificationDaemonStatus() map[string]interface{} {
	if a.DaemonMgr == nil {
		return map[string]interface{}{
			"installed": false,
			"running":   false,
			"error":     "Daemon manager not initialized",
		}
	}

	return map[string]interface{}{
		"installed": a.DaemonMgr.IsDaemonInstalled(),
		"running":   a.DaemonMgr.IsDaemonRunning(),
		"error":     nil,
	}
}

// Add method to rebuild daemon (for updates)
func (a *App) RebuildNotificationDaemon() error {
	if a.DaemonMgr == nil {
		return fmt.Errorf("daemon manager not initialized")
	}
	return a.DaemonMgr.RebuildDaemon()
}

// LinkCourse links a course to a list of users
func (a *App) RequestLinkCourse(courseCode string, usersID []uint) error {
	return client.RequestLinkCourse(courseCode, usersID)
}

func (a *App) AcceptLink(courseData string) error {

	// Unmarshal the course data
	var c course.LocalCourse
	if err := json.Unmarshal([]byte(courseData), &c); err != nil {
		return err
	}

	runtime.LogInfof(a.ctx, "AcceptLink: %v", c)

	//Determine if the course already exists
	var existingCourse course.LocalCourse
	err := a.DB.GetDB().Where("code = ?", c.Code).First(&existingCourse).Error
	if err != nil {
		// If the course doesn't exist, create it
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.CreateCourse(&c)
		}
		return err
	}
	// Update the course with the new link ID
	a.UpdateCourse(&existingCourse, "link_id", c.LinkID.String())

	return nil
}
