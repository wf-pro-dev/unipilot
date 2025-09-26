package server

import (
	"os"
	"path/filepath"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"io"
	"log"

	"unipilot/internal/models/document"
	"unipilot/internal/services/cloud_storage"

	"gorm.io/gorm"
)

// DocumentMetadata represents document metadata for API responses
type DocumentMetadata struct {
	ID           uint   `json:"id"`
	LocalID      uint   `json:"local_id"`
	AssignmentID uint   `json:"assignment_id"`
	LocalAssignmentID uint   `json:"local_assignment_id"`
	UserID       uint   `json:"user_id"`
	Type         string `json:"type"`
	FileName     string `json:"file_name"`
	FileType     string `json:"file_type"`
	FileSize     int64  `json:"file_size"`
	Version      int    `json:"version"`
	IsOriginal   bool   `json:"is_original"`
	HasLocalFile bool   `json:"has_local_file"`
	CreatedAt    string `json:"created_at"`
}

// CreateDocumentMetadataHandler stores document metadata remotely
func CreateDocumentMetadataHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	db = db.Debug()
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	// Parse multipart form with max memory (32MB in memory, rest on disk)
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, "Unable to parse multipart form: "+err.Error())
		return
	}
	defer r.MultipartForm.RemoveAll() // Clean up temp files

	log.Println("New document route called")

	//var req struct {
	//	AssignmentID uint   	`json:"assignment_id"`
	//	LocalID      uint   	`json:"local_id"`
	//	Type         string 	`json:"type"`
	//	FileName     string 	`json:"file_name"`
	//	FileContent  os.File 	`json:"file_content"`
	//	FileType     string 	`json:"file_type"`
	//	FileSize     int64  	`json:"file_size"`
	//}

	assignment_id, _ :=  strconv.Atoi(r.FormValue("assignment_id"))
	local_id, _ :=  strconv.Atoi(r.FormValue("local_id"))
	doc_type := r.FormValue("type")
	filename := r.FormValue("file_name")
	filetype := r.FormValue("file_type")
	local_filepath := r.FormValue("file_path")
	filesize, _ := strconv.Atoi(r.FormValue("file_size"))


	// Get the file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, "Unable to get file from form: "+err.Error())
		return
	}
	defer file.Close()

	fileName := fileHeader.Filename
	fileSize := fileHeader.Size

	
	log.Printf("File name: %s, file size: %d, local_id : %d", fileName, fileSize, local_id)
	// Use configuration directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}

	configDir = filepath.Join(configDir,"unipilot")

	// Create user, assignment  directory 
	assignmentDir := fmt.Sprintf("users_data/user_%d/documents/assign_%d", userID, assignment_id)

	// Generate unique filename
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), filename)
	key := fmt.Sprintf("%s/%s", assignmentDir, uniqueFileName)

	filePath := filepath.Join(configDir, uniqueFileName)

	// Save file to disk
	destFile, err := os.Create(filePath)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, "Unable to create destination file")
		return
	}
	defer destFile.Close()

	// Copy file content
	bytesWritten, err := io.Copy(destFile, file)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, "Error saving file")
		return
	}

	log.Printf("File saved: %s (%d bytes)", filePath, bytesWritten)

	// Upload to aws S3
	if err := cloudstorage.UploadFile(filePath, uniqueFileName, key); err != nil {
		
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file: %v",err))
		return
	}

	// Clean up local file after S3 upload
	os.Remove(filePath)

	//if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//	PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
	//	return
	//}


	// Create document metadata record
	doc := &document.Document{
		AssignmentID: uint(assignment_id),
		LocalID:      uint(local_id),
		UserID:       userID,
		Type:         document.DocumentType(doc_type),
		FileName:     filename,
		FileType:     filetype,
		FileSize:     int64(filesize),
		Version:      1,
		IsOriginal:   true,
		FilePath:     local_filepath,
		StorageKey:   key,
	}

	if err := db.Create(doc).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save document metadata: %v", err))
		return
	}

	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		// Log warning but don't fail the request
		fmt.Printf("Warning: Failed to update remote storage info for user %d: %v\n", userID, err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"document": doc,
	})
}

// GetAssignmentDocumentsHandler retrieves document metadata for an assignment
func GetAssignmentDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	currentUserID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	assignmentIDStr := r.URL.Query().Get("assignment_id")
	if assignmentIDStr == "" {
		PrintERROR(w, http.StatusBadRequest, "Assignment ID required")
		return
	}

	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 32)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	var documents []document.Document
	err = db.Preload("User").
		Where("assignment_id = ?", assignmentID).
		Order("created_at DESC").
		Find(&documents).Error

	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, "Failed to get documents")
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

// DeleteDocumentMetadataHandler removes document metadata
func DeleteDocumentMetadataHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)
	db = db.Debug()

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	docID := r.URL.Query().Get("document_id")
	if docID == "" {
		PrintERROR(w, http.StatusBadRequest, "Document ID required")
		return
	}

	var doc document.Document
	if err := db.Where("local_id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		PrintERROR(w, http.StatusNotFound, "Document not found")
		return
	}

	// Delete the document on S3
	if err := cloudstorage.DeleteFile(doc.FilePath); err != nil {
	
		PrintERROR(w, http.StatusNotFound, fmt.Sprint("Failed to delete document on AWS/S3i: %v",err))
		return
	}

	if err := db.Delete(&doc).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, "Failed to delete document")
		return
	}

	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		// Log warning but don't fail the request
		fmt.Printf("Warning: Failed to update remote storage info for user %d: %v\n", userID, err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document metadata deleted",
	})
}
