package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"unipilot/internal/auth"
	"unipilot/internal/client"

	Errors "unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/network"
	"unipilot/internal/services/daemon"
	"unipilot/internal/services/database"
	"unipilot/internal/services/fileops"
	"unipilot/internal/services/fileops/progress"
	"unipilot/internal/services/utils"
)

// App struct
type App struct {
	ctx    context.Context
	Auth   *auth.Auth
	DB     *database.Database
	Daemon *daemon.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {

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

	return &App{
		Auth: authService,
		DB:   dbService,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Check if user is already authenticated and initialize HTTP client if needed
	if a.Auth.IsAuthenticated() {

		// Service will be installed on login, not on startup
		daemon, err := daemon.NewManager(a.Auth.User.ID, a.ctx)
		if err != nil {
			log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Daemon manager initialization failed").Error())
		} else {
			a.Daemon = daemon
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
			Title:    assignmentData.Title,
			Todo:     assignmentData.Todo,
			Deadline: assignmentData.Deadline,
			CourseID: assignmentData.CourseID,
			Type:     assignmentData.Type,
			Status:   assignmentData.Status,
			Priority: assignmentData.Priority,
			Link:     assignmentData.Link,
			ParentID: assignmentData.ParentID,
		},
	}

	if err := localAssignment.Validate(); err != nil {
		return nil, err
	}

	// Create the assignment locally first
	if err := db.Create(localAssignment).Error; err != nil {
		return nil, Errors.HandleDBWriteError(err)
	}

	// Always try to sync with server

	remoteAssignment := localAssignment.ToRemote(a.Auth.User.ID)

	//Online operation
	isOnline := network.IsOnline()
	if isOnline {
		if err := client.CreateAssignment(remoteAssignment); err != nil {
			return nil, Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to create assignment")
		}
		if err := db.Model(localAssignment).Update("synced_at", time.Now()).Error; err != nil {
			return nil, Errors.HandleDBWriteError(err)
		}
	}

	return localAssignment, nil
}

func (a *App) CopyAssignment(assignment *models.LocalAssignment, includeDocuments bool) error {

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	newAssignment, err := a.CreateAssignment(assignment)
	if err != nil {
		return Errors.Wrap(err, Errors.InternalError, "Failed to create assignment")
	}

	if includeDocuments {
		for _, document := range assignment.Documents {

			newDocument := &models.LocalDocument{
				Base: models.Base{
					ID: "",
				},
				BaseDocument: models.BaseDocument{
					AssignmentID: newAssignment.ID,
					Type:         document.Type,
					FileName:     document.FileName,
					FileSize:     document.FileSize,
					StorageKey:   document.StorageKey,
					HasLocalFile: false,
					Version:      1,
				},
			}
			newDocument.FilePath = fileops.GetFilePath(newDocument)
			if err := a.CreateDocument(newDocument); err != nil {
				return Errors.Wrap(err, Errors.InternalError, "Failed to create document")
			}
		}
	}

	return nil
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

	if courseData.ClusterID != nil {
		localCourse.ClusterID = courseData.ClusterID
	}

	if err := localCourse.Validate(); err != nil {
		return err
	}

	// Create the course within the transaction
	if err := db.Create(localCourse).Error; err != nil {
		return Errors.HandleDBWriteError(err)
	}

	// Convert LocalCourse to Course (Base model) for server API
	remoteCourse := localCourse.ToRemote(a.Auth.User.ID)
	isOnline := network.IsOnline()
	if isOnline {
		if err := client.CreateCourse(remoteCourse); err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to create course")
		}
		if err := db.Model(localCourse).Update("synced_at", time.Now()).Error; err != nil {
			return Errors.HandleDBWriteError(err)
		}

	}

	return nil
}

func (a *App) CreateNote(noteData models.LocalNote) error {

	if !network.IsOnline() {
		return Errors.Wrap(fmt.Errorf("no network connection"), Errors.NetworkOffline, "No network connection")
	}

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}

	if err := noteData.Validate(); err != nil {
		return err
	}

	db := a.DB.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {

		newNote := &noteData

		log.Println("note Before Create", newNote.ID)

		if err := tx.Create(newNote).Error; err != nil {
			log.Println("note Create Error", err)
			return Errors.HandleDBWriteError(err)
		}

		log.Println("note After Create", newNote.ID)

		remoteNote := newNote.ToRemote(a.Auth.User.ID)

		isOnline := network.IsOnline()
		if isOnline {

			if err := client.CreateNote(remoteNote); err != nil {
				return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to create note")
			}

			if err := tx.Model(newNote).Update("synced_at", time.Now()).Error; err != nil {
				return Errors.HandleDBWriteError(err)
			}

			log.Println("note After Update", newNote.ID)
		}

		return nil
	})
	if err != nil {
		return Errors.Wrap(err, Errors.DBTransactionFailed, "Failed to create note")
	}

	return nil

}

type FileInfo struct {
	FileName string
	FileSize int64
}

func (a *App) GetFileInfo(filePath string) (*FileInfo, error) {
	fileName := fileops.GetFileName(filePath)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, Errors.Wrap(err, Errors.FSFileNotFound, "File not found")
	}
	fileSize := fileInfo.Size()
	return &FileInfo{
		FileName: fileName,
		FileSize: fileSize,
	}, nil
}

func (a *App) PickFile() (string, error) {
	filePath, err := fileops.PickFile(a.ctx)
	if err != nil {
		return "", Errors.Wrap(err, Errors.FSOpenFailed, "Failed to pick file")
	}
	return filePath, nil
}

// UploadDocument opens a file dialog and uploads a document to an assignment
func (a *App) UploadDocument(documentID, assignmentID, documentType, filePath string) (*models.LocalDocument, error) {

	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}
	var err error
	if filePath == "" {
		filePath, err = fileops.PickFile(a.ctx)
		if err != nil {
			return nil, Errors.Wrap(err, Errors.FSOpenFailed, "Failed to pick file")
		}
	}

	if filePath == "" {
		return nil, Errors.Wrap(fmt.Errorf("no file selected"), Errors.ValidationRequired, "No file selected")
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, Errors.Wrap(err, Errors.FSFileNotFound, "File not found")
	}

	// Create upload request
	document := &models.LocalDocument{
		Base: models.Base{
			ID: documentID,
		},
		BaseDocument: models.BaseDocument{
			AssignmentID: assignmentID,
			Type:         models.DocumentType(documentType),
			FileName:     fileops.GetFileName(filePath),
			FilePath:     filePath,
			FileSize:     fileInfo.Size(),
			Version:      1,
			HasLocalFile: true,
		},
	}

	log.Println("Uploading document", document.ID, document.FileSize)

	err = a.CreateDocument(document)
	if err != nil {
		return nil, err
	}

	return document, nil

}

func (a *App) CreateDocument(document *models.LocalDocument) error {

	log.Println("Creating document (app)", document.ID, document.FileSize)

	var err error

	err = a.DB.CreateDocument(a.ctx, document)
	if err != nil {
		return err
	}

	log.Println("Document created (app)", document.ID, document.FileSize, document.HasLocalFile)

	if document.HasLocalFile {
		return a.sendDocumentWithProgress(document)
	} else {
		err = a.SendDocument(document)
		if err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to send document")
		}
	}

	return nil
}

func (a *App) sendDocumentWithProgress(document *models.LocalDocument) error {

	// Initialize progress manager (reuse if exists, or create new)
	progressManager := progress.GetManager(a.ctx)

	// Create progress tracker
	fileProgress := progressManager.Create(document.ID, document.FileSize)

	// Register progress callback to emit events to frontend
	fileProgress.OnProgress(func(p *progress.Progress) {
		snapshot := p.Snapshot()

		if snapshot.Error != nil {
			runtime.EventsEmit(a.ctx, fmt.Sprintf("upload:error:%s", document.ID), map[string]interface{}{
				"upload_id": document.ID,
				"error":     snapshot.Error.Error(),
			})
		} else {
			snapshot.Percentage = snapshot.Percentage * 0.2
			runtime.EventsEmit(a.ctx, fmt.Sprintf("upload:progress:%s", document.ID), snapshot)
		}

	})

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()
	// Upload to server

	// Send document with progress tracking
	storageKey, clientErr := client.SendDocumentWithProgress(a.ctx, document, fileProgress)
	if clientErr != nil {
		return Errors.Wrap(clientErr, Errors.ClientRequestFailed, "Failed to send document")
	}

	if err := db.Model(document).
		Update("synced_at", time.Now()).
		Update("storage_key", storageKey).Error; err != nil {
		return Errors.HandleDBWriteError(err)
	}

	return nil
}

func (a *App) SendDocument(document *models.LocalDocument) error {

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if a.Auth.IsAuthenticated() {

		db := a.DB.GetDB()

		storageKey, clientErr := client.SendDocument(document)
		if clientErr != nil {
			return Errors.Wrap(clientErr, Errors.ClientRequestFailed, "Failed to send document")
		}

		if err := db.Model(document).
			Update("synced_at", time.Now()).
			Update("storage_key", storageKey).Error; err != nil {
			return Errors.HandleDBWriteError(err)
		}

	}

	return nil

}

// GetActiveUploads returns all active uploads
func (a *App) GetActiveUploads() []progress.ProgressSnapshot {
	return progress.GetManager(a.ctx).List()
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

// DownloadDocument retrieves a document file for download
func (a *App) DownloadDocument(document *models.LocalDocument) error {

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	// Send document to Server to create remote document & download from cloud
	downloadResp, err := client.DownloadDocument(a.ctx, document)
	if err != nil {
		return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to download document")
	}

	// test if file is empty
	if downloadResp == nil {
		return Errors.Wrap(fmt.Errorf("file not found"), Errors.FSFileNotFound, "File not found")
	}

	filePath := fileops.GetFilePath(document)

	// write file to disk
	if err := fileops.WriteFile(filePath, downloadResp); err != nil {
		return Errors.Wrap(err, Errors.FSWriteFailed, "Failed to write file")
	}

	if err := db.Model(document).
		Update("has_local_file", true).
		Update("file_path", filePath).Error; err != nil {
		return Errors.HandleDBWriteError(err)
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

func (a *App) DeleteDocumentRAG(assignmentID, documentID string) error {

	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	if err := client.DeleteDocumentRAG(assignmentID, documentID); err != nil {
		return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to delete document from RAG")
	}

	return nil
}

func (a *App) GetAssignmentDocumentIDsRAG(assignmentID string) ([]string, error) {
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

	isOnline := network.IsOnline()

	if isOnline {
		if err := client.UpdateAssignment(LocalAssignment.ID, column, value); err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to update assignment")
		}
		if err := db.Model(LocalAssignment).Update("synced_at", time.Now()).Error; err != nil {
			return Errors.HandleDBWriteError(err)
		}
	}
	return nil
}

// UpdateCourse updates an existing course
func (a *App) UpdateCourse(course *models.LocalCourse, column, value string) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	db := a.DB.GetDB()

	if err := db.Model(course).Update(column, value).Error; err != nil {
		return Errors.HandleDBWriteError(err)
	}

	isOnline := network.IsOnline()
	if isOnline {
		if err := client.UpdateCourse(course.ID, column, value); err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to update course")
		}
		if err := db.Model(course).Update("synced_at", time.Now()).Error; err != nil {
			return Errors.HandleDBWriteError(err)
		}
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

	isOnline := network.IsOnline()
	if isOnline {
		if err := client.UpdateNote(LocalNote.ID, column, value); err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to update note")
		}
		if err := db.Model(LocalNote).Update("synced_at", time.Now()).Error; err != nil {
			return Errors.HandleDBWriteError(err)
		}
	}

	return nil
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
	if isOnline {
		if err := client.UpdateUser(column, value); err != nil {
			return nil, Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to update user")
		}
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

	isOnline := network.IsOnline()
	if isOnline {
		if err := client.DeleteAssignment(assignment.ID); err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to delete assignment")
		}
		if err := db.Model(assignment).Update("synced_at", time.Now()).Error; err != nil {
			return Errors.HandleDBWriteError(err)
		}
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

	isOnline := network.IsOnline()
	if isOnline {
		if err := client.DeleteCourse(course.ID); err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to delete course")
		}
		if err := db.Model(course).Update("synced_at", time.Now()).Error; err != nil {
			return Errors.HandleDBWriteError(err)
		}
	}

	return nil
}

// DeleteDocument removes a document and its file
func (a *App) DeleteDocument(documentID string) error {
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

	db := a.DB.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Delete(&doc).Error; err != nil {
			return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to delete document record")
		}

		// Also store metadata remotely for sharing
		if !a.Auth.IsAuthenticated() {
			return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
		}

		if err := client.DeleteDocument(doc.ID); err != nil {
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

	if err := a.DB.DeleteNote(note); err != nil {
		return err
	}

	isOnline := network.IsOnline()
	if isOnline {
		if err := client.DeleteNote(note.ID); err != nil {
			return Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to delete note")
		}
	}

	return nil
}

// ========================================
// OTHER OPERATIONS
// ========================================

// Register handles user registration
func (a *App) Register(userData *models.User) (*models.User, error) {

	dbService, authService, err := a.Auth.Register(userData)
	if err != nil {
		fmt.Println("Register error: ", err)
		return nil, err
	}

	a.Auth = authService
	a.DB = dbService
	user := authService.User

	// Initialize daemon manager for the newly registered user
	if user != nil && user.ID != "" {
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
	dbService, authService, err := auth.Login(username, password)
	if err != nil {
		return nil, err
	}
	// Set the auth and db services
	a.Auth = authService
	a.DB = dbService

	// Get the user from the auth service
	user := authService.User

	// Initialize daemon manager for the logged-in user
	if user != nil && user.ID != "" {
		daemonMgr, err := daemon.NewManager(user.ID, a.ctx)
		if err != nil {
			log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to initialize daemon manager").Error())
		} else {
			a.Daemon = daemonMgr

			// Install daemon service only if not already installed
			// This ensures each user gets their own service instance
			fmt.Println("Checking if daemon is installed for user", daemonMgr.IsDaemonInstalled())
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

	a.Auth = authService
	a.DB = dbService

	//Stop and uninstall daemon service before logout
	//This ensures the service is removed so another user can install their own

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic in logout daemon cleanup:", r)
			}
		}()

		if a.Daemon != nil {

			// Stop the daemon first
			if err := a.Daemon.StopDaemon(); err != nil {
				log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to stop notification daemon"))
			} else {
				log.Println("Notification daemon stopped successfully")
			}

			// Uninstall the daemon service
			if err := a.Daemon.UninstallDaemon(); err != nil {
				log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to uninstall notification daemon"))
			} else {
				log.Println("Notification daemon uninstalled successfully")
			}

			// Clear daemon manager reference
			a.Daemon = nil
		}
	}()

	return nil

}

/********************************************************
GET OPERATIONS
********************************************************/

// IsAuthenticated checks if the user is currently authenticated
func (a *App) GetCurrentUser() (*models.User, error) {
	return a.Auth.User, nil
}

func (a *App) GetAuthToken() (string, error) {
	token, err := client.GetAuthToken()
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetAssignments returns all assignments for the current user
func (a *App) GetAssignments() ([]models.LocalAssignment, error) {
	if a.DB == nil {
		return []models.LocalAssignment{}, nil
	}
	assignments, err := models.GetLAssignments(
		a.DB.GetDB().
			Preload("Course").
			Preload("Documents").
			Order("deadline DESC").
			Order("created_at DESC"))
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

func (a *App) GetRCourses(userID string) ([]models.Course, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	courses, err := client.GetCourses(userID)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

// GetCourse returns a course by ID
func (a *App) GetCourse(id string) (*models.LocalCourse, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return models.GetLCourse(id, a.DB.GetDB())
}

// GetUser returns a user by ID
func (a *App) GetUser(id string) (*models.User, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return a.DB.GetUser(id)
}

func (a *App) GetUserCourseInvitations() ([]models.CourseInvitation, error) {
	invitations, err := client.GetUserCourseInvitations()
	if err != nil {
		return nil, err
	}
	return invitations, nil
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

// GetSupportDocuments retrieves only support documents for an assignment
func (a *App) GetSupportDocuments(assignmentID string) ([]models.LocalDocument, error) {
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
func (a *App) GetSubmissionDocuments(assignmentID string) ([]models.LocalDocument, error) {
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

func (a *App) SaveUIMessage(assignmentID string, vercelMessage map[string]interface{}) error {
	if a.DB == nil {
		return Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return a.DB.SaveUIMessage(assignmentID, vercelMessage)
}

func (a *App) GetConversationHistory(assignmentID string) ([]models.LocalAiMessage, error) {
	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}
	return a.DB.GetConversationHistory(assignmentID)
}

// OpenDocument opens a document file with the system default application
func (a *App) OpenDocument(documentID string) error {
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
func (a *App) SaveDocumentAs(documentID string) error {
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

	// Calculate storage info on-demand
	storageInfo, err := models.GetLocalStorageInfo(a.DB.GetDB())
	if err != nil {
		return nil, Errors.Wrap(err, Errors.DBQueryFailed, "Failed to get storage info")
	}

	return storageInfo, nil
}

func (a *App) GetAssignmentStorageInfo(assignmentID string) (*models.LocalAssignmentStorage, error) {

	if a.DB == nil {
		return nil, Errors.Wrap(fmt.Errorf("database not initialized"), Errors.InitDatabaseNotInitialized, "Database not initialized")
	}

	storageInfo, err := models.GetLocalAssignmentStorage(assignmentID, a.DB.GetDB())
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
func (a *App) GetRemoteUsers(cursor *models.Cursor, limit int) (*models.PageResponse[models.User], error) {
	if !a.Auth.IsAuthenticated() {
		return nil, Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}
	users, err := client.GetRemoteUsers(cursor, limit)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (a *App) GetFriendShipStatus(userID string) (*client.FriendStatusResponse, error) {
	if !a.Auth.IsAuthenticated() {
		return nil, nil
	}
	friendshipStatus, err := client.GetFriendShipStatus(userID)
	if err != nil {
		return nil, err
	}
	return friendshipStatus, nil
}

func (a *App) GetFriends(userID string, cursor *models.Cursor, limit int) (*models.PageResponse[models.User], error) {
	if !a.Auth.IsAuthenticated() {
		return nil, nil
	}
	users, err := client.GetFriends(userID, cursor, limit)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (a *App) SendFriendRequest(userID string) error {
	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}
	err := client.SendFriendRequest(userID)
	if err != nil {
		return Errors.Wrap(err, Errors.SysExecFailed, "Failed to send friend request")
	}
	return nil
}

func (a *App) AcceptFriendRequest(userID string) error {
	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}
	err := client.AcceptFriendRequest(userID)
	if err != nil {
		return Errors.Wrap(err, Errors.SysExecFailed, "Failed to accept friend request")
	}
	return nil
}

func (a *App) CancelFriendRequest(userID string) error {
	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}
	err := client.CancelFriendRequest(userID)
	if err != nil {
		return Errors.Wrap(err, Errors.SysExecFailed, "Failed to cancel friend request")
	}
	return nil
}

func (a *App) RemoveFriend(userID string) error {
	if !a.Auth.IsAuthenticated() {
		return Errors.Wrap(fmt.Errorf("user not authenticated"), Errors.InitUserNotAuthenticated, "User not authenticated")
	}
	err := client.RemoveFriend(userID)
	if err != nil {
		return Errors.Wrap(err, Errors.SysExecFailed, "Failed to remove friend")
	}
	return nil
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

func (a *App) GetClusterStatus(courseID string) (*client.CourseStatusResponse, error) {
	if !a.Auth.IsAuthenticated() {
		return nil, nil
	}
	courseStatus, err := client.GetClusterStatus(courseID)
	if err != nil {
		return nil, err
	}
	return courseStatus, nil
}

// LinkCourse links a course to a list of users
func (a *App) CourseShare(c *models.LocalCourse, usersID []string) error {
	if err := client.CourseShare(c, usersID); err != nil {
		return err
	}
	return nil
}

func (a *App) SendClusterRequest(courseID string) error {
	if err := client.SendClusterRequest(courseID); err != nil {
		return err
	}
	return nil
}

func (a *App) AcceptCourseInvitation(invitation *models.CourseInvitation) error {

	err := client.AcceptCourseInvitation(invitation)
	if err != nil {
		log.Println(Errors.Wrap(err, Errors.SysExecFailed, "Failed to accept course invitation").Error())
	}

	return nil
}

func (a *App) DeclineCourseInvitation(invitation *models.CourseInvitation) error {
	if err := client.DeclineCourseInvitation(invitation); err != nil {
		return err
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

	localDoc.ID = ""                      // will be set by the database
	localDoc.AssignmentID = assignment.ID // Link to the assignment
	localDoc.Version = 1                  // New document
	localDoc.HasLocalFile = false         // No local file yet

	if err := a.CreateDocument(localDoc); err != nil {
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

	if err := a.CreateNote(*localNote); err != nil {
		return Errors.Wrap(err, Errors.DBQueryFailed, "Failed to create note")
	}

	return nil

}
