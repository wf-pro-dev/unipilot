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
	"unipilot/internal/server"
	"unipilot/internal/server/sse/grpc/messages"
	cloudstorage "unipilot/internal/services/cloud_storage"
	"unipilot/internal/services/fileops"
	"unipilot/internal/services/fileops/progress"

	"unipilot/internal/errors"
)

func GetDocumentsHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}
	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}
	currentUserID := currentUser.ID

	// Get all documents for the current user
	var documents []models.Document
	if err := db.Where("user_id = ?", currentUserID).Find(&documents).Error; err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(
				err,
				errors.DBRecordNotFound,
				"Documents not found",
				fiber.StatusNotFound,
			)
		}
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting documents from database",
			fiber.StatusInternalServerError,
		)
	}

	return c.JSON(documents)
}

// CreateDocumentHandler creates a new document record with file upload to cloud storage.
// Handles multipart form uploads, processes document metadata, stores files in AWS S3,
// and sends real-time notifications to users linked to the associated models.
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
	var localDoc models.LocalDocument
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
	assignmentDir := fmt.Sprintf("users/%s/assignments/%s/documents", userID, localDoc.AssignmentID)

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
	} else if localDoc.StorageKey != nil {
		// Copy existing file within S3 storage system
		if err := cloudstorage.CopyFile(*localDoc.StorageKey, newKey); err != nil {
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
	doc := localDoc.ToRemote(userID)
	doc.StorageKey = &newKey

	if err := doc.Validate(db); err != nil {
		return errors.WrapServer(
			err,
			errors.ValidationInvalid,
			"Error validating document",
			fiber.StatusBadRequest,
		)
	}

	if err := db.Preload("Assignment.Course").Create(doc).First(doc).Error; err != nil {
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
	if err := models.UpdateStorageInfo(userID, db); err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error updating user storage quota",
			fiber.StatusInternalServerError,
		)
	}

	go func() {
		if GrpcClient != nil && doc.Assignment.Course.IsInCluster(db) && !doc.Assignment.IsCopy() {
			clusterRootID := doc.Assignment.ClusterRoot()
			users_course, err := CacheService.GetClusterUsers(context.Background(), clusterRootID, db)
			if err != nil {
				server.LogWarn(context.Background(), errors.WrapServer(err, errors.DBQueryFailed, "Error getting users linked to course", fiber.StatusInternalServerError))
				return
			}

			for _, sendeeID := range users_course {

				_, err := (*GrpcClient).SendMessage(context.Background(),
					&messages.Message{
						ReceiverId: sendeeID,
						SenderId:   userID,
						Title:      doc.Assignment.Title,
						Message:    fmt.Sprintf("%s shared a new document for %s", currentUser.Username, doc.Assignment.Course.Code),
						Data:       []byte(""),
						Type:       string(models.MessageNoContent),
					},
				)
				if err != nil {
					server.LogWarn(context.Background(), errors.WrapServer(err, errors.GRPCFailed, "Failed to send notification", fiber.StatusInternalServerError))
				}

			}
		}
	}()

	// Step 13: Send successful response with document metadata
	return c.JSON(fiber.Map{
		"remote_id":            doc.ID,
		"remote_assignment_id": doc.AssignmentID,
		"storage_key":          doc.StorageKey,
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

	server.LogDebug(c.Context(), "File written to disk", "filePath", filePath, "bytesWritten", bytesWritten)

	return filePath, bytesWritten, nil
}

// UploadFileToS3Fiber handles the complete file upload pipeline from Fiber multipart form to AWS S3.
func UploadFile(localDoc models.LocalDocument, key string, fileHeader *multipart.FileHeader, c *fiber.Ctx) error {
	filePath, _, err := WriteMultipartFile(key, fileHeader, c)
	if err != nil {
		return errors.Wrap(
			err,
			errors.FSWriteFailed,
			"Error writing file to disk",
		)
	}
	defer os.Remove(filePath)

	documentID := localDoc.ID
	progressManager := progress.GetManager()

	progressTracker := progressManager.Create(documentID, localDoc.FileSize)
	progressTracker.SetStatus("Uploading file to cloud storage")
	snapshot := progressTracker.Snapshot()

	CacheService.PublishProgress(c.Context(), documentID, &snapshot)

	progressTracker.OnProgress(func(t *progress.Tracker) {
		snapshot := t.Snapshot()
		// Calculate progress pe
		progress := float64(snapshot.Current) / float64(snapshot.Total) * 100
		snapshot.Percentage = progress
		// Store in cache
		CacheService.PublishProgress(c.Context(), documentID, &snapshot)
	})

	// Upload to aws S3
	if err := cloudstorage.UploadFile(filePath, localDoc.FileName, key, progressTracker); err != nil {
		return errors.Wrap(
			err,
			errors.StorageUploadFailed,
			"Error uploading file to cloud storage",
		)
	}

	// Complete upload
	progressTracker.Complete()
	snapshot = progressTracker.Snapshot()
	CacheService.PublishProgress(c.Context(), documentID, &snapshot)

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

	// Step 1: Parse document download request from JSON body
	var localDoc models.LocalDocument
	if err := c.BodyParser(&localDoc); err != nil {
		return errors.WrapServer(
			err,
			errors.ReqBodyInvalid,
			"Invalid request body",
			fiber.StatusBadRequest,
		)
	}

	documentID := localDoc.ID
	progressManager := progress.GetManager()

	progressTracker := progressManager.Create(documentID, localDoc.FileSize)
	progressTracker.SetStatus("Downloading file from cloud storage")
	snapshot := progressTracker.Snapshot()

	CacheService.PublishProgress(c.Context(), documentID, &snapshot)

	progressTracker.OnProgress(func(t *progress.Tracker) {
		snapshot := t.Snapshot()
		// Calculate progress pe
		progress := float64(snapshot.Current) / float64(snapshot.Total) * 100
		snapshot.Percentage = progress
		// Store in cache
		CacheService.PublishProgress(c.Context(), documentID, &snapshot)
	})

	// Step 2: Download file from AWS S3 using storage key
	// Download from aws S3
	file, err := cloudstorage.DownloadFileWithProgress(*localDoc.StorageKey, progressTracker)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.StorageDownloadFailed,
			"Error downloading file",
			fiber.StatusInternalServerError,
		)
	}
	defer file.Close()

	// Step 3: Set appropriate HTTP headers for secure file download
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", localDoc.FileName))
	c.Set("Content-Length", strconv.FormatInt(localDoc.FileSize, 10))

	// Step 4: Stream file content directly to client response
	_, err = io.Copy(c.Response().BodyWriter(), file)
	if err != nil {

		return errors.WrapServer(
			err,
			errors.FSStreamFailed,
			"Error streaming file",
			fiber.StatusInternalServerError,
		)
	}

	progressTracker.Complete()
	snapshot = progressTracker.Snapshot()
	CacheService.PublishProgress(c.Context(), documentID, &snapshot)

	return nil
}

// GetAssignmentDocumentsHandler retrieves all document metadata for a specific models.
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

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	var assignmentID string
	// Step 2: Extract and validate assignment ID from query parameters
	if assignmentID = c.Query("assignment_id"); assignmentID != "" {
		return errors.WrapServer(
			fmt.Errorf("assignment ID required"),
			errors.ReqParamMissing,
			"Assignment ID required",
			fiber.StatusBadRequest,
		)
	}

	assignment := models.Assignment{Base: models.Base{ID: assignmentID}}

	// Step 4: Query documents for the specified assignment with user information
	var documents []models.Document

	if documents, err = assignment.GetDocumentsByAssignment(db); err != nil {
		return errors.WrapServer(
			err,
			errors.DBQueryFailed,
			"Error getting documents from database",
			fiber.StatusInternalServerError,
		)
	}

	// Step 6: Send successful response with document metadata array
	return c.JSON(documents)
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
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}
	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}
	// Step 2: Extract document ID from path parameter
	var docID string
	if docID = c.Params("id"); docID == "" {
		return errors.WrapServer(
			fmt.Errorf("document ID required"),
			errors.ReqParamMissing,
			"Document ID required",
			fiber.StatusBadRequest,
		)
	}

	var doc *models.Document
	if doc, err = models.GetDocument(docID, db.Select("user_id", "has_local_file", "storage_key")); err != nil {
		return errors.WrapServer(err, errors.DBRecordNotFound, "Document not found", fiber.StatusNotFound)
	}

	if doc.UserID != currentUser.ID {
		return errors.WrapServer(
			fmt.Errorf("document not found"),
			errors.DBRecordNotFound,
			"Document not found",
			fiber.StatusNotFound,
		)
	}
	// Step 4: Remove document record from database
	if err := db.Set("qdrantClient", QdrantClient).Delete(&doc).Error; err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(
				err,
				errors.DBRecordNotFound,
				"Document not found",
				fiber.StatusNotFound,
			)
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

	// Step 7: Send successful deletion confirmation
	return nil
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

	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	metadata := c.FormValue("metadata")
	if metadata == "" {
		return errors.WrapServer(
			fmt.Errorf("metadata missing from form"),
			errors.ReqParamMissing,
			"Metadata missing from form",
			fiber.StatusBadRequest,
		)
	}

	var localDoc models.LocalDocument
	if err := json.Unmarshal([]byte(metadata), &localDoc); err != nil {
		return errors.WrapServer(
			err,
			errors.ProcJSONUnmarshalFailed,
			"Invalid metadata format",
			fiber.StatusBadRequest,
		)
	}

	var storageKey string
	if localDoc.StorageKey != nil {
		storageKey = *localDoc.StorageKey
	}

	if err := models.ValidateFileTypeRAG(localDoc.FileName); err != nil {
		return errors.WrapServer(
			err,
			errors.FSFileTypeNotSupported,
			"File type not supported for RAG",
			fiber.StatusUnsupportedMediaType,
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
		_, _, err = WriteMultipartFile(storageKey, fileHeader, c)
		if err != nil {
			return errors.WrapServer(
				err,
				errors.FSWriteFailed,
				"Error writing file to disk",
				fiber.StatusInternalServerError,
			)
		}
	} else {

		fileReader, err := cloudstorage.DownloadFile(storageKey)
		if err != nil {
			return errors.WrapServer(
				err,
				errors.StorageDownloadFailed,
				"Error downloading file",
				fiber.StatusInternalServerError,
			)
		}

		_, _, err = WriteFile(storageKey, fileReader, c)
		if err != nil {
			return errors.WrapServer(
				err,
				errors.FSWriteFailed,
				"Error writing file to disk",
				fiber.StatusInternalServerError,
			)
		}

	}

	doc, err := models.GetDocument(localDoc.ID, db)
	if doc == nil {
		return errors.WrapServer(
			fmt.Errorf("document not found"),
			errors.DBRecordNotFound,
			"Document not found",
			fiber.StatusNotFound,
		)
	}

	vectors, err := fileops.GetQdrantVectors(doc)
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

	collectionName := fmt.Sprintf("unipilot-qdrant-db-%d", localDoc.AssignmentID)

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

	var docID string
	if docID = c.Params("id"); docID == "" {
		return errors.WrapServer(
			fmt.Errorf("document ID required"),
			errors.ReqParamMissing,
			"Document ID required",
			fiber.StatusBadRequest,
		)
	}

	var assignmentID string
	if assignmentID = c.Params("assignment_id"); assignmentID == "" {
		return errors.WrapServer(
			fmt.Errorf("assignment ID required"),
			errors.ReqParamMissing,
			"Assignment ID required",
			fiber.StatusBadRequest,
		)
	}

	var doc = models.Document{
		Base: models.Base{
			ID: docID,
		},
		BaseDocument: models.BaseDocument{
			AssignmentID: assignmentID,
		},
	}

	if err := models.DeleteDocumentVectors(&doc, QdrantClient); err != nil {
		return errors.WrapServer(
			err,
			errors.QdrantDeletePointsError,
			"Error deleting document vectors from Qdrant",
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
	var assignmentID string
	if assignmentID = c.Params("id"); assignmentID == "" {

		return errors.WrapServer(
			fmt.Errorf("assignment ID required"),
			errors.ReqParamMissing,
			"Assignment ID required",
			fiber.StatusBadRequest,
		)
	}

	documentIDs, err := models.GetAssignmentDocumentIDsRAG(assignmentID, QdrantClient)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.QdrantFailed,
			"Error getting assignment document IDs from Qdrant",
			fiber.StatusInternalServerError,
		)
	}

	// Step 5: Return the list of uploaded document IDs
	return c.JSON(documentIDs)
}
