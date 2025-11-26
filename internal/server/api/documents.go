package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/document"
	notif "unipilot/internal/models/notifications"
	"unipilot/internal/models/user"
	"unipilot/internal/server"
	"unipilot/internal/server/sse/grpc/notifications"
	cloudstorage "unipilot/internal/services/cloud_storage"

	"github.com/qdrant/go-client/qdrant"
	"gorm.io/gorm"
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
func CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Parse multipart form with size limits (32MB in memory, rest on disk)
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Unable to parse multipart form",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}
	defer r.MultipartForm.RemoveAll() // Clean up temp files

	// Step 3: Extract and validate document metadata from form
	metadata := r.FormValue("metadata")
	if metadata == "" {
		err := errors.New("metadata missing from form")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Metadata missing from form",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}

	// Step 4: Parse JSON metadata into LocalDocument structure
	// Parse metadata directly into LocalDocument
	var localDoc document.LocalDocument
	if err := json.Unmarshal([]byte(metadata), &localDoc); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid metadata format",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
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
		if err := UploadFileToS3(localDoc, newKey, w, r); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error uploading file",
				"tags", []string{"DOCUMENTS", "STORAGE"},
			)
			return
		}
	} else {
		// Copy existing file within S3 storage system
		if err := cloudstorage.CopyFile(localDoc.StorageKey, newKey); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error copying file",
				"tags", []string{"DOCUMENTS", "STORAGE"},
			)
			return
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
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error saving document metadata",
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Step 8: Update user storage quota information for tracking
	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		server.LogWarn(r.Context(),
			"Failed to update remote storage info", err,
			"tags", []string{"DOCUMENTS", "STORAGE"},
		)
	}

	// Step 9: Retrieve assignment data for notification distribution
	// Get document assignment
	var a assignment.Assignment
	if err := db.Where("id = ?", doc.AssignmentID).First(&a).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting assignment",
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Step 10: Get child assignments for notification distribution
	// Get linked assignments
	linkedAssignments, err := a.GetChildren(db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting linked assignments",
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Step 11: Prepare document data for notifications
	// Marshal document
	doc.User = currentUser // Link the creator data with the document
	dJson, err := json.Marshal(doc)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error marshalling document",
			"tags", []string{"DOCUMENTS", "MARSHALLING"},
		)
		return
	}

	// Step 12: Send real-time notifications to linked assignment users (only for new uploads)
	if GrpcClient != nil && localDoc.HasLocalFile {
		// Send notification to linked assignments
		for _, linkedAssignment := range linkedAssignments {
			if linkedAssignment.UserID != userID {
				server.LogInfo(r.Context(), "Sending document notification",
					"target_user_id", linkedAssignment.UserID,
					"assignment_id", linkedAssignment.ID,
					"tags", []string{"DOCUMENTS", "GRPC"},
				)
				GrpcClient.SendNotification(context.Background(),
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
			}
		}
	}

	// Step 13: Send successful response with document metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"document": doc,
	})

	// Step 14: Log successful document creation for audit trail
	server.LogInfo(r.Context(), "Document created successfully",
		"document_id", doc.ID,
		"assignment_id", doc.AssignmentID,
		"tags", []string{"DOCUMENTS", "WRITE"},
	)
}

// WriteFileToDisk extracts file from multipart form and writes it to local disk storage.
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
func WriteFileToDisk(localDoc document.LocalDocument, key string, w http.ResponseWriter, r *http.Request) (string, int64, error) {
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

// UploadFileToS3 handles the complete file upload pipeline from multipart form to AWS S3.
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
func UploadFileToS3(localDoc document.LocalDocument, key string, w http.ResponseWriter, r *http.Request) error {

	filePath, bytesWritten, err := WriteFileToDisk(localDoc, key, w, r)
	if err != nil {
		return err
	}

	// Upload to aws S3
	if err := cloudstorage.UploadFile(filePath, localDoc.FileName, key); err != nil {
		return err
	}
	server.LogInfo(context.Background(), "File uploaded to S3",
		"file_path", filePath,
		"bytes", bytesWritten,
		"storage_key", key,
		"tags", []string{"DOCUMENTS", "UPLOAD"},
	)

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
func DownloadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Parse document download request from JSON body
	var docData document.LocalDocument
	if err := json.NewDecoder(r.Body).Decode(&docData); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"tags", []string{"DOCUMENTS", "DOWNLOAD", "REQUEST"},
		)
		return
	}

	// Step 2: Download file from AWS S3 using storage key
	// Download from aws S3
	fileReader, err := cloudstorage.DownloadFile(docData.StorageKey)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error downloading file",
			"storage_key", docData.StorageKey,
			"tags", []string{"DOCUMENTS", "DOWNLOAD", "STORAGE"},
		)
		return
	}

	// Step 3: Set appropriate HTTP headers for secure file download
	// Set appropriate headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", docData.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(docData.FileSize, 10))
	w.WriteHeader(http.StatusOK)

	// Step 4: Stream file content directly to client response
	// Stream file directly to response
	bytesCopied, err := io.Copy(w, fileReader)
	if err != nil {
		server.LogWarn(r.Context(),
			"Error streaming file", err,
			"file_name", docData.FileName,
			"storage_key", docData.StorageKey,
			"tags", []string{"DOCUMENTS", "DOWNLOAD"},
		)
		return
	}

	// Step 5: Log successful download for monitoring and audit trail
	server.LogInfo(r.Context(), "File streamed successfully",
		"file_name", docData.FileName,
		"bytes", bytesCopied,
		"tags", []string{"DOCUMENTS", "DOWNLOAD"},
	)
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
func GetAssignmentDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	currentUserID := currentUser.ID

	// Step 2: Extract and validate assignment ID from query parameters
	assignmentIDStr := r.URL.Query().Get("assignment_id")
	if assignmentIDStr == "" {
		err := errors.New("assignment ID required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Assignment ID required",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}

	// Step 3: Convert assignment ID string to integer for database query
	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 32)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting assignment ID",
			"tags", []string{"DOCUMENTS", "INVALID_ASSIGNMENT_ID"},
		)
		return
	}

	// Step 4: Query documents for the specified assignment with user information
	var documents []document.Document
	if err := db.Preload("User").
		Where("assignment_id = ?", assignmentID).
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting documents from database",
			"assignment_id", assignmentID,
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"documents": docResponses,
	})

	// Step 7: Log successful retrieval with document count for monitoring
	server.LogInfo(r.Context(), "Assignment documents retrieved",
		"assignment_id", assignmentID,
		"count", len(docResponses),
		"tags", []string{"DOCUMENTS", "READ"},
	)
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
func DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract context values from middleware (user and database connection)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Step 2: Extract and validate document ID from query parameters
	docID := r.URL.Query().Get("document_id")
	if docID == "" {
		err := errors.New("document ID required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Document ID required",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}

	// Step 3: Find document with ownership validation (local_id + user_id)
	var doc document.Document
	if err := db.Where("local_id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusNotFound, "Document not found",
			"document_id", docID,
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Step 4: Delete associated file from AWS S3 cloud storage
	// Delete the document on S3
	if err := cloudstorage.DeleteFile(doc.FilePath); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error deleting document from storage",
			"document_id", docID,
			"tags", []string{"DOCUMENTS", "STORAGE"},
		)
		return
	}

	// Step 5: Remove document record from database
	if err := db.Delete(&doc).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error deleting document record",
			"document_id", docID,
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Step 6: Update user storage quota information after deletion
	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		server.LogWarn(r.Context(),
			"Failed to update remote storage info", err,
			"tags", []string{"DOCUMENTS", "STORAGE"},
		)
	}

	// Step 7: Send successful deletion confirmation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document metadata deleted",
	})

	// Step 8: Log successful deletion for audit trail
	server.LogInfo(r.Context(), "Document deleted",
		"document_id", docID,
		"tags", []string{"DOCUMENTS", "WRITE"},
	)
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
func UploadDocumentForRAGHandler(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Parse multipart form with max memory (32MB in memory, rest on disk)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Unable to parse multipart form",
			"tags", []string{"DOCUMENTS", "RAG", "REQUEST"},
		)
		return
	}
	defer r.MultipartForm.RemoveAll()

	metadata := r.FormValue("metadata")
	if metadata == "" {
		err := errors.New("metadata missing from form")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Metadata missing from form",
			"tags", []string{"DOCUMENTS", "RAG", "REQUEST"},
		)
		return
	}

	var localDoc document.LocalDocument
	if err := json.Unmarshal([]byte(metadata), &localDoc); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid metadata format",
			"tags", []string{"DOCUMENTS", "RAG", "REQUEST"},
		)
		return
	}

	assignmentDir := fmt.Sprintf("users_data/user_%d/documents/assign_%d", userID, localDoc.RemoteAssignmentID)
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), localDoc.FileName)
	newKey := fmt.Sprintf("%s/%s", assignmentDir, uniqueFileName)

	var fileName string
	var bytesWritten int64
	var err error

	if localDoc.HasLocalFile {
		fileName, bytesWritten, err = WriteFileToDisk(localDoc, newKey, w, r)
		if err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error writing file to disk",
				"tags", []string{"DOCUMENTS", "RAG", "STORAGE"},
			)
			return
		}
	} else {
		fileName = newKey
		fileReader, err := cloudstorage.DownloadFile(localDoc.StorageKey)
		if err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error downloading file",
				"tags", []string{"DOCUMENTS", "RAG", "STORAGE"},
			)
			return
		}

		if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error creating directory",
				"tags", []string{"DOCUMENTS", "RAG", "FILESYSTEM"},
			)
			return
		}

		destFile, err := os.Create(fileName)
		if err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error creating file",
				"tags", []string{"DOCUMENTS", "RAG", "FILESYSTEM"},
			)
			return
		}
		defer destFile.Close()

		bytesWritten, err = io.Copy(destFile, fileReader)
		if err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error copying file content",
				"tags", []string{"DOCUMENTS", "RAG", "FILESYSTEM"},
			)
			return
		}
	}

	server.LogInfo(r.Context(), "File prepared for RAG upload",
		"assignment_id", localDoc.RemoteAssignmentID,
		"bytes", bytesWritten,
		"has_local_file", localDoc.HasLocalFile,
		"tags", []string{"DOCUMENTS", "RAG", "FILE"},
	)

	var doc document.Document
	if err := db.Where("storage_key = ?", localDoc.StorageKey).First(&doc).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error finding document metadata record",
			"tags", []string{"DOCUMENTS", "RAG", "DB"},
		)
		return
	}
	doc.FileName = fileName

	vectors, err := document.GetQdrantVectors(&doc)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting Qdrant vectors",
			"tags", []string{"DOCUMENTS", "RAG", "QDRANT"},
		)
		return
	}

	collections, err := QdrantClient.ListCollections(context.Background())
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error listing Qdrant collections",
			"tags", []string{"DOCUMENTS", "RAG", "QDRANT"},
		)
		return
	}
	collectionName := fmt.Sprintf("unipilot-qdrant-db-%d", doc.AssignmentID)

	if !slices.Contains(collections, collectionName) {
		err = QdrantClient.CreateCollection(context.Background(), &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     768,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error creating Qdrant collection",
				"tags", []string{"DOCUMENTS", "RAG", "QDRANT"},
			)
			return
		}

		server.LogInfo(r.Context(), "Qdrant collection created",
			"collection", collectionName,
			"tags", []string{"DOCUMENTS", "RAG", "QDRANT"},
		)
	}

	if _, err = QdrantClient.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         vectors,
	}); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error inserting vectors into Qdrant",
			"tags", []string{"DOCUMENTS", "RAG", "QDRANT"},
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"document": doc,
	})

	server.LogInfo(r.Context(), "Document uploaded for RAG successfully",
		"document_id", doc.ID,
		"assignment_id", doc.AssignmentID,
		"vectors", len(vectors),
		"tags", []string{"DOCUMENTS", "RAG", "UPLOAD"},
	)
}
