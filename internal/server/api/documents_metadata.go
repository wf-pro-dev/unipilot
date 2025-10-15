package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	//"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/document"

	//"unipilot/internal/models/notifications"
	"context"
	"unipilot/internal/models"
	notif "unipilot/internal/models/notifications"
	"unipilot/internal/models/user"
	"unipilot/internal/server"
	"unipilot/internal/server/sse/grpc/notifications"
	cloudstorage "unipilot/internal/services/cloud_storage"

	"gorm.io/gorm"
)

// DocumentMetadata represents document metadata for API responses
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

// CreateDocumentHandler stores document metadata remotely
func CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {

	db := r.Context().Value("db").(*gorm.DB)

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	var currentUser user.User
	if err := db.First(&currentUser, userID).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Parse multipart form with max memory (32MB in memory, rest on disk)
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, "Unable to parse multipart form: "+err.Error())
		return
	}
	defer r.MultipartForm.RemoveAll() // Clean up temp files

	metadata := r.FormValue("metadata")
	if metadata == "" {
		server.PrintERROR(w, http.StatusBadRequest, "Unable to get metadata from form: "+err.Error())
		return
	}

	// Parse metadata directly into LocalDocument

	var localDoc document.LocalDocument
	err = json.Unmarshal([]byte(metadata), &localDoc)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, "Invalid metadata format: "+err.Error())
		return
	}

	// Create user, assignment  directory
	assignmentDir := fmt.Sprintf("users_data/user_%d/documents/assign_%d", userID, localDoc.RemoteAssignmentID)

	// Generate unique filename
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), localDoc.FileName)

	newKey := fmt.Sprintf("%s/%s", assignmentDir, uniqueFileName)

	if localDoc.HasLocalFile {
		if err := UploadFile(localDoc, newKey, w, r); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file: %v", err))
			return
		}
	} else {

		// Copy file in aws S3
		if err := cloudstorage.CopyFile(localDoc.StorageKey, newKey); err != nil {
			server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to copy file: %v", err))
			return
		}

	}

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
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save document metadata: %v", err))
		return
	}

	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		// Log warning but don't fail the request
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Warning: Failed to update remote storage info for user %d: %v \n", userID, err))
	}

	// Get document assignment
	var a assignment.Assignment
	if err := db.Where("id = ?", doc.AssignmentID).First(&a).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get assignment: %v", err))
		return
	}

	// Get linked assignments
	var linkedAssignments []assignment.Assignment
	if linkedAssignments, err = a.GetChildren(db); err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get linked assignments: %v", err))
		return
	}

	// Marshal document
	doc.User = currentUser // Link the creator data with the document
	dJson, err := json.Marshal(doc)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to marshal document: %v", err))
		return
	}

	if GrpcClient != nil && localDoc.HasLocalFile {
		// Send notification to linked assignments

		for _, linkedAssignment := range linkedAssignments {
			if linkedAssignment.UserID != userID {
				server.PrintLOG([]string{"SSE", "GRPC"}, fmt.Sprintf("sending to : %d ", linkedAssignment.UserID))
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"document": doc,
	})

	server.PrintLOG([]string{"SUCCESS", "CREATE", "DOCUMENT"}, fmt.Sprintf("User ID : %v, Document ID : %v", userID, doc.ID))
}

func UploadFile(localDoc document.LocalDocument, key string, w http.ResponseWriter, r *http.Request) error {

	// Get the file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, "Unable to get file from form: "+err.Error())
		return err
	}
	defer file.Close()

	log.Printf("File name: %s, file size: %d, local_id : %d", localDoc.FileName, localDoc.FileSize, localDoc.ID)
	// Use uploads directory from Docker volume
	uploadsDir := "/app/uploads"
	filePath := filepath.Join(uploadsDir, key)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	// Save file to disk
	destFile, err := os.Create(filePath)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Unable to create destination file")
		return err
	}
	defer destFile.Close()

	// Copy file content
	bytesWritten, err := io.Copy(destFile, file)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Error saving file")
		return err
	}

	log.Printf("File saved: %s (%d bytes)", filePath, bytesWritten)

	// Upload to aws S3
	if err := cloudstorage.UploadFile(filePath, localDoc.FileName, key); err != nil {

		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file: %v", err))
		return err
	}

	// Clean up local file after S3 upload
	os.Remove(filePath)

	return nil
}

// DownloadDocumentHandler stores document metadata remotely
func DownloadDocumentHandler(w http.ResponseWriter, r *http.Request) {

	/*	userIDVal := r.Context().Value("user_id")
		if userIDVal == nil {
			server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
			return
		}

		userID, ok := userIDVal.(uint)
		if !ok {
			server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
			return
		}
	*/
	var err error

	var docData document.LocalDocument
	if err := json.NewDecoder(r.Body).Decode(&docData); err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// 1. Download file content

	// Donwload from aws S3
	var fileReader io.Reader
	if fileReader, err = cloudstorage.DownloadFile(docData.StorageKey); err != nil {

		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to downloadfile: %v", err))
		return
	}

	/*filePath := filepath.Join(configDir, uniqueFileName)

	// Save file to disk
	destFile, err := os.Create(filePath)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Unable to create destination file")
		return
	}
	defer destFile.Close()

	// Copy file content
	bytesWritten, err := io.Copy(destFile, file)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Error saving file")
		return
	}

	log.Printf("File saved: %s (%d bytes)", filePath, bytesWritten)

	// Clean up local file after S3 upload
	os.Remove(filePath)*/

	// Set appropriate headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", docData.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(docData.FileSize, 10))
	w.WriteHeader(http.StatusOK)

	// Stream file directly to response
	bytesCopied, err := io.Copy(w, fileReader)
	if err != nil {
		log.Printf("Error streaming file: %v", err)
		return
	}
	log.Printf("Successfully streamed file: %s (%d bytes)", docData.FileName, bytesCopied)

}

// GetAssignmentDocumentsHandler retrieves document metadata for an assignment
func GetAssignmentDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	currentUserID, ok := userIDVal.(uint)
	if !ok {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	assignmentIDStr := r.URL.Query().Get("assignment_id")
	if assignmentIDStr == "" {
		server.PrintERROR(w, http.StatusBadRequest, "Assignment ID required")
		return
	}

	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 32)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	var documents []document.Document
	err = db.Preload("User").
		Where("assignment_id = ?", assignmentID).
		Order("created_at DESC").
		Find(&documents).Error

	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Failed to get documents")
		return
	}

	var docResponses []DocumentMetadata
	for _, doc := range documents {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"documents": docResponses,
	})
}

// DeleteDocumentHandler removes document metadata
func DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {

	db := r.Context().Value("db").(*gorm.DB)

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		server.PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	docID := r.URL.Query().Get("document_id")
	if docID == "" {
		server.PrintERROR(w, http.StatusBadRequest, "Document ID required")
		return
	}

	var doc document.Document
	if err := db.Where("local_id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		server.PrintERROR(w, http.StatusNotFound, "Document not found")
		return
	}

	// Delete the document on S3
	if err := cloudstorage.DeleteFile(doc.FilePath); err != nil {

		server.PrintERROR(w, http.StatusNotFound, fmt.Sprintf("Failed to delete document on AWS S3: %v", err))
		return
	}

	if err := db.Delete(&doc).Error; err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Failed to delete document")
		return
	}

	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		// Log warning but don't fail the request
		server.PrintLOG([]string{"WARNING", "UPDATE", "DOCUMENT"}, fmt.Sprintf("Failed to update remote storage info for user %d: %v\n", userID, err))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document metadata deleted",
	})
}
