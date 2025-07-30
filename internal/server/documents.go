package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"unipilot/internal/models/document"

	"gorm.io/gorm"
)

// DocumentMetadata represents document metadata for API responses
type DocumentMetadata struct {
	ID           uint   `json:"id"`
	LocalID      uint   `json:"local_id"`
	AssignmentID uint   `json:"assignment_id"`
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

	var req struct {
		AssignmentID uint   `json:"assignment_id"`
		LocalID      uint   `json:"local_id"`
		Type         string `json:"type"`
		FileName     string `json:"file_name"`
		FileType     string `json:"file_type"`
		FileSize     int64  `json:"file_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Create document metadata record (FilePath empty for local-only files)
	// Create document metadata record (FilePath empty for local-only files)
	doc := &document.Document{
		AssignmentID: req.AssignmentID,
		LocalID:      req.LocalID,
		UserID:       userID,
		Type:         document.DocumentType(req.Type),
		FileName:     req.FileName,
		FileType:     req.FileType,
		FileSize:     req.FileSize,
		Version:      1,
		IsOriginal:   true,
		FilePath:     "", // Empty since file is local only
	}

	if err := db.Create(doc).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save document metadata: %v", err))
		return
	}

	response := DocumentMetadata{
		ID:           doc.ID,
		LocalID:      doc.LocalID,
		AssignmentID: doc.AssignmentID,
		UserID:       doc.UserID,
		Type:         string(doc.Type),
		FileName:     doc.FileName,
		FileType:     doc.FileType,
		FileSize:     doc.FileSize,
		Version:      doc.Version,
		IsOriginal:   doc.IsOriginal,
		HasLocalFile: true,
		CreatedAt:    doc.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"document": response,
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

	if err := db.Delete(&doc).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, "Failed to delete document")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document metadata deleted",
	})
}
