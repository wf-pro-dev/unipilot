package main

import (
	"context"
	"encoding/base64"
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

	"gorm.io/gorm"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"unipilot/internal/auth"
	"unipilot/internal/client"
	"unipilot/internal/database"
	Errors "unipilot/internal/errors"
	"unipilot/internal/events"
	"unipilot/internal/models"
	"unipilot/internal/network"
	"unipilot/internal/services/daemon"
	dbservice "unipilot/internal/services/database"
	"unipilot/internal/services/fileops"
	syncservice "unipilot/internal/services/sync"
	"unipilot/internal/services/utils"
)

// App struct
type App struct {
	ctx    context.Context
	Auth   *auth.Auth
	DB     *database.Database
	Events *events.Events
	Daemon *daemon.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Ensure client database is initialized if user is connected
	// This creates the database file if it doesn't exist (works offline)
	if err := dbservice.EnsureClientDBInitialized(); err != nil {
		log.Println(Errors.Wrap(err, Errors.DBConnectionFailed, "Failed to ensure client database is initialized").Error())
	}

	authService := auth.NewAuth()

	user, err := utils.GetUserFromFile()
	if err != nil {
		log.Println(Errors.Wrap(err, Errors.FSFileNotFound, "Failed to get user from file"))
	}
	authService.User = user
	var dbService *database.Database
	if authService.User != nil {
		var err error
		dbService, err = database.NewDatabase(authService.User)
		if err != nil {
			log.Fatal(Errors.Wrap(err, Errors.DBConnectionFailed, "Failed to initialize database").Error())
		}

	}

	var eventsService *events.Events
	if dbService != nil {
		eventsService = events.NewEvents(dbService.GetDB())
	}

	return &App{
		Auth:   authService,
		DB:     dbService,
		Events: eventsService,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize daemon manager only if user is authenticated
	// Service will be installed on login, not on startup
	if a.Auth.User != nil && a.Auth.User.ID > 0 {
		daemon, err := daemon.NewManager(a.Auth.User.ID, a.ctx)
		if err != nil {
			log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to initialize daemon manager").Error())
		} else {
			a.Daemon = daemon
			// Note: Service installation happens on Login(), not here
			// This ensures the service is user-specific and installed only when logged in
		}
	}

	// Check if user is already authenticated and initialize HTTP client if needed
	if a.Auth.IsAuthenticated() {
		// Ensure database is initialized and trigger server-to-client sync if needed
		if a.DB != nil {
			// Trigger server-to-client sync to download latest data from server
			go func() {
				if err := dbservice.EnsureClientDBAndSyncWithExisting(a.DB.GetDB()); err != nil {
					log.Println(Errors.Wrap(err, Errors.SyncFailed, "Failed to sync from server to client").Error())
				}
			}()
		}

		if network.IsOnline() {
			// Start background sync manager for client-to-server sync
			syncManager := syncservice.NewManager(a.DB.GetDB())
			syncManager.StartBackgroundSync()

			// Process any pending syncs on startup
			go func() {
				if err := syncManager.ProcessPendingSyncs(); err != nil {
					log.Println(Errors.Wrap(err, Errors.DBQueryFailed, "Failed to process pending syncs").Error())
				}
			}()
		}

	}
}

// ========================================
// CREATE OPERATIONS
// ========================================

// CreateAssignment creates a new assignment
func (a *App) CreateAssignment(assignmentData *models.LocalAssignment) (*models.LocalAssignment, error) {

	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	db := a.DB.GetDB()

	localAssignment := &models.LocalAssignment{
		BaseAssignment: models.BaseAssignment{
			Title:      assignmentData.Title,
			Todo:       assignmentData.Todo,
			Deadline:   assignmentData.Deadline,
			CourseID:   assignmentData.CourseID,
			CourseCode: assignmentData.CourseCode,
			Type:       assignmentData.Type,
			Status:     assignmentData.Status,
			Priority:   assignmentData.Priority,
			Link:       assignmentData.Link,
			ParentID:   assignmentData.ParentID,
		},
		RemoteCourseID: assignmentData.RemoteCourseID,
	}

	log.Println("localAssignment", localAssignment.Status)

	if err := localAssignment.Validate(); err != nil {
		return nil, err
	}

	// Create the assignment locally first
	if err := db.Create(localAssignment).Error; err != nil {
		return nil, Errors.HandleDBWriteError(err)
	}

	// Always try to sync with server

	remoteAssignment := localAssignment.ToRemote()

	//Online operation
	isOnline := network.IsOnline()
	var remoteID uint
	var clientErr error
	if isOnline {
		remoteID, clientErr = client.CreateAssignment(remoteAssignment)
	} else {

		clientErr = Errors.Wrap(fmt.Errorf("user is offline"), Errors.NetworkOffline, "User is offline")
	}

	if clientErr != nil {
		// Server operation failed, create sync log
		syncManager := syncservice.NewManager(db)
		if syncErr := syncManager.CreateSyncLog(
			models.EntityAssignment,
			localAssignment.ID,
			"create",
			"",
			"",
			clientErr,
		); syncErr != nil {
			return nil, Errors.Wrap(syncErr, Errors.ClientRequestFailed, "Failed to create sync log")
		}

		//Commit the transaction with the sync log

		return nil, nil
	}

	// Server operation succeeded
	localAssignment.RemoteID = remoteID

	if err := db.Save(localAssignment).Error; err != nil {
		return nil, Errors.HandleDBWriteError(err)
	}

	return localAssignment, nil
}

// CreateCourse creates a new course
func (a *App) CreateCourse(courseData *models.LocalCourse) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	db := a.DB.GetDB()

	localCourse := &models.LocalCourse{
		BaseCourse: models.BaseCourse{
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
		},
	}

	if err := localCourse.Validate(); err != nil {
		return err
	}

	// Create the course within the transaction
	if err := db.Create(localCourse).Error; err != nil {
		return Errors.HandleDBWriteError(err)
	}

	// Convert LocalCourse to Course (Base model) for server API
	remoteCourse := localCourse.ToRemote()

	isOnline := network.IsOnline()
	var remoteID uint
	var clientErr error
	if isOnline {
		remoteID, clientErr = client.CreateCourse(remoteCourse)
	} else {
		clientErr = Errors.Wrap(fmt.Errorf("user is offline"), Errors.NetworkOffline, "User is offline")
	}

	if clientErr != nil {
		syncManager := syncservice.NewManager(db)
		if syncErr := syncManager.CreateSyncLog(
			models.EntityCourse,
			localCourse.ID,
			"create",
			"",
			"",
			clientErr,
		); syncErr != nil {
			return Errors.Wrap(syncErr, Errors.ClientRequestFailed, "Failed to create sync log")
		}
		return nil
	}

	localCourse.RemoteID = remoteID

	if err := db.Save(localCourse).Error; err != nil {
		return Errors.HandleDBWriteError(err)
	}

	return nil
}

func (a *App) CreateNote(noteData *models.LocalNote) error {

	if !network.IsOnline() {
		return Errors.Wrap(fmt.Errorf("no network connection"), Errors.NetworkOffline, "No network connection")
	}

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	if noteData.Content == "" {
		return Errors.Wrap(fmt.Errorf("note data is missing"), Errors.ValidationRequired, "Note data is missing")
	}

	db := a.DB.GetDB()

	db.Transaction(func(tx *gorm.DB) error {

		// Create the note within the transaction
		if err := tx.Create(noteData).Error; err != nil {
			return Errors.HandleDBWriteError(err)
		}

		remoteNote := noteData.ToRemote()

		remoteID, err := client.CreateNote(remoteNote)
		if err != nil {
			tx.Rollback()
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to create note")
		}

		noteData.RemoteID = remoteID

		if err := tx.Save(&noteData).Error; err != nil {
			tx.Rollback()
			return Errors.HandleDBWriteError(err)
		}
		return nil
	})

	return nil

}

// UploadDocument opens a file dialog and uploads a document to an assignment
func (a *App) UploadDocument(assignmentID uint, remoteAssignmentID uint, documentType string) (*models.LocalDocument, error) {

	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	// Validate document type
	if documentType != string(models.DocumentTypeSupport) && documentType != string(models.DocumentTypeSubmission) {
		return nil, Errors.Wrap(fmt.Errorf("invalid document type: %s", documentType), Errors.ValidationInvalid, "Invalid document type")
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
		return nil, Errors.Wrap(err, Errors.FSOpenFailed, "Failed to open file dialog")
	}

	if filePath == "" {
		return nil, Errors.Wrap(fmt.Errorf("no file selected"), Errors.ValidationRequired, "No file selected")
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, Errors.Wrap(err, Errors.FSFileNotFound, "File not found")
	}

	fileContent, err := os.Open(filePath)
	if err != nil {
		return nil, Errors.Wrap(err, Errors.FSOpenFailed, "Failed to open file")
	}
	defer fileContent.Close()

	// Create upload request
	uploadReq := fileops.FileUploadRequest{
		AssignmentID:       assignmentID,
		RemoteAssignmentID: remoteAssignmentID,
		UserID:             a.Auth.User.ID,
		Type:               models.DocumentType(documentType),
		FileName:           filepath.Base(filePath),
		FilePath:           filePath,
		FileSize:           fileInfo.Size(),
		FileContent:        fileContent,
		StorageKey:         "",
	}

	document, err := a.CreateDocument(uploadReq, true)
	if err != nil {
		return nil, err
	}

	return document, nil

}

func (a *App) UploadProfilePicture() (string, error) {

	if a.DB == nil {
		return "", Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return "", Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	// Open file dialog
	filters := []runtime.FileFilter{

		{
			DisplayName: "Images",
			Pattern:     "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.svg",
		},
	}

	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Select Document to Upload",
		Filters: filters,
	})

	if err != nil {
		return "", Errors.Wrap(err, Errors.FSOpenFailed, "Failed to open file dialog")
	}

	if filePath == "" {
		return "", Errors.Wrap(fmt.Errorf("no file selected"), Errors.ValidationRequired, "No file selected")
	}

	fileContent, err := os.Open(filePath)
	if err != nil {
		return "", Errors.Wrap(err, Errors.FSOpenFailed, "Failed to open file")
	}

	// Copy the file to the user's profile picture directory
	profilePicturePath, err := utils.GetProfilePicturePath()
	if err != nil {
		return "", Errors.Wrap(err, Errors.FSFileNotFound, "Failed to get user directory")
	}

	if err := fileops.WriteFile(profilePicturePath, fileContent); err != nil {
		return "", Errors.Wrap(err, Errors.FSWriteFailed, "Failed to move file to profile picture directory")
	}

	a.Auth.User.Avatar = profilePicturePath
	if err := utils.SetCredentials(a.Auth.User); err != nil {
		return "", Errors.Wrap(err, Errors.FSWriteFailed, "Failed to set user to storage")
	}

	//Update the user's profile picture in the database
	clientErr := client.UpdateProfilePicture(profilePicturePath)
	if clientErr != nil {
		return "", Errors.Wrap(clientErr, Errors.ClientRequestFailed, "Failed to send profile picture")
	}

	return profilePicturePath, nil

}

// GetFileAsDataURL reads a local file and returns it as a data URL (base64 encoded)
// This is needed because webviews block file:// URLs for security reasons
func (a *App) GetFileAsDataURL(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path is empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	// Read file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Detect MIME type based on file extension
	mimeType := "image/png" // default
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".svg":
		mimeType = "image/svg+xml"
	case ".webp":
		mimeType = "image/webp"
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(fileData)

	// Return as data URL
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data), nil
}

func (a *App) CreateDocument(uploadReq fileops.FileUploadRequest, hasLocalFile bool) (*models.LocalDocument, error) {

	var response *models.LocalDocument
	var err error

	uploadResp, err := a.DB.CreateDocument(a.ctx, uploadReq, hasLocalFile)
	if err != nil {
		return nil, Errors.HandleDBWriteError(err)
	}

	response, err = a.SendDocument(uploadResp)
	if err != nil {
		return nil, Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to send document")
	}

	return response, err
}

func (a *App) SendDocument(uploadResp *fileops.FileUploadResponse) (*models.LocalDocument, error) {

	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if a.Auth.IsAuthenticated() {

		db := a.DB.GetDB()

		serverResponse, clientErr := client.SendDocument(uploadResp.LocalDocument)
		if clientErr != nil {
			return nil, Errors.Wrap(clientErr, Errors.ClientRequestFailed, "Failed to send document")
		}

		uploadResp.LocalDocument.StorageKey = serverResponse.StorageKey
		uploadResp.LocalDocument.RemoteID = serverResponse.RemoteID
		uploadResp.LocalDocument.RemoteAssignmentID = serverResponse.RemoteAssignmentID
		if err := db.Save(uploadResp.LocalDocument).Error; err != nil {
			return nil, Errors.HandleDBWriteError(err)
		}

	}

	return uploadResp.LocalDocument, nil

}

// DownloadDocument retrieves a document file for download
func (a *App) DownloadDocument(doc *models.LocalDocument) error {

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	// Send document to Server to create remote document & download from cloud
	downloadResp, err := client.DownloadDocument(doc)
	if err != nil {
		return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to download document")
	}

	// test if file is empty
	if downloadResp == nil {
		return Errors.Wrap(fmt.Errorf("file not found"), Errors.FSFileNotFound, "File not found")
	}

	// write file to disk
	if _, err := fileops.WriteDocument(doc, downloadResp, db); err != nil {
		return Errors.Wrap(err, Errors.FSWriteFailed, "Failed to write file")
	}

	return nil
}

// UploadDocumentRAG  uploads a document to the qdrant database for RAG
func (a *App) UploadDocumentRAG(doc *models.LocalDocument) error {

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if a.Auth.IsAuthenticated() {

		clientErr := client.UploadDocumentRAG(doc)
		if clientErr != nil {
			return Errors.Wrap(clientErr, Errors.ClientRequestFailed, "Failed to upload document to RAG")
		}

	}

	return nil

}

func (a *App) DeleteDocumentRAG(assignmentID, documentID uint) error {

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if err := client.DeleteDocumentRAG(assignmentID, documentID); err != nil {
		return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to delete document from RAG")
	}

	return nil
}

func (a *App) GetAssignmentDocumentIDsRAG(assignmentID uint) ([]uint, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	docIds, err := client.GetAssignmentDocumentIDsRAG(assignmentID)
	if err != nil {
		return nil, Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to get assignment document IDs")
	}

	return docIds, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// UpdateAssignment updates an existing assignment
func (a *App) UpdateAssignment(LocalAssignment *models.LocalAssignment, column, value string) error {

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	if err := db.
		Model(&models.LocalAssignment{}).
		Where("id = ?", LocalAssignment.ID).
		Update(column, value).Error; err != nil {
		return err
	}

	assignment_id_int := int(LocalAssignment.RemoteID)

	assignment_id := strconv.Itoa(assignment_id_int)

	isOnline := network.IsOnline()
	var clientErr error
	if isOnline {
		clientErr = client.UpdateAssignment(assignment_id, column, value)
	} else {
		clientErr = Errors.Wrap(fmt.Errorf("user is offline"), Errors.NetworkOffline, "User is offline")
	}

	if clientErr != nil {

		sm := syncservice.NewManager(db)
		syncLog, err := sm.GetSyncLog(models.EntityAssignment, LocalAssignment.ID, "update", column)

		// If no sync log is found, create a new one
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityAssignment,
				LocalAssignment.ID,
				"update",
				column,
				value,
				clientErr,
			); syncErr != nil {
				return Errors.Wrap(syncErr, Errors.DBQueryFailed, "Failed to create sync log")
			}
			return nil
		}
		// If a sync log is found, update it
		syncLog.Value = value
		if err := db.Save(&syncLog).Error; err != nil {
			return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to save sync log")
		}
		return nil
	}

	return nil
}

// UpdateCourse updates an existing course
func (a *App) UpdateCourse(course *models.LocalCourse, column, value string) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	if err := a.DB.UpdateCourse(course, column, value); err != nil {
		return err
	}

	course_id_int := int(course.RemoteID)

	course_id := strconv.Itoa(course_id_int)

	isOnline := network.IsOnline()
	var clientErr error
	if isOnline {
		clientErr = client.UpdateCourse(course_id, column, value)
	}

	if clientErr != nil && isOnline {
		sm := syncservice.NewManager(db)
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
				return Errors.Wrap(syncErr, Errors.DBQueryFailed, "Failed to create sync log")
			}
			return nil
		}
		syncLog.Value = value
		if err := db.Save(&syncLog).Error; err != nil {
			return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to save sync log")
		}
		return nil
	}

	return nil
}

func (a *App) UpdateNote(LocalNote *models.LocalNote, column, value string) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	if err := a.DB.UpdateNote(LocalNote, column, value); err != nil {
		return err
	}

	note_id_int := int(LocalNote.RemoteID)

	note_id := strconv.Itoa(note_id_int)

	isOnline := network.IsOnline()
	var clientErr error
	if isOnline {
		clientErr = client.UpdateNote(note_id, column, value)
	} else {
		clientErr = Errors.Wrap(fmt.Errorf("user is offline"), Errors.NetworkOffline, "User is offline")
	}

	if clientErr != nil {
		sm := syncservice.NewManager(db)
		syncLog, err := sm.GetSyncLog(models.EntityNote, LocalNote.RemoteID, "update", column)
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityNote,
				LocalNote.RemoteID,
				"update",
				column,
				value,
				clientErr,
			); syncErr != nil {
				return Errors.Wrap(syncErr, Errors.DBQueryFailed, "Failed to create sync log")
			}
			return nil
		}
		syncLog.Value = value
		if err := db.Save(&syncLog).Error; err != nil {
			return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to save sync log")
		}
		return nil
	}

	return nil
}

// UploadNewDocumentVersion uploads a new version of an existing document
func (a *App) UploadNewDocumentVersion(existingDocumentID uint) (*models.LocalDocument, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	userID := a.Auth.User.ID

	// Verify the existing document belongs to the user
	var existingDoc models.LocalDocument
	if err := a.DB.GetDB().Where("id = ? AND user_id = ?", existingDocumentID, userID).First(&existingDoc).Error; err != nil {
		return nil, Errors.Wrap(err, Errors.DBRecordNotFound, "Document not found or access denied")
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
		return nil, Errors.Wrap(err, Errors.FSOpenFailed, "Failed to open file dialog")
	}

	if filePath == "" {
		return nil, Errors.Wrap(fmt.Errorf("no file selected"), Errors.ValidationRequired, "No file selected")
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, Errors.Wrap(err, Errors.FSFileNotFound, "Failed to get file info")
	}

	// Open the file
	fileContent, err := os.Open(filePath)
	if err != nil {
		return nil, Errors.Wrap(err, Errors.FSOpenFailed, "Failed to open file")
	}
	defer fileContent.Close()

	// Create new version request
	uploadReq := fileops.FileUploadRequest{
		AssignmentID:       existingDoc.AssignmentID,
		RemoteAssignmentID: existingDoc.RemoteAssignmentID,
		UserID:             userID,
		Type:               existingDoc.Type,
		FileName:           filepath.Base(filePath),
		FilePath:           filePath,
		FileSize:           fileInfo.Size(),
	}

	// Upload new version locally
	response, err := fileops.UploadNewVersion(existingDocumentID, uploadReq, a.DB.GetDB())
	if err != nil {
		return nil, Errors.Wrap(err, Errors.StorageUploadFailed, "Version upload failed")
	}

	if !response.Success {
		return nil, Errors.Wrap(fmt.Errorf("%s", response.Message), Errors.StorageUploadFailed, "Version upload failed")
	}

	// // Also update metadata remotely for sharing (async)
	// if a.Auth.IsAuthenticated() {
	// 	metadataReq := map[string]interface{}{
	// 		"assignment_id": existingDoc.AssignmentID,
	// 		"local_id":      existingDoc.ID,
	// 		"type":          string(existingDoc.Type),
	// 		"file_name":     filepath.Base(filePath),
	// 		"file_type":     fileops.GetMimeType(filepath.Base(filePath)),
	// 		"file_size":     fileInfo.Size(),
	// 		"version":       response.LocalDocument.Version,
	// 	}

	// 	api_url := secrets.CONSTANTS["API_URL"]

	// 	go func() {
	// 		jsonData, _ := json.Marshal(metadataReq)
	// 		resp, err := http.Post(fmt.Sprintf("%s/documents/metadata", api_url),
	// 			"application/json", strings.NewReader(string(jsonData)))
	// 		if err == nil {
	// 			defer resp.Body.Close()
	// 		}
	// 	}()
	// }

	return response.LocalDocument, nil
}

func (a *App) UpdateUser(column, value string) (*models.User, error) {
	// Get the current user from storage
	currentUser := a.Auth.User

	if currentUser == nil {
		return nil, Errors.Wrap(fmt.Errorf("current user not found"), Errors.DBRecordNotFound, "Current user not found")
	}

	// Update the specific field
	switch column {
	case "email":
		currentUser.Email = value
	case "username":
		currentUser.Username = value
	case "university":
		currentUser.University = value
	case "semester":
		currentUser.Semester = value
	case "year":
		currentUser.Year = value
	case "language":
		currentUser.Language = value
	case "avatar":
		currentUser.Avatar = value
	default:
		return nil, Errors.Wrap(fmt.Errorf("unknown column: %s", column), Errors.ValidationInvalid, "Unknown column")
	}

	// Update the timestamp
	currentUser.UpdatedAt = time.Now()

	// Set the user to the auth struct
	a.Auth.User = currentUser

	// Set the user to the storage
	if err := utils.SetCredentials(currentUser); err != nil {
		return nil, Errors.Wrap(err, Errors.FSWriteFailed, "Failed to set user to storage")
	}

	isOnline := network.IsOnline()
	var clientErr error
	if isOnline {
		clientErr = client.UpdateUser(column, value)
	} else {
		clientErr = Errors.Wrap(fmt.Errorf("user is offline"), Errors.NetworkOffline, "User is offline")
	}

	if clientErr != nil {
		db := a.DB.GetDB()
		sm := syncservice.NewManager(db)
		syncLog, err := sm.GetSyncLog(models.EntityUser, currentUser.ID, "update", column)
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityUser,
				currentUser.ID,
				"update",
				column,
				value,
				clientErr,
			); syncErr != nil {
				return nil, Errors.Wrap(syncErr, Errors.DBQueryFailed, "Failed to create sync log")
			}
			return currentUser, nil
		}
		syncLog.Value = value

		if err := db.Save(&syncLog).Error; err != nil {
			return nil, Errors.Wrap(err, Errors.DBQueryFailed, "Failed to save sync log")
		}
		return currentUser, nil
	}

	return currentUser, nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// DeleteAssignment deletes an assignment
func (a *App) DeleteAssignment(assignment *models.LocalAssignment) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	if err := db.Delete(&assignment).Error; err != nil {
		return Errors.HandleDBWriteError(err)
	}

	remote_assignment_id_str := strconv.Itoa(int(assignment.RemoteID))

	isOnline := network.IsOnline()
	var clientErr error
	if isOnline {
		clientErr = client.DeleteAssignment(remote_assignment_id_str)
	} else {
		clientErr = Errors.Wrap(fmt.Errorf("user is offline"), Errors.NetworkOffline, "User is offline")
	}

	if clientErr != nil {
		sm := syncservice.NewManager(db)
		_, err := sm.GetSyncLog(models.EntityAssignment, assignment.ID, "create", "")
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityAssignment,
				assignment.ID,
				"delete",
				"deleted_at",
				time.Now().Format(time.RFC3339),
				clientErr,
			); syncErr != nil {
				return Errors.Wrap(syncErr, Errors.DBQueryFailed, "Failed to create sync log")
			}
			return nil
		}

		if err := sm.Undo(models.EntityAssignment, assignment.ID); err != nil {
			return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to delete sync log")
		}

		return nil
	}

	return nil
}

// DeleteCourse deletes a course
func (a *App) DeleteCourse(course *models.LocalCourse) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	if err := db.Delete(&course).Error; err != nil {
		return Errors.HandleDBWriteError(err)
	}

	course_id_str := strconv.Itoa(int(course.RemoteID))

	isOnline := network.IsOnline()
	var clientErr error
	if isOnline {
		clientErr = client.DeleteCourse(course_id_str)
	} else {
		clientErr = Errors.Wrap(fmt.Errorf("user is offline"), Errors.NetworkOffline, "User is offline")
	}

	if clientErr != nil {

		sm := syncservice.NewManager(db)
		_, err := sm.GetSyncLog(models.EntityCourse, course.ID, "create", "")
		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityCourse,
				course.ID,
				"delete",
				"deleted_at",
				time.Now().Format(time.RFC3339),
				clientErr,
			); syncErr != nil {
				return Errors.Wrap(syncErr, Errors.DBQueryFailed, "Failed to create sync log")
			}
			return nil
		}

		if err := sm.Undo(models.EntityCourse, course.ID); err != nil {
			return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to save sync log")
		}

		return nil
	}

	return nil
}

// DeleteDocument removes a document and its file
func (a *App) DeleteDocument(documentID uint) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	// Get local document record
	var doc models.LocalDocument
	if err := a.DB.GetDB().Where("id = ?", documentID).First(&doc).Error; err != nil {
		return Errors.Wrap(err, Errors.DBRecordNotFound, "Document not found or access denied")
	}

	// Delete physical file if it exists
	if doc.HasLocalFile && doc.FilePath != "" {
		if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
			return Errors.Wrap(err, Errors.FSDeleteFailed, "Failed to delete file")
		}
	}

	db := a.DB.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Delete(&doc).Error; err != nil {
			return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to delete document record")
		}

		// Also store metadata remotely for sharing
		if !a.Auth.IsAuthenticated() {
			return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
		}

		if err := client.DeleteDocument(doc.RemoteID); err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to delete document metadata")
		}

		return nil
	})
	if err != nil {
		return Errors.Wrap(err, Errors.DBTransactionFailed, "Failed to delete document transaction")
	}
	// Delete database record

	return nil
}

func (a *App) DeleteNote(note *models.LocalNote) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	if err := a.DB.DeleteNote(note); err != nil {
		return err
	}

	note_id_str := strconv.Itoa(int(note.RemoteID))

	deleted_at := time.Now().Format(time.RFC3339)

	isOnline := network.IsOnline()
	var clientErr error
	if isOnline {
		clientErr = client.UpdateNote(note_id_str, "deleted_at", deleted_at)
	} else {
		clientErr = Errors.Wrap(fmt.Errorf("user is offline"), Errors.NetworkOffline, "User is offline")
	}
	if clientErr != nil {
		sm := syncservice.NewManager(db)
		_, err := sm.GetSyncLog(models.EntityNote, note.RemoteID, "create", "")

		if err != nil {
			if syncErr := sm.CreateSyncLog(
				models.EntityNote,
				note.RemoteID,
				"delete",
				"deleted_at",
				deleted_at,
				clientErr,
			); syncErr != nil {
				return Errors.Wrap(syncErr, Errors.DBQueryFailed, "Failed to create sync log")
			}
			return nil
		}

		if err := sm.Undo(models.EntityNote, note.RemoteID); err != nil {
			return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to save sync log")
		}

		return nil
	}

	return nil
}

func (a *App) DeleteNotification(notification *models.LocalNotification) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return a.DB.DeleteNotification(notification)
}

// ========================================
// OTHER OPERATIONS
// ========================================

// Register handles user registration
func (a *App) Register(userData *models.User) (*models.User, error) {

	user, err := a.Auth.Register(userData)
	if err != nil {
		fmt.Println("Register error: ", err)
		return nil, err
	}

	// Set the user to the auth struct
	a.Auth.User = user
	dbService, err := database.NewDatabase(user)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", Errors.Wrap(err, Errors.DBConnectionFailed, "Failed to initialize database"))
	}
	a.DB = dbService
	a.Events = events.NewEvents(a.DB.GetDB())

	// Initialize daemon manager for the newly registered user
	if user != nil && user.ID > 0 {
		daemonMgr, err := daemon.NewManager(user.ID, a.ctx)
		if err != nil {
			log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to initialize daemon manager").Error())
		} else {
			a.Daemon = daemonMgr

			// Install daemon service for the new user
			if !daemonMgr.IsDaemonInstalled() {
				if err := daemonMgr.InstallDaemon(); err != nil {
					log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to install notification daemon").Error())
					runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
						Type:    runtime.WarningDialog,
						Title:   "Notification Setup",
						Message: "Failed to set up background models. Notifications will only work when the app is running.",
					})
				} else {
					log.Println("Notification daemon installed successfully for new user", user.ID)
				}
			}

			// Start daemon if not running
			if !daemonMgr.IsDaemonRunning() {
				if err := daemonMgr.StartDaemon(); err != nil {
					log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to start notification daemon").Error())
				} else {
					log.Println("Notification daemon started successfully for new user", user.ID)
				}
			}
		}
	}

	return user, nil
}

// Login handles user authentication
func (a *App) Login(username, password string) (*models.User, error) {
	user, err := a.Auth.Login(username, password)
	if err != nil {
		return nil, err
	}

	// Set the user to the auth struct
	a.Auth.User = user

	dbService, err := database.NewDatabase(user)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", Errors.Wrap(err, Errors.DBConnectionFailed, "Failed to initialize database"))
	}
	a.DB = dbService
	a.Events = events.NewEvents(a.DB.GetDB())

	// Initialize daemon manager for the logged-in user
	if user != nil && user.ID > 0 {
		daemonMgr, err := daemon.NewManager(user.ID, a.ctx)
		if err != nil {
			log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to initialize daemon manager").Error())
		} else {
			a.Daemon = daemonMgr

			// Install daemon service only if not already installed
			// This ensures each user gets their own service instance
			if !daemonMgr.IsDaemonInstalled() {
				if err := daemonMgr.InstallDaemon(); err != nil {
					log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to install notification daemon").Error())
					runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
						Type:    runtime.WarningDialog,
						Title:   "Notification Setup",
						Message: "Failed to set up background models. Notifications will only work when the app is running.",
					})
				} else {
					log.Println("Notification daemon installed successfully for user", user.ID)
				}
			}

			// Start daemon if not running
			if !daemonMgr.IsDaemonRunning() {
				if err := daemonMgr.StartDaemon(); err != nil {
					log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to start notification daemon").Error())
				} else {
					log.Println("Notification daemon started successfully for user", user.ID)
				}
			}
		}
	}

	return user, nil
}

// Logout handles user logout
func (a *App) Logout() error {
	// Stop and uninstall daemon service before logout
	// This ensures the service is removed so another user can install their own
	if a.Daemon != nil {
		// Stop the daemon first
		if err := a.Daemon.StopDaemon(); err != nil {
			log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to stop notification daemon").Error())
		} else {
			log.Println("Notification daemon stopped successfully")
		}

		// Uninstall the daemon service
		if err := a.Daemon.UninstallDaemon(); err != nil {
			log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to uninstall notification daemon").Error())
		} else {
			log.Println("Notification daemon uninstalled successfully")
		}

		// Clear daemon manager reference
		a.Daemon = nil
	}

	if err := a.Auth.Logout(); err != nil {
		return err
	}

	authService := auth.NewAuth()

	var dbService *database.Database
	if authService.User != nil {
		var err error
		dbService, err = database.NewDatabase(authService.User)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", Errors.Wrap(err, Errors.DBConnectionFailed, "Failed to initialize database"))
		}
		a.DB = dbService
	}

	var eventsService *events.Events
	if dbService != nil {
		eventsService = events.NewEvents(dbService.GetDB())
	}

	a.Auth = authService
	a.DB = dbService
	a.Events = eventsService

	return nil
}

/********************************************************
GET OPERATIONS
********************************************************/

// IsAuthenticated checks if the user is currently authenticated
func (a *App) GetCurrentUser() (*models.User, error) {
	return a.Auth.User, nil
}

// Sync performs synchronization of local changes with the remote server
func (a *App) Sync() error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	// Check if user is authenticated
	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	// Check if we're online
	if !network.IsOnline() {
		return Errors.Wrap(fmt.Errorf("not online, cannot sync"), Errors.NetworkOffline, "Not online, cannot sync")
	}

	sm := syncservice.NewManager(a.DB.GetDB())

	// Perform the sync
	return sm.ProcessPendingSyncs()
}

func (a *App) GetAuthToken() (string, error) {
	token, err := client.LoadToken()
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetLAssignment returns an assignment by ID
func (a *App) GetLAssignment(id uint) (*models.LocalAssignment, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return a.DB.GetLAssignment(id)
}

// GetCourse returns a course by ID
func (a *App) GetCourse(id uint) (*models.LocalCourse, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return models.GetLCourse(id, a.DB.GetDB())
}

// GetUser returns a user by ID
func (a *App) GetUser(id uint) (*models.User, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return a.DB.GetUser(id)
}

// GetAssignments returns all assignments for the current user
func (a *App) GetAssignments() ([]models.LocalAssignment, error) {
	if a.DB == nil {
		return []models.LocalAssignment{}, nil
	}
	assignments, err := models.GetLAssignments(a.DB.GetDB().Preload("Course").Preload("Documents"))
	if err != nil {
		return nil, Errors.HandleDBReadError(err)
	}
	return assignments, nil
}

// GetCourses returns all courses for the current user
func (a *App) GetCourses() ([]models.LocalCourse, error) {
	if a.DB == nil {
		return []models.LocalCourse{}, nil
	}
	return a.DB.GetCourses()
}

// GetCoursesLinked returns all courses linked for the current user
func (a *App) GetCoursesLinked() ([]models.Course, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	courses, err := client.GetCoursesLinked()
	if err != nil {
		return nil, err
	}
	return courses, nil
}

// GetNotes returns all notes for the current user
func (a *App) GetNotes() ([]models.LocalNote, error) {
	if a.DB == nil {
		return []models.LocalNote{}, nil
	}
	return a.DB.GetNotes()
}

// GetNotifications returns all notifications for the current user
func (a *App) GetNotifications() ([]models.LocalNotification, error) {
	if a.DB == nil {
		return []models.LocalNotification{}, nil
	}
	return a.DB.GetNotifications()
}

// GetSupportDocuments retrieves only support documents for an assignment
func (a *App) GetSupportDocuments(assignmentID uint) ([]models.LocalDocument, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	var documents []models.LocalDocument
	err := a.DB.GetDB().Where(
		"assignment_id = ? AND type = ? AND has_local_file = ?",
		assignmentID, models.DocumentTypeSupport, true,
	).Order("created_at DESC").Find(&documents).Error

	return documents, err
}

// GetSubmissionDocuments retrieves only submission documents for an assignment
func (a *App) GetSubmissionDocuments(assignmentID uint) ([]models.LocalDocument, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	var documents []models.LocalDocument
	err := a.DB.GetDB().Where(
		"assignment_id = ? AND type = ? AND has_local_file = ?",
		assignmentID, models.DocumentTypeSubmission, true,
	).Order("created_at DESC").Find(&documents).Error

	return documents, err
}

func (a *App) SaveUIMessage(assignmentID uint, vercelMessage map[string]interface{}) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return a.DB.SaveUIMessage(assignmentID, vercelMessage)
}

func (a *App) GetConversationHistory(assignmentID uint) ([]models.LocalAiMessage, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return a.DB.GetConversationHistory(assignmentID)
}

// OpenDocument opens a document file with the system default application
func (a *App) OpenDocument(documentID uint) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	// Get local document record
	var doc models.LocalDocument
	if err := a.DB.GetDB().Where("id = ?", documentID).First(&doc).Error; err != nil {
		return Errors.Wrap(err, Errors.DBRecordNotFound, "Document not found or access denied")
	}

	// Check if we have the file locally
	if !doc.HasLocalFile {
		return Errors.Wrap(fmt.Errorf("file not available offline"), Errors.NetworkOffline, "File not available offline - please sync to download")
	}

	// Check if file actually exists on disk
	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		// Update database to reflect missing file
		a.DB.GetDB().Model(&doc).Update("has_local_file", false)
		return Errors.Wrap(err, Errors.FSFileNotFound, "File not found on disk")
	}

	// Open with system default application
	runtime.BrowserOpenURL(a.ctx, "file://"+doc.FilePath)
	return nil
}

// SaveDocumentAs opens a save dialog and copies the document to chosen location
func (a *App) SaveDocumentAs(documentID uint) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	// Get local document record
	var doc models.LocalDocument
	if err := a.DB.GetDB().Where("id = ?", documentID).First(&doc).Error; err != nil {
		return Errors.Wrap(err, Errors.DBRecordNotFound, "Document not found or access denied")
	}

	// Check if we have the file locally
	if !doc.HasLocalFile {
		return Errors.Wrap(fmt.Errorf("file not available offline"), Errors.NetworkOffline, "File not available offline - please sync to download")
	}

	// Check if file actually exists on disk
	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		// Update database to reflect missing file
		a.DB.GetDB().Model(&doc).Update("has_local_file", false)
		return Errors.Wrap(err, Errors.FSFileNotFound, "File not found on disk")
	}

	// Open save dialog
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           fmt.Sprintf("Save %s", doc.FileName),
		DefaultFilename: doc.FileName,
	})

	if err != nil {
		return Errors.Wrap(err, Errors.FSOpenFailed, "Failed to open save dialog")
	}

	if savePath == "" {
		return Errors.Wrap(fmt.Errorf("no save location selected"), Errors.ValidationRequired, "No save location selected")
	}

	// Copy file
	sourceFile, err := os.Open(doc.FilePath)
	if err != nil {
		return Errors.Wrap(err, Errors.FSOpenFailed, "Failed to open source file")
	}
	defer sourceFile.Close()

	destFile, err := os.Create(savePath)
	if err != nil {
		return Errors.Wrap(err, Errors.FSCreateFailed, "Failed to create destination file")
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return Errors.Wrap(err, Errors.StorageCopyFailed, "Failed to copy file")
	}

	return nil
}

// GetUserStorageInfo returns storage statistics for the current user
func (a *App) GetUserStorageInfo() (*models.DocumentStorage, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	userID := a.Auth.User.ID

	// Calculate storage info on-demand
	storageInfo, err := models.GetUserStorageInfo(userID, a.DB.GetDB())
	if err != nil {
		return nil, Errors.Wrap(err, Errors.DBQueryFailed, "Failed to get storage info")
	}

	return storageInfo, nil
}

func (a *App) GetCourseAssignments(course *models.LocalCourse) ([]models.LocalAssignment, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	var assignments []models.LocalAssignment
	err := a.DB.GetDB().Where("course_code = ?", course.Code).Find(&assignments).Order("created_at ASC").Error
	return assignments, err
}

// GetRemoteUsers returns all users from the remote server
func (a *App) GetRemoteUsers() ([]client.RemoteUser, error) {
	if !a.Auth.IsAuthenticated() {
		return []client.RemoteUser{}, nil
	}
	users, err := client.GetRemoteUsers()
	if err != nil {
		return []client.RemoteUser{}, err
	}
	return users, nil
}

// Follow a user
func (a *App) Follow(followedID uint) (bool, error) {
	return client.Follow(followedID)
}

// FollowResponse represents the response for follow-related queries
type FollowResponse struct {
	Users []models.User `json:"users"`
	Count int           `json:"count"`
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
	if a.Daemon == nil {
		return Errors.Wrap(fmt.Errorf("daemon manager not initialized"), Errors.InitDatabaseNotInitialized, "Daemon manager not initialized")
	}
	return a.Daemon.InstallDaemon()
}

// UninstallNotificationDaemon uninstalls the notification daemon
func (a *App) UninstallNotificationDaemon() error {
	if a.Daemon == nil {
		return Errors.Wrap(fmt.Errorf("daemon manager not initialized"), Errors.InitDatabaseNotInitialized, "Daemon manager not initialized")
	}
	return a.Daemon.UninstallDaemon()
}

// IsNotificationDaemonInstalled checks if the daemon is installed
func (a *App) IsNotificationDaemonInstalled() bool {
	if a.Daemon == nil {
		return false
	}
	return a.Daemon.IsDaemonInstalled()
}

// IsNotificationDaemonRunning checks if the daemon is running
func (a *App) IsNotificationDaemonRunning() bool {
	if a.Daemon == nil {
		return false
	}
	return a.Daemon.IsDaemonRunning()
}

// StartNotificationDaemon starts the daemon
func (a *App) StartNotificationDaemon() error {
	if a.Daemon == nil {
		return Errors.Wrap(fmt.Errorf("daemon manager not initialized"), Errors.InitDatabaseNotInitialized, "Daemon manager not initialized")
	}
	return a.Daemon.StartDaemon()
}

// StopNotificationDaemon stops the daemon
func (a *App) StopNotificationDaemon() error {
	if a.Daemon == nil {
		return Errors.Wrap(fmt.Errorf("daemon manager not initialized"), Errors.InitDatabaseNotInitialized, "Daemon manager not initialized")
	}
	return a.Daemon.StopDaemon()
}

// GetNotificationDaemonStatus returns the daemon status
func (a *App) GetNotificationDaemonStatus() map[string]interface{} {
	if a.Daemon == nil {
		return map[string]interface{}{
			"installed": false,
			"running":   false,
			"error":     "Daemon manager not initialized",
		}
	}

	return map[string]interface{}{
		"installed": a.Daemon.IsDaemonInstalled(),
		"running":   a.Daemon.IsDaemonRunning(),
		"error":     nil,
	}
}

// Add method to rebuild daemon (for updates)
func (a *App) RebuildNotificationDaemon() error {
	if a.Daemon == nil {
		return Errors.Wrap(fmt.Errorf("daemon manager not initialized"), Errors.InitDatabaseNotInitialized, "Daemon manager not initialized")
	}
	return a.Daemon.RebuildDaemon()
}

// LinkCourse links a course to a list of users
func (a *App) RequestLinkCourse(c *models.LocalCourse, usersID []uint) error {
	return client.RequestLinkCourse(c, usersID)
}

func (a *App) AcceptLink(courseData string) error {

	// Unmarshal the course data (Base model from server)
	var c models.Course
	if err := json.Unmarshal([]byte(courseData), &c); err != nil {
		return err
	}

	// Convert Base Course to LocalCourse for local operations
	localC := c.ToLocal()

	//Determine if the course already exists
	var existingCourse models.LocalCourse
	err := a.DB.GetDB().Where("code = ?", localC.Code).First(&existingCourse).Error
	if err != nil {
		// If the course doesn't exist, create it
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.CreateCourse(localC)
		}
		return err
	}
	// Update the course with the new link ID
	//a.UpdateCourse(&existingCourse, "link_id", c.LinkID.String())

	assignments, err := client.AcceptLinkCourse(&c)
	if err != nil {
		return err
	}

	// Create the assignments
	for _, remoteAssignment := range assignments {

		// Convert Base Assignment to LocalAssignment
		localAssignment := remoteAssignment.ToLocal()

		// Set status for new assignment
		localAssignment.Status = "Not started"

		var newAssignment *models.LocalAssignment
		if newAssignment, err = a.CreateAssignment(localAssignment); err != nil {
			return err
		}

		for _, document := range remoteAssignment.Documents {
			// Convert Base Document to LocalDocument
			localDoc := document.ToLocal()
			uploadReq := fileops.FileUploadRequest{
				AssignmentID:       newAssignment.ID, // Update to new assignment ID
				RemoteAssignmentID: newAssignment.RemoteID,
				UserID:             a.Auth.User.ID, // Update to new user ID
				Type:               localDoc.Type,
				FileName:           localDoc.FileName,
				FileSize:           localDoc.FileSize,
				StorageKey:         localDoc.StorageKey,
			}
			_, err := a.CreateDocument(uploadReq, false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) AcceptAssignment(assignmentData string) error {

	// Unmarshal Base Assignment from server
	var remoteAssignment models.Assignment
	if err := json.Unmarshal([]byte(assignmentData), &remoteAssignment); err != nil {
		return err
	}

	// Convert Base Assignment to LocalAssignment
	localAssignment := remoteAssignment.ToLocal()

	if _, err := a.CreateAssignment(localAssignment); err != nil {
		return err
	}

	return nil
}

func (a *App) AcceptDocument(documentData string) error {

	// Unmarshal Base Document from server
	var remoteDocument models.Document
	if err := json.Unmarshal([]byte(documentData), &remoteDocument); err != nil {
		return err
	}

	// Convert Base Document to LocalDocument
	localDoc := remoteDocument.ToLocal()

	var assignment *models.LocalAssignment
	err := a.DB.GetDB().Where("parent_id = ?", remoteDocument.AssignmentID).First(&assignment).Error
	if err != nil {
		return err
	}

	uploadReq := fileops.FileUploadRequest{
		AssignmentID:       assignment.ID,
		RemoteAssignmentID: assignment.RemoteID,
		UserID:             a.Auth.User.ID,
		Type:               localDoc.Type,
		FileName:           localDoc.FileName,
		FileSize:           localDoc.FileSize,
		StorageKey:         localDoc.StorageKey,
	}

	if _, err := a.CreateDocument(uploadReq, false); err != nil {
		return err
	}

	return nil
}

func (a *App) AcceptNote(noteData string) error {

	// Unmarshal Base Note from server
	var remoteNote models.Note
	if err := json.Unmarshal([]byte(noteData), &remoteNote); err != nil {
		return err
	}

	// Convert Base Note to LocalNote
	localNote := remoteNote.ToLocal()

	if err := a.CreateNote(localNote); err != nil {
		return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to create note")
	}

	return nil

}
