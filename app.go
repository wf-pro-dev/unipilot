package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unipilot/internal/app"
	"unipilot/internal/auth"
	"unipilot/internal/client"
	"unipilot/internal/events"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"
	"unipilot/internal/models/user"
	"unipilot/internal/services/fileops"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	Auth   *auth.Auth
	Events *events.Events
	DB     *app.DatabaseHelper
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
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize database helper
	dbHelper, err := app.NewDatabaseHelper()
	if err != nil {
		fmt.Printf("Warning: Could not initialize database helper: %v\n", err)
	} else {
		a.DB = dbHelper
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Login handles user authentication
func (a *App) Login(username, password string) error {
	if err := a.Auth.Login(username, password); err != nil {
		return err
	}

	// Create a dedicated HTTP client for SSE connections
	sseClient, err := client.NewSSEClient()
	if err != nil {
		return fmt.Errorf("failed to create SSE client: %w", err)
	}

	go a.Auth.SSE.Connect(sseClient)

	a.Events.Start(a.Auth.SSE)

	// Reinitialize database helper after login
	dbHelper, err := app.NewDatabaseHelper()
	if err != nil {
		fmt.Printf("Warning: Could not initialize database helper after login: %v\n", err)
	} else {
		a.DB = dbHelper
	}

	return nil
}

// Logout handles user logout
func (a *App) Logout() error {
	if err := a.Auth.Logout(); err != nil {
		return err
	}

	a.Auth.SSE.StopListener()
	a.Events.Stop()
	a.DB = nil

	return nil
}

// IsAuthenticated checks if the user is currently authenticated
func (a *App) IsAuthenticated() bool {
	return a.Auth.IsAuthenticated()
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

// Document Management Methods

// UploadDocument opens a file dialog and uploads a document to an assignment
func (a *App) UploadDocument(assignmentID uint, documentType string) (*document.Document, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
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
	if a.Auth.IsAuthenticated() && a.Auth.Client != nil {
		metadataReq := map[string]interface{}{
			"assignment_id": assignmentID,
			"type":          documentType,
			"file_name":     filepath.Base(filePath),
			"file_type":     fileops.GetMimeType(filepath.Base(filePath)),
			"file_size":     fileInfo.Size(),
		}

		// Make API call to store metadata (non-blocking, continue if it fails)
		go func() {
			jsonData, _ := json.Marshal(metadataReq)
			resp, err := a.Auth.Client.Post("https://newsroom.dedyn.io/acc-homework/documents/metadata",
				"application/json", strings.NewReader(string(jsonData)))
			if err == nil {
				defer resp.Body.Close()
			}
			// We don't block on this - local file upload is the priority
		}()
	}

	return response.Document, nil
}

// GetAssignmentDocuments retrieves all documents for an assignment
func (a *App) GetAssignmentDocuments(assignmentID uint) ([]document.Document, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	return document.GetLatestVersions(assignmentID, userID, a.DB.GetDB())
}

// GetSupportDocuments retrieves only support documents for an assignment
func (a *App) GetSupportDocuments(assignmentID uint) ([]document.Document, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	var documents []document.Document
	err := a.DB.GetDB().Preload("User").
		Where("assignment_id = ? AND user_id = ? AND type = ?", assignmentID, userID, document.DocumentTypeSupport).
		Order("created_at DESC").
		Find(&documents).Error

	return documents, err
}

// GetSubmissionDocuments retrieves only submission documents for an assignment
func (a *App) GetSubmissionDocuments(assignmentID uint) ([]document.Document, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	var documents []document.Document
	err := a.DB.GetDB().Preload("User").
		Where("assignment_id = ? AND user_id = ? AND type = ?", assignmentID, userID, document.DocumentTypeSubmission).
		Order("created_at DESC").
		Find(&documents).Error

	return documents, err
}

// OpenDocument opens a document file with the system default application
func (a *App) OpenDocument(documentID uint) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	// Get document record
	var doc document.Document
	if err := a.DB.DB.Where("id = ? AND user_id = ?", documentID, userID).First(&doc).Error; err != nil {
		return fmt.Errorf("document not found or access denied")
	}

	// Check if file exists
	if !doc.FileExists() {
		return fmt.Errorf("file not found on disk")
	}

	// Get full path
	fullPath, err := doc.GetFullPath()
	if err != nil {
		return fmt.Errorf("failed to get file path: %w", err)
	}

	// Open with system default application
	runtime.BrowserOpenURL(a.ctx, "file://"+fullPath)
	return nil
}

// SaveDocumentAs opens a save dialog and copies the document to chosen location
func (a *App) SaveDocumentAs(documentID uint) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}

	// Get document record
	var doc document.Document
	if err := a.DB.DB.Where("id = ? AND user_id = ?", documentID, currentUser.ID).First(&doc).Error; err != nil {
		return fmt.Errorf("document not found or access denied")
	}

	// Check if file exists
	if !doc.FileExists() {
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

	// Get source file path
	sourcePath, err := doc.GetFullPath()
	if err != nil {
		return fmt.Errorf("failed to get source file path: %w", err)
	}

	// Copy file
	sourceFile, err := os.Open(sourcePath)
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

// DeleteDocument removes a document and its file
func (a *App) DeleteDocument(documentID uint) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	return fileops.DeleteDocument(documentID, userID, a.DB.GetDB())
}

// GetUserStorageInfo returns storage statistics for the current user
func (a *App) GetUserStorageInfo() (*document.DocumentStorageInfo, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	return fileops.GetUserStorageInfo(userID, a.DB.GetDB())
}

// UploadNewDocumentVersion uploads a new version of an existing document
func (a *App) UploadNewDocumentVersion(existingDocumentID uint) (*document.Document, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
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

	// Upload new version
	response, err := fileops.UploadNewVersion(existingDocumentID, filepath.Base(filePath), file, fileInfo.Size(), a.DB.GetDB())
	if err != nil {
		return nil, fmt.Errorf("version upload failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("version upload failed: %s", response.Message)
	}

	return response.Document, nil
}

// GetRemoteDocumentMetadata retrieves document metadata from remote server (for shared assignments)
func (a *App) GetRemoteDocumentMetadata(assignmentID uint) ([]map[string]interface{}, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
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

// CreateAssignment creates a new assignment
func (a *App) CreateAssignment(assignment *assignment.Assignment) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.CreateAssignment(assignment)
}

// UpdateAssignment updates an existing assignment
func (a *App) UpdateAssignment(LocalAssignment *assignment.LocalAssignment, column, value string) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := a.DB.UpdateAssignment(LocalAssignment, column, value); err != nil {
		return err
	}

	assignment_id_int := int(LocalAssignment.RemoteID)

	assignment_id := strconv.Itoa(assignment_id_int)

	if err := client.SendUpdate(assignment_id, column, value); err != nil {
		return err
	}

	return nil
}

// DeleteAssignment deletes an assignment
func (a *App) DeleteAssignment(assignment *assignment.Assignment) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.DeleteAssignment(assignment)
}

// CreateCourse creates a new course
func (a *App) CreateCourse(course *course.Course) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.CreateCourse(course)
}

// UpdateCourse updates an existing course
func (a *App) UpdateCourse(course *course.Course) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.UpdateCourse(course)
}

// DeleteCourse deletes a course
func (a *App) DeleteCourse(course *course.Course) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.DeleteCourse(course)
}
