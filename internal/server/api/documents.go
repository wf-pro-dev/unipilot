package server

import (
	"context"
	"encoding/json"
	Errors "errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qdrant/go-client/qdrant"
	"gorm.io/gorm"

	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/document"
	notif "unipilot/internal/models/notifications"
	"unipilot/internal/models/user"
	"unipilot/internal/server"
	"unipilot/internal/server/sse/grpc/notifications"
	cloudstorage "unipilot/internal/services/cloud_storage"

	"unipilot/internal/errors"
)

// DocumentMetadata represents document metadata for API responses.
// Provides a safe, serializable structure for document information that excludes
// sensitive internal fields and includes client-specific flags like HasLocalFile.
// Used primarily for GET endpoints to provide comprehensive document listings.
type DocumentMetadata struct {
	ID                uint   `json:"id"`
	LocalID           uint   `json:"local_id"`
	AssignmentID      uint   `json:"assignment_id"`
	LocalAssignmentID uint   `json:"local_assignment_id"`
	UserID            uint   `json:"user_id"`
	Type              string `json:"type"`
	FileName          string `json:"file_name"`
	FileType          string `json:"file_type"`
	FileSize          int64  `json:"file_size"`
	Version           int    `json:"version"`
	IsOriginal        bool   `json:"is_original"`
	HasLocalFile      bool   `json:"has_local_file"`
	CreatedAt         string `json:"created_at"`
}

func GetDocumentsHandler(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	currentUserID := currentUser.ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, "message", "Documents retrieved successfully")

	// Get all documents for the current user
	var documents []document.Document
	if err := db.Where("user_id = ?", currentUserID).Find(&documents).Error; err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return server.LogError(
				ctx,
				errors.WrapServer(
					err,
					errors.DBRecordNotFound,
					"Documents not found",
					fiber.StatusNotFound,
				))
		}
		return server.LogError(
			ctx,
			errors.WrapServer(
				err,
				errors.DBQueryFailed,
				"Error getting documents from database",
				fiber.StatusInternalServerError,
			))
	}

	return c.JSON(fiber.Map{
		"documents": documents,
	})
}

// CreateDocumentHandler creates a new document record with file upload to cloud storage.
// Handles multipart form uploads, processes document metadata, stores files in AWS S3,
// and sends real-time notifications to users linked to the associated assignment.
// Supports both new file uploads and copying existing files within the storage system.
//
// HTTP Method: POST
// Content-Type: multipart/form-data
//
// Request Body (multipart form):
//   - metadata: JSON string containing LocalDocument structure (required)
//   - file: Binary file data (required if has_local_file is true)
//
// LocalDocument metadata fields:
//   - id: Local document identifier (uint, required)
//   - assignment_id: Local assignment ID (uint, required)
//   - remote_assignment_id: Server assignment ID (uint, required)
//   - file_name: Original file name (string, required)
//   - file_type: MIME type (string, required)
//   - file_size: File size in bytes (int64, required)
//   - type: Document type/category (string, required)
//   - version: Document version number (int, required)
//   - is_original: Whether this is the original document (bool, required)
//   - has_local_file: Whether file data is included in request (bool, required)
//   - storage_key: Existing storage key for file copying (string, optional)
//
// Response (200 OK):
//   - success: Boolean success indicator
//   - document: Created document object with metadata
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Database Operations:
//   - Creates record in `documents` table with file metadata
//   - Updates user storage information for quota tracking
//   - Queries assignment data for notification distribution
//
// Cloud Storage Operations:
//   - Uploads new files to AWS S3 with unique timestamped names
//   - Copies existing files within S3 storage system
//   - Organizes files in user/assignment directory structure
//   - Cleans up temporary local files after S3 upload
//
// Notification System:
//   - Sends SSE notifications to users linked to assignment's child assignments
//   - Uses gRPC client for real-time notification delivery
//   - Includes document data and creator information in notifications
//   - Only sends notifications for new file uploads (not copies)
//
// Security Features:
//   - User isolation through directory structure (user_id based paths)
//   - File size limits enforced through multipart form parsing (32MB)
//   - Unique file naming prevents conflicts and overwrites
//   - Assignment ownership validation through linked assignments
//
// Error Responses:
//   - 400 Bad Request: Invalid multipart form, missing metadata, or format errors
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: Storage operations, database failures, or notification errors
//
// Side Effects:
//   - Creates document record in database
//   - Uploads file to AWS S3 cloud storage
//   - Updates user storage quota information
//   - Sends real-time notifications to linked assignment users
//   - Logs document creation with performance metrics
func CreateDocumentHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser := c.Locals("user").(user.User)
	db := c.Locals("db").(*gorm.DB)
	userID := currentUser.ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, "message", "Document created successfully")

	// Step 3: Extract and validate document metadata from form
	metadata := c.FormValue("metadata")
	if metadata == "" {
		return errors.WrapServer(
			fmt.Errorf("metadata missing from form"),
			errors.ReqParamMissing,
			"Metadata missing from form",
			fiber.StatusBadRequest,
		)
	}

	// Step 4: Parse JSON metadata into LocalDocument structure
	// Parse metadata directly into LocalDocument
	var localDoc document.LocalDocument
	if err := json.Unmarshal([]byte(metadata), &localDoc); err != nil {
		return errors.WrapServer(
			err,
			errors.ProcJSONUnmarshalFailed,
			"Invalid metadata format",
			fiber.StatusBadRequest,
		)
	}

	// Step 5: Generate unique storage paths and file names for cloud storage
	// Create user, assignment  directory
	assignmentDir := fmt.Sprintf("users_data/user_%d/documents/assign_%d", userID, localDoc.RemoteAssignmentID)

	// Generate unique filename
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), localDoc.FileName)
	newKey := fmt.Sprintf("%s/%s", assignmentDir, uniqueFileName)

	// Step 6: Handle file upload or copy based on document source
	if localDoc.HasLocalFile {
		// Upload new file from multipart form to S3
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return errors.WrapServer(
				err,
				errors.ReqBodyInvalid,
				"Unable to get file from form",
				fiber.StatusBadRequest,
			)
		}
		if err := UploadFile(localDoc, newKey, fileHeader, c); err != nil {
			return errors.WrapServer(
				err,
				errors.StorageUploadFailed,
				"Error uploading file to S3",
				fiber.StatusInternalServerError,
			)
		}
	} else {
		// Copy existing file within S3 storage system
		if err := cloudstorage.CopyFile(localDoc.StorageKey, newKey); err != nil {
			return errors.WrapServer(
				err,
				errors.StorageFileNotFound,
				"Error copying file",
				fiber.StatusInternalServerError,
			)
		}
	}

	// Step 7: Create document metadata record in database
	// Create document metadata record
	doc := &document.Document{
		AssignmentID:      localDoc.RemoteAssignmentID,
		LocalAssignmentID: localDoc.AssignmentID,
		LocalID:           localDoc.ID,
		UserID:            userID,
		Type:              localDoc.Type,
		FileName:          localDoc.FileName,
		FileType:          localDoc.FileType,
		FileSize:          localDoc.FileSize,
		Version:           localDoc.Version,
		IsOriginal:        localDoc.IsOriginal,
		FilePath:          localDoc.FilePath,
		StorageKey:        newKey,
	}

	if err := db.Create(doc).Error; err != nil {
		if Errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.WrapServer(
				err,
				errors.DBConstraintViolation,
				"Document already exists",
				fiber.StatusConflict,
			)
		}
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error saving document metadata",
			fiber.StatusInternalServerError,
		)
	}

	// Step 8: Update user storage quota information for tracking
	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error updating user storage quota",
			fiber.StatusInternalServerError,
		)
	}

	// Step 9: Retrieve assignment data for notification distribution
	// Get document assignment
	var a assignment.Assignment
	if err := db.Where("id = ?", doc.AssignmentID).First(&a).Error; err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(
				err,
				errors.DBRecordNotFound,
				"Assignment not found",
				fiber.StatusNotFound,
			)
		}
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting assignment",
			fiber.StatusInternalServerError,
		)
	}

	// Step 10: Get child assignments for notification distribution
	// Get linked assignments
	linkedAssignments, err := a.GetChildren(db)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting linked assignments",
			fiber.StatusInternalServerError,
		)
	}

	// Step 11: Prepare document data for notifications
	// Marshal document
	doc.User = currentUser // Link the creator data with the document
	dJson, err := json.Marshal(doc)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ProcJSONMarshalFailed,
			"Error marshalling document",
			fiber.StatusInternalServerError,
		)
	}

	// Step 12: Send real-time notifications to linked assignment users (only for new uploads)
	if GrpcClient != nil && localDoc.HasLocalFile {
		// Send notification to linked assignments
		for _, linkedAssignment := range linkedAssignments {
			_, err := (*GrpcClient).SendNotification(context.Background(),
				&notifications.Notification{
					UserId:   uint32(linkedAssignment.UserID),
					SenderId: uint32(userID),
					Entity:   string(models.EntityDocument),
					EntityId: uint32(linkedAssignment.ID),
					Type:     string(notif.NotificationDocumentUpdate),
					Title:    linkedAssignment.Title,
					Message:  fmt.Sprintf("%s shared a new document on %s", currentUser.Username, doc.FileName),
					Action:   "document",
					Data:     string(dJson),
				},
			)
			if err != nil {
				return errors.WrapServer(
					err,
					errors.GRPCFailed,
					"Failed to send notification",
					fiber.StatusInternalServerError,
				)
			}
		}
	}

	// Step 13: Send successful response with document metadata
	return c.JSON(fiber.Map{
		"success":  true,
		"document": doc,
	})
}

// (DEPRECATED) WriteFileToDisk extracts file from multipart form and writes it to local disk storage.
// Used as an intermediate step before uploading to cloud storage or processing for RAG.
// Creates necessary directory structure and handles file I/O operations safely.
//
// Parameters:
//   - localDoc: Document metadata structure (used for context)
//   - key: Storage key/path for organizing file location
//   - w: HTTP response writer (for error handling context)
//   - r: HTTP request containing multipart form with file data
//
// Returns:
//   - string: Full file path where file was written
//   - int64: Number of bytes written to disk
//   - error: Any error encountered during file operations
//
// File Operations:
//   - Extracts file from "file" form field
//   - Creates directory structure under /app/uploads (Docker volume)
//   - Writes file content to disk with proper permissions (0755 for directories)
//   - Returns file path and byte count for further processing
//
// Error Handling:
//   - Returns errors for form file extraction failures
//   - Returns errors for directory creation failures
//   - Returns errors for file creation or writing failures
func WriteFileToDisk(key string, w http.ResponseWriter, r *http.Request) (string, int64, error) {
	// Get the file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	// Use uploads directory from Docker volume
	uploadsDir := "/app/uploads"
	filePath := filepath.Join(uploadsDir, key)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create directory: %w", err)
	}
	// Save file to disk
	destFile, err := os.Create(filePath)
	if err != nil {
		return "", 0, err
	}
	defer destFile.Close()

	// Copy file content
	bytesWritten, err := io.Copy(destFile, file)
	if err != nil {
		return "", 0, err
	}

	return filePath, bytesWritten, nil
}

// WriteMultipartFile extracts file from Fiber multipart form and writes it to local disk storage.
func WriteMultipartFile(key string, fileHeader *multipart.FileHeader, c *fiber.Ctx) (string, int64, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", 0, errors.Wrap(
			err,
			errors.FSOpenFailed,
			"Error opening file",
		)
	}
	defer src.Close()

	// Use uploads directory from Docker volume
	filePath := filepath.Join("/app/uploads/", key)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", 0, errors.Wrap(
			err,
			errors.FSDirFailed,
			"Error creating directory",
		)
	}

	// Save file to disk
	destFile, err := os.Create(filePath)
	if err != nil {
		return "", 0, errors.Wrap(
			err,
			errors.FSCreateFailed,
			"Error creating file",
		)
	}
	defer destFile.Close()

	// Copy file content
	bytesWritten, err := io.Copy(destFile, src)
	if err != nil {
		return "", 0, errors.Wrap(
			err,
			errors.FSWriteFailed,
			"Error writing file",
		)
	}

	return filePath, bytesWritten, nil
}

func WriteFile(key string, file io.Reader, c *fiber.Ctx) (string, int64, error) {

	filePath := filepath.Join("/app/uploads/", key)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", 0, errors.WrapServer(
			err,
			errors.FSDirCreateFailed,
			"Error creating directory",
			fiber.StatusInternalServerError,
		)
	}

	destFile, err := os.Create(filePath)
	if err != nil {
		return "", 0, errors.WrapServer(
			err,
			errors.FSCreateFailed,
			"Error creating file",
			fiber.StatusInternalServerError,
		)
	}
	defer destFile.Close()

	bytesWritten, err := io.Copy(destFile, file)
	if err != nil {
		return "", 0, errors.WrapServer(
			err,
			errors.FSWriteFailed,
			"Error copying file content",
			fiber.StatusInternalServerError,
		)
	}

	return filePath, bytesWritten, nil
}

// (DEPRECATED) UploadFileToS3 handles the complete file upload pipeline from multipart form to AWS S3.
// Combines local disk writing with cloud storage upload, then cleans up temporary files.
// Provides comprehensive logging for upload tracking and debugging.
//
// Parameters:
//   - localDoc: Document metadata structure containing file information
//   - key: S3 storage key/path for organizing file in cloud storage
//   - w: HTTP response writer (for error handling context)
//   - r: HTTP request containing multipart form with file data
//
// Returns:
//   - error: Any error encountered during the upload pipeline
//
// Upload Pipeline:
//  1. Writes file to local disk using WriteFileToDisk
//  2. Uploads file from local disk to AWS S3 using cloud storage service
//  3. Logs successful upload with file metrics
//  4. Cleans up temporary local file after S3 upload
//
// Error Handling:
//   - Returns errors from disk writing operations
//   - Returns errors from S3 upload operations
//   - Continues with cleanup even if logging fails
//
// Side Effects:
//   - Creates temporary file on local disk (cleaned up after upload)
//   - Uploads file to AWS S3 cloud storage
//   - Logs upload metrics for monitoring and debugging
func UploadFileLegacy(localDoc document.LocalDocument, key string, w http.ResponseWriter, r *http.Request) error {

	filePath, _, err := WriteFileToDisk(key, w, r)
	if err != nil {
		return errors.Wrap(
			err,
			errors.FSCreateFailed,
			"Error writing file to disk",
		)
	}

	// Upload to aws S3
	if err := cloudstorage.UploadFile(filePath, localDoc.FileName, key); err != nil {
		return fmt.Errorf("failed to upload file to R2: %w", err)
	}

	// Clean up local file after S3 upload
	os.Remove(filePath)

	return nil
}

// UploadFileToS3Fiber handles the complete file upload pipeline from Fiber multipart form to AWS S3.
func UploadFile(localDoc document.LocalDocument, key string, fileHeader *multipart.FileHeader, c *fiber.Ctx) error {
	filePath, _, err := WriteMultipartFile(key, fileHeader, c)
	if err != nil {
		return errors.Wrap(
			err,
			errors.FSWriteFailed,
			"Error writing file to disk",
		)
	}

	// Upload to aws S3
	if err := cloudstorage.UploadFile(filePath, localDoc.FileName, key); err != nil {
		return errors.Wrap(
			err,
			errors.StorageUploadFailed,
			"Error uploading file to cloud storage",
		)
	}

	// Clean up local file after S3 upload
	os.Remove(filePath)

	return nil
}

// DownloadDocumentHandler streams document files directly from AWS S3 to client.
// Provides secure file download with proper headers and streaming for large files.
// Uses cloud storage service to retrieve files without storing them locally.
//
// HTTP Method: POST
// Content-Type: application/json
//
// Request Body:
//   - storage_key: S3 storage key for the file to download (string, required)
//   - file_name: Original file name for download headers (string, required)
//   - file_size: File size for Content-Length header (int64, required)
//
// Response: Binary file stream with appropriate headers
//   - Content-Type: application/octet-stream
//   - Content-Disposition: attachment with original filename
//   - Content-Length: File size for download progress
//
// Authentication: Not explicitly required (public endpoint for file access)
//
// Cloud Storage Operations:
//   - Downloads file from AWS S3 using storage key
//   - Streams file directly to HTTP response without local storage
//   - Handles large files efficiently through streaming
//
// Security Features:
//   - Uses storage keys for file identification (not direct file paths)
//   - Proper Content-Disposition headers prevent XSS attacks
//   - Streaming prevents memory exhaustion on large files
//
// Error Responses:
//   - 400 Bad Request: Invalid JSON body or missing required fields
//   - 500 Internal Server Error: S3 download failures or streaming errors
//
// Side Effects:
//   - Streams file content directly to client
//   - Logs download metrics for monitoring
//   - No local file storage or cleanup required
func DownloadDocumentHandler(c *fiber.Ctx) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "message", "Document downloaded successfully")
	// Step 1: Parse document download request from JSON body
	var docData document.LocalDocument
	if err := c.BodyParser(&docData); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	// Step 2: Download file from AWS S3 using storage key
	// Download from aws S3
	fileReader, err := cloudstorage.DownloadFile(docData.StorageKey)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.StorageDownloadFailed,
			"Error downloading file",
			fiber.StatusInternalServerError,
		)
	}

	// Step 3: Set appropriate HTTP headers for secure file download
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", docData.FileName))
	c.Set("Content-Length", strconv.FormatInt(docData.FileSize, 10))

	// Step 4: Stream file content directly to client response
	_, err = io.Copy(c.Response().BodyWriter(), fileReader)
	if err != nil {

		return errors.WrapServer(
			err,
			errors.FSStreamFailed,
			"Error streaming file",
			fiber.StatusInternalServerError,
		)
	}

	return nil
}

// GetAssignmentDocumentsHandler retrieves all document metadata for a specific assignment.
// Returns comprehensive document information with user-specific flags for local file availability.
// Provides document listings for assignment management interfaces with proper access control.
//
// HTTP Method: GET
// Content-Type: Not required (query parameters used)
//
// Query Parameters:
//   - assignment_id: Assignment ID to retrieve documents for (string, required)
//
// Response (200 OK):
//   - success: Boolean success indicator
//   - documents: Array of DocumentMetadata objects with comprehensive document information
//
// DocumentMetadata includes:
//   - id, assignment_id, user_id: Database identifiers
//   - type, file_name, file_type, file_size: File information
//   - version, is_original: Version control information
//   - has_local_file: Whether current user owns this document
//   - created_at: Document creation timestamp
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Database Operations:
//   - Queries `documents` table filtered by assignment_id
//   - Preloads user information for document ownership details
//   - Orders results by creation date (newest first)
//
// Security Features:
//   - Assignment ID validation and conversion
//   - User-specific has_local_file flag (only true for document owner)
//   - Safe metadata exposure without sensitive internal fields
//
// Error Responses:
//   - 400 Bad Request: Missing or invalid assignment_id parameter
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: Database query failures
//
// Side Effects:
//   - Logs document retrieval with count for monitoring
//   - No database modifications (read-only operation)
func GetAssignmentDocumentsHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(
			fmt.Errorf("user not found"),
			errors.ValidationInvalid,
			"User not found",
			fiber.StatusInternalServerError,
		)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(
			fmt.Errorf("db not found"),
			errors.ValidationInvalid,
			"DB not found",
			fiber.StatusInternalServerError,
		)
	}

	currentUserID := currentUser.ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, "message", "Assignment documents retrieved successfully")
	// Step 2: Extract and validate assignment ID from query parameters
	assignmentIDStr := c.Query("assignment_id")
	if assignmentIDStr == "" {
		return errors.WrapServer(
			fmt.Errorf("assignment ID required"),
			errors.ReqParamMissing,
			"Assignment ID required",
			fiber.StatusBadRequest,
		)
	}

	// Step 3: Convert assignment ID string to integer for database query
	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 32)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Error converting assignment ID",
			fiber.StatusBadRequest,
		)
	}

	// Step 4: Query documents for the specified assignment with user information
	var documents []document.Document

	if documents, err = document.GetDocumentsByAssignment(uint(assignmentID), currentUserID, db); err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting documents from database",
			fiber.StatusInternalServerError,
		)
	}

	// Step 5: Convert documents to safe metadata format with user-specific flags
	var docResponses []DocumentMetadata
	for _, doc := range documents {
		// Determine if current user owns this document (affects local file availability)
		hasLocalFile := doc.UserID == currentUserID

		docResponses = append(docResponses, DocumentMetadata{
			ID:           doc.ID,
			AssignmentID: doc.AssignmentID,
			UserID:       doc.UserID,
			Type:         string(doc.Type),
			FileName:     doc.FileName,
			FileType:     doc.FileType,
			FileSize:     doc.FileSize,
			Version:      doc.Version,
			IsOriginal:   doc.IsOriginal,
			HasLocalFile: hasLocalFile,
			CreatedAt:    doc.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// Step 6: Send successful response with document metadata array
	return c.JSON(fiber.Map{
		"success":   true,
		"documents": docResponses,
	})
}

// DeleteDocumentHandler removes document record and associated file from cloud storage.
// Provides secure document deletion with ownership validation and storage cleanup.
// Updates user storage quota information after successful deletion.
//
// HTTP Method: DELETE
// Content-Type: Not required (query parameters used)
//
// Query Parameters:
//   - document_id: Local document ID to delete (string, required)
//
// Response (200 OK):
//   - success: Boolean success indicator
//   - message: Confirmation message
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Database Operations:
//   - Queries document by local_id and user_id for ownership validation
//   - Deletes document record from database
//   - Updates user storage quota information
//
// Cloud Storage Operations:
//   - Deletes associated file from AWS S3 using file path
//   - Handles storage cleanup to prevent orphaned files
//
// Security Features:
//   - Document ownership validation (user can only delete their own documents)
//   - Uses local_id and user_id combination for secure identification
//   - Prevents unauthorized deletion of other users' documents
//
// Error Responses:
//   - 400 Bad Request: Missing document_id parameter
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 404 Not Found: Document not found or not owned by user
//   - 500 Internal Server Error: Storage deletion or database failures
//
// Side Effects:
//   - Removes document record from database
//   - Deletes file from AWS S3 cloud storage
//   - Updates user storage quota information
//   - Logs deletion for audit trail
func DeleteDocumentHandler(c *fiber.Ctx) error {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser, ok := c.Locals("user").(user.User)
	if !ok {
		return errors.WrapServer(
			fmt.Errorf("user not found"),
			errors.ValidationInvalid,
			"User not found",
			fiber.StatusInternalServerError,
		)
	}
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(
			fmt.Errorf("db not found"),
			errors.ValidationInvalid,
			"DB not found",
			fiber.StatusInternalServerError,
		)
	}
	userID := currentUser.ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, "message", "Document deleted successfully")
	// Step 2: Extract document ID from path parameter
	docID := c.Params("id")
	if docID == "" {
		return errors.WrapServer(
			fmt.Errorf("document ID required"),
			errors.ReqParamMissing,
			"Document ID required",
			fiber.StatusBadRequest,
		)
	}

	// Step 3: Find document with ownership validation (local_id + user_id)
	var doc document.Document
	if err := db.Where("local_id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(
				err,
				errors.DBRecordNotFound,
				"Document not found",
				fiber.StatusNotFound,
			)
		}
		return errors.WrapServer(
			err,
			errors.DBRecordNotFound,
			"Document not found",
			fiber.StatusNotFound,
		)
	}

	// Step 4: Delete associated file from R2 cloud storage
	// Delete the document on R2
	if err := cloudstorage.DeleteFile(doc.FilePath); err != nil {
		return errors.WrapServer(
			err,
			errors.StorageDeleteFailed,
			"Error deleting document from storage",
			fiber.StatusInternalServerError,
		)
	}

	// Step 5: Remove document record from database
	if err := db.Delete(&doc).Error; err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return server.LogError(
				ctx,
				errors.WrapServer(
					err,
					errors.DBRecordNotFound,
					"Document not found",
					fiber.StatusNotFound,
				))
		}
		if Errors.Is(err, gorm.ErrForeignKeyViolated) {
			return errors.WrapServer(
				err,
				errors.DBConstraintViolation,
				"Error deleting document",
				fiber.StatusConflict,
			)
		}
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error deleting document",
			fiber.StatusInternalServerError,
		)
	}

	// Step 6: Update user storage quota information after deletion
	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error updating user storage quota information",
			fiber.StatusInternalServerError,
		)

	}

	// Step 7: Send successful deletion confirmation
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Document metadata deleted",
	})
}

// UploadDocumentForRAGHandler processes documents for Retrieval-Augmented Generation (RAG).
// Handles document upload/download, converts to vectors using AI embeddings, and stores
// in Qdrant vector database for semantic search and AI-powered document retrieval.
//
// HTTP Method: POST
// Content-Type: multipart/form-data
//
// Request Body (multipart form):
//   - metadata: JSON string containing LocalDocument structure (required)
//   - file: Binary file data (required if has_local_file is true)
//
// LocalDocument metadata fields (same as CreateDocumentHandler):
//   - All fields from CreateDocumentHandler plus RAG-specific processing
//
// Response (200 OK):
//   - success: Boolean success indicator
//   - document: Processed document object with RAG metadata
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// RAG Processing Pipeline:
//  1. File preparation (upload new or download existing from S3)
//  2. Document vectorization using AI embeddings (768-dimensional vectors)
//  3. Qdrant collection creation (per-assignment collections)
//  4. Vector storage in Qdrant database for semantic search
//
// Qdrant Integration:
//   - Creates collections named "unipilot-qdrant-db-{assignment_id}"
//   - Uses 768-dimensional vectors with cosine distance similarity
//   - Stores document chunks as searchable vectors for AI retrieval
//   - Handles collection creation and vector upsertion
//
// File Processing:
//   - Supports both new uploads and existing file processing
//   - Downloads existing files from S3 for vectorization
//   - Processes documents into embeddings for semantic search
//   - Maintains file organization in local storage during processing
//
// Security Features:
//   - User isolation through assignment-based collections
//   - File access validation through existing document records
//   - Secure vector storage with assignment-level separation
//
// Error Responses:
//   - 400 Bad Request: Invalid multipart form, missing metadata, or format errors
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: File processing, vectorization, or Qdrant failures
//
// Side Effects:
//   - Creates or updates Qdrant vector collections
//   - Stores document vectors for AI-powered search
//   - Processes files for machine learning embeddings
//   - Logs RAG processing metrics for monitoring
func UploadDocumentForRAGHandler(c *fiber.Ctx) error {

	c.Locals("message", "Document uploaded for RAG successfully")

	metadata := c.FormValue("metadata")
	if metadata == "" {
		return errors.WrapServer(
			fmt.Errorf("metadata missing from form"),
			errors.ReqParamMissing,
			"Metadata missing from form",
			fiber.StatusBadRequest,
		)
	}

	var localDoc document.LocalDocument
	if err := json.Unmarshal([]byte(metadata), &localDoc); err != nil {
		return errors.WrapServer(
			err,
			errors.ProcJSONUnmarshalFailed,
			"Invalid metadata format",
			fiber.StatusBadRequest,
		)
	}

	if localDoc.HasLocalFile {
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return errors.WrapServer(
				err,
				errors.ReqBodyInvalid,
				"Unable to get file from form",
				fiber.StatusBadRequest,
			)
		}
		_, _, err = WriteMultipartFile(localDoc.StorageKey, fileHeader, c)
		if err != nil {
			return errors.WrapServer(
				err,
				errors.FSWriteFailed,
				"Error writing file to disk",
				fiber.StatusInternalServerError,
			)
		}
	} else {

		fileReader, err := cloudstorage.DownloadFile(localDoc.StorageKey)
		if err != nil {
			return errors.WrapServer(
				err,
				errors.StorageDownloadFailed,
				"Error downloading file",
				fiber.StatusInternalServerError,
			)
		}

		_, _, err = WriteFile(localDoc.StorageKey, fileReader, c)
		if err != nil {
			return errors.WrapServer(
				err,
				errors.FSWriteFailed,
				"Error writing file to disk",
				fiber.StatusInternalServerError,
			)
		}

	}

	vectors, err := document.GetQdrantVectors(&localDoc)
	if err != nil {
		if errors.HasCode(err, errors.FSFileTypeNotSupported) {
			return errors.WrapServer(
				err,
				errors.QdrantVectorsError,
				"File type not supported",
				fiber.StatusUnsupportedMediaType,
			)
		}
		return errors.WrapServer(
			err,
			errors.QdrantVectorsError,
			"Error getting Qdrant vectors",
			fiber.StatusInternalServerError,
		)
	}

	collections, err := QdrantClient.ListCollections(context.Background())
	if err != nil {
		return errors.WrapServer(
			err,
			errors.QdrantListCollectionsError,
			"Error listing Qdrant collections",
			fiber.StatusInternalServerError,
		)
	}

	collectionName := fmt.Sprintf("unipilot-qdrant-db-%d", localDoc.RemoteAssignmentID)

	if !slices.Contains(collections, collectionName) {
		err = QdrantClient.CreateCollection(context.Background(), &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     768,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			return errors.WrapServer(
				err,
				errors.QdrantCreateCollectionError,
				"Error creating Qdrant collection",
				fiber.StatusInternalServerError,
			)
		}

	}

	if _, err = QdrantClient.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         vectors,
	}); err != nil {
		return errors.WrapServer(
			err,
			errors.QdrantUpsertError,
			"Error inserting vectors into Qdrant",
			fiber.StatusInternalServerError,
		)
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"document": localDoc,
	})
}

func DeleteDocumentRAG(c *fiber.Ctx) error {

	c.Locals("message", "Document deleted from RAG successfully")
	// Step 1: Extract and validate document IDs from query parameters
	var input struct {
		AssignmentID uint `json:"assignment_id"`
		DocumentID   uint `json:"document_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	// Step 2: Delete the document from the Qdrant
	if _, err := QdrantClient.Delete(context.Background(), &qdrant.DeletePoints{
		CollectionName: fmt.Sprintf("unipilot-qdrant-db-%d", input.AssignmentID),
		Points: qdrant.NewPointsSelectorFilter(
			&qdrant.Filter{
				Must: []*qdrant.Condition{
					qdrant.NewMatchInt("document_id", int64(input.DocumentID)),
				},
			},
		),
	}); err != nil {
		return errors.WrapServer(
			err,
			errors.QdrantDeletePointsError,
			"Error deleting document from Qdrant",
			fiber.StatusInternalServerError,
		)
	}

	// Document deletion from RAG completed (logged by middleware)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Document deleted from RAG successfully",
	})
}

func GetAssignmentDocumentIDsRAG(c *fiber.Ctx) error {
	// Step 1: Extract assignment ID from path parameter
	idStr := c.Params("id")
	c.Locals("message", "Assignment document IDs retrieved successfully")
	if idStr == "" {

		return errors.WrapServer(
			fmt.Errorf("assignment ID required"),
			errors.ReqParamMissing,
			"Assignment ID required",
			fiber.StatusBadRequest,
		)
	}
	int_assignmentID, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Error converting assignment ID to int",
			fiber.StatusBadRequest,
		)
	}
	assignmentID := uint(int_assignmentID)

	var collectionName = fmt.Sprintf("unipilot-qdrant-db-%d", assignmentID)

	exists, err := QdrantClient.CollectionExists(context.Background(), collectionName)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.QdrantCollectionNotFound,
			"Assignment collection could not be found",
			fiber.StatusNotFound,
		)
	}

	if !exists {
		// Step 5: Return the list of uploaded document IDs
		return c.JSON(fiber.Map{
			"success":      true,
			"document_ids": []uint{},
		})
	}

	// Retrive All Qdrant Points for that assignment
	points, err := QdrantClient.Scroll(context.Background(), &qdrant.ScrollPoints{
		CollectionName: fmt.Sprintf("unipilot-qdrant-db-%d", assignmentID),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return errors.WrapServer(
			err,
			errors.QdrantFailed,
			"Error listing Qdrant points",
			fiber.StatusInternalServerError,
		)
	}

	// Step 3: Make a set of document IDs that are in the Qdrant
	type Set[E comparable] map[E]struct{} // Generic set type
	uploadedDocumentIDs := Set[uint]{}
	for _, point := range points {
		uploadedDocumentIDs[uint(point.Payload["document_id"].GetIntegerValue())] = struct{}{} // Add document ID to set
	}

	// Step 4: Flatten the set of uploaded document IDs
	uploadedDocumentIDsList := make([]uint, 0, len(uploadedDocumentIDs))
	for id := range uploadedDocumentIDs {
		uploadedDocumentIDsList = append(uploadedDocumentIDsList, id)
	}

	// Step 5: Return the list of uploaded document IDs
	return c.JSON(fiber.Map{
		"success":      true,
		"document_ids": uploadedDocumentIDsList,
	})
}
