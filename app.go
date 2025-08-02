package main

import (
	"context"
	"encoding/json"
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
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"
	"unipilot/internal/models/user"
	"unipilot/internal/network"
	"unipilot/internal/services/fileops"
	"unipilot/internal/sse"
	"unipilot/internal/storage"

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

	fmt.Println("Creating assignment:", localAssignment)

	if !a.Auth.IsAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}
	fmt.Println("User ID:", a.DB.GetCurrentUserID())

	// Create the assignment within the transaction
	if err := tx.Create(localAssignment).Error; err != nil {
		tx.Rollback()
		return err
	}

	fmt.Println(" local assignment success ")
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

	responseAssignment, err := client.CreateAssignment(remoteAssignment)
	if err != nil {
		tx.Rollback()
		fmt.Println("Error creating remote assignment:", err)
		return err
	}

	tx.Commit()
	log.Printf("Response assignment: %v\n", responseAssignment)

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
		RoomNumber:      courseData.RoomNumber,
		Instructor:      courseData.Instructor,
		InstructorEmail: courseData.InstructorEmail,
		StartDate:       courseData.StartDate,
		EndDate:         courseData.EndDate,
	}

	fmt.Println("Creating course:", localCourse)

	if !a.Auth.IsAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}
	fmt.Println("User ID:", a.DB.GetCurrentUserID())

	// Create the assignment within the transaction
	if err := tx.Create(localCourse).Error; err != nil {
		tx.Rollback()
		return err
	}

	fmt.Println(" local assignment success ")
	remoteCourse := &course.Course{
		LocalID:         localCourse.ID,
		Name:            localCourse.Name,
		Code:            localCourse.Code,
		Color:           localCourse.Color,
		Semester:        localCourse.Semester,
		Schedule:        localCourse.Schedule,
		Credits:         localCourse.Credits,
		RoomNumber:      localCourse.RoomNumber,
		Instructor:      localCourse.Instructor,
		InstructorEmail: localCourse.InstructorEmail,
		StartDate:       localCourse.StartDate,
		EndDate:         localCourse.EndDate,
	}

	responseCourse, err := client.CreateCourse(remoteCourse)
	if err != nil {
		tx.Rollback()
		fmt.Println("Error creating remote course:", err)
		return err
	}

	tx.Commit()
	log.Printf("Response course: %v\n", responseCourse)

	return nil
}

// UploadDocument opens a file dialog and uploads a document to an assignment
func (a *App) UploadDocument(assignmentID uint, documentType string) (*document.LocalDocument, error) {
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
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := a.DB.UpdateAssignment(LocalAssignment, column, value); err != nil {
		return err
	}

	assignment_id_int := int(LocalAssignment.ID)

	assignment_id := strconv.Itoa(assignment_id_int)

	if err := client.SendAssignmentUpdate(assignment_id, column, value); err != nil {
		return err
	}

	return nil
}

// UpdateCourse updates an existing course
func (a *App) UpdateCourse(course *course.LocalCourse, column, value string) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	if err := a.DB.UpdateCourse(course, column, value); err != nil {
		return err
	}

	course_id_int := int(course.ID)

	course_id := strconv.Itoa(course_id_int)

	if err := client.SendCourseUpdate(course_id, column, value); err != nil {
		return err
	}

	return nil
}

// UploadNewDocumentVersion uploads a new version of an existing document
func (a *App) UploadNewDocumentVersion(existingDocumentID uint) (*document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
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
	if a.Auth.IsAuthenticated() && a.Auth.Client != nil {
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

// ========================================
// DELETE OPERATIONS
// ========================================

// DeleteAssignment deletes an assignment
func (a *App) DeleteAssignment(assignment *assignment.LocalAssignment) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

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

	if err := client.SendAssignmentUpdate(assignment_id_str, "deleted_at", time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	return nil
}

// DeleteCourse deletes a course
func (a *App) DeleteCourse(course *course.LocalCourse) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

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

	if err := client.SendCourseUpdate(course_id_str, "deleted_at", time.Now().Format(time.RFC3339)); err != nil {
		return err
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
	db = db.Debug()
	// Delete database record
	if err := db.Delete(&doc).Error; err != nil {
		return fmt.Errorf("failed to delete document record: %w", err)
	}

	// Update local storage cache
	document.UpdateLocalStorageCache(userID, db)

	// Also store metadata remotely for sharing
	if a.Auth.IsAuthenticated() && a.Auth.Client != nil {

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

// ========================================
// OTHER OPERATIONS
// ========================================

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

	// Check if user is already authenticated and initialize HTTP client + SSE if needed
	if a.Auth.IsAuthenticated() && network.IsOnline() {
		log.Println("[App] User already authenticated, initializing HTTP client and SSE connection...")
		if err := a.initializeAuthenticatedClient(); err != nil {
			log.Printf("[App] Failed to initialize authenticated client: %v", err)
		} else {
			a.startSSEConnection()
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

// startSSEConnection initializes and starts the SSE connection with proper authentication check
func (a *App) startSSEConnection() {
	// Ensure we have an authenticated HTTP client
	if a.Auth.Client == nil {
		log.Println("[App] No HTTP client available, cannot start SSE")
		return
	}

	// Stop any existing SSE connection first
	a.stopSSEConnection()

	// Initialize new SSE connection
	a.Auth.SSE = sse.NewSSE()

	// Start the SSE connection in a goroutine
	go a.Auth.SSE.Connect(a.Auth.Client)

	// Start the event handler
	a.Events.Start(a.Auth.SSE)

	log.Println("[App] SSE connection started successfully")
}

// stopSSEConnection stops the current SSE connection and event handling
func (a *App) stopSSEConnection() {
	if a.Auth.SSE != nil {
		log.Println("[App] Stopping existing SSE connection...")
		a.Auth.SSE.StopListener()
		a.Auth.SSE = nil
	}

	if a.Events != nil {
		a.Events.Stop()
		// Recreate events handler for next connection
		a.Events = events.NewEvents()
	}

	log.Println("[App] SSE connection stopped")
}

// ensureSSEConnection ensures SSE is running if user is authenticated and online
func (a *App) ensureSSEConnection() {
	if !a.Auth.IsAuthenticated() {
		log.Println("[App] User not authenticated, stopping SSE if running")
		a.stopSSEConnection()
		return
	}

	if !network.IsOnline() {
		log.Println("[App] Network offline, stopping SSE if running")
		a.stopSSEConnection()
		return
	}

	// If we don't have an HTTP client, try to initialize it
	if a.Auth.Client == nil {
		log.Println("[App] No HTTP client available, trying to initialize from stored cookies")
		if err := a.initializeAuthenticatedClient(); err != nil {
			log.Printf("[App] Failed to initialize HTTP client: %v", err)
			return
		}
	}

	// If we don't have an active SSE connection, start one
	if a.Auth.SSE == nil {
		log.Println("[App] No active SSE connection, starting new connection")
		a.startSSEConnection()
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

	// Start SSE connection after successful login
	if network.IsOnline() {
		a.startSSEConnection()
	}

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
	// Stop SSE connection first
	a.stopSSEConnection()

	if err := a.Auth.Logout(); err != nil {
		return err
	}

	a.DB = nil

	return nil
}

// IsAuthenticated checks if the user is currently authenticated
func (a *App) IsAuthenticated() (*storage.LocalCredentials, error) {
	creds, err := storage.GetCurrentUser()
	if err != nil {
		// When no credentials exist, return a LocalCredentials object with IsAuthenticated = false
		// instead of an error, so frontend can properly handle the unauthenticated state
		log.Printf("[App] No credentials found: %v", err)
		return &storage.LocalCredentials{
			IsAuthenticated: false,
			User: struct {
				UserID   uint   `json:"user_id"`
				Username string `json:"username"`
			}{
				UserID:   0,
				Username: "",
			},
		}, nil
	}
	if creds == nil {
		// Same handling for nil credentials
		return &storage.LocalCredentials{
			IsAuthenticated: false,
			User: struct {
				UserID   uint   `json:"user_id"`
				Username string `json:"username"`
			}{
				UserID:   0,
				Username: "",
			},
		}, nil
	}
	return creds, nil
}

// EnsureSSEConnection manually checks and ensures SSE connection is in correct state
func (a *App) EnsureSSEConnection() error {
	log.Println("[App] Manual SSE connection check requested")
	a.ensureSSEConnection()
	return nil
}

// GetSSEConnectionStatus returns the current status of the SSE connection
func (a *App) GetSSEConnectionStatus() map[string]interface{} {
	status := map[string]interface{}{
		"connected":     false,
		"authenticated": a.Auth.IsAuthenticated(),
		"online":        network.IsOnline(),
		"client_ready":  a.Auth.Client != nil,
	}

	if a.Auth.SSE != nil {
		status["connected"] = true
		status["sse_instance"] = "active"
	} else {
		status["sse_instance"] = "inactive"
	}

	return status
}

// ReconnectSSE manually triggers an SSE reconnection (useful for network recovery)
func (a *App) ReconnectSSE() error {
	if !a.Auth.IsAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}

	if !network.IsOnline() {
		return fmt.Errorf("network is offline")
	}

	if a.Auth.Client == nil {
		return fmt.Errorf("HTTP client not available")
	}

	log.Println("[App] Manual SSE reconnection requested")
	a.startSSEConnection()

	return nil
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

// GetAssignmentDocuments retrieves all documents for an assignment
func (a *App) GetAssignmentDocuments(assignmentID uint) ([]document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	// Use LocalDocument and only return documents we have locally
	var documents []document.LocalDocument
	err := a.DB.GetDB().Where(
		"assignment_id = ? AND user_id = ? AND has_local_file = ?",
		assignmentID, userID, true,
	).Order("created_at DESC").Find(&documents).Error

	return documents, err
}

// GetSupportDocuments retrieves only support documents for an assignment
func (a *App) GetSupportDocuments(assignmentID uint) ([]document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	var documents []document.LocalDocument
	err := a.DB.GetDB().Where(
		"assignment_id = ? AND user_id = ? AND type = ? AND has_local_file = ?",
		assignmentID, userID, document.DocumentTypeSupport, true,
	).Order("created_at DESC").Find(&documents).Error

	return documents, err
}

// GetSubmissionDocuments retrieves only submission documents for an assignment
func (a *App) GetSubmissionDocuments(assignmentID uint) ([]document.LocalDocument, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	var documents []document.LocalDocument
	err := a.DB.GetDB().Where(
		"assignment_id = ? AND user_id = ? AND type = ? AND has_local_file = ?",
		assignmentID, userID, document.DocumentTypeSubmission, true,
	).Order("created_at DESC").Find(&documents).Error

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

	// Get local document record
	var doc document.LocalDocument
	if err := a.DB.GetDB().Where("id = ? AND user_id = ?", documentID, userID).First(&doc).Error; err != nil {
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

	if !a.Auth.IsAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	// Get local document record
	var doc document.LocalDocument
	if err := a.DB.GetDB().Where("id = ? AND user_id = ?", documentID, userID).First(&doc).Error; err != nil {
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
func (a *App) GetUserStorageInfo() (*document.LocalDocumentCache, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	// Get or create local storage cache
	var cache document.LocalDocumentCache
	err := a.DB.GetDB().FirstOrCreate(&cache, document.LocalDocumentCache{UserID: userID}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get storage info: %w", err)
	}

	// Update cache if it's stale (older than 1 hour)
	if time.Since(cache.LastCalculatedAt) > time.Hour {
		document.UpdateLocalStorageCache(userID, a.DB.GetDB())
		// Reload updated cache
		a.DB.GetDB().First(&cache, "user_id = ?", userID)
	}

	return &cache, nil
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

func (a *App) GetCourseAssignments(course *course.LocalCourse) ([]assignment.LocalAssignment, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if !a.Auth.IsAuthenticated() {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID := a.DB.GetCurrentUserID()

	var assignments []assignment.LocalAssignment
	err := a.DB.GetDB().Where("course_code = ? AND user_id = ?", course.Code, userID).Find(&assignments).Order("created_at ASC").Error
	return assignments, err
}
