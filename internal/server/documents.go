package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"unipilot/internal/models/document"
	"unipilot/internal/services/fileops"

	"github.com/gorilla/mux"
)

// DocumentResponse represents a document in API responses
type DocumentResponse struct {
	ID           uint   `json:"id"`
	AssignmentID uint   `json:"assignment_id"`
	Type         string `json:"type"`
	FileName     string `json:"file_name"`
	FileType     string `json:"file_type"`
	FileSize     int64  `json:"file_size"`
	Version      int    `json:"version"`
	IsOriginal   bool   `json:"is_original"`
	CreatedAt    string `json:"created_at"`
}

// StorageInfoResponse represents storage statistics
type StorageInfoResponse struct {
	TotalSize       int64   `json:"total_size"`
	DocumentCount   int     `json:"document_count"`
	MaxUserQuota    int64   `json:"max_user_quota"`
	UsagePercentage float64 `json:"usage_percentage"`
}

// UploadDocumentHandler handles document uploads
func (s *Server) UploadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Only enable if feature flag is on (for safe deployment)
	if !s.isDocumentFeatureEnabled() {
		http.Error(w, "Document feature not enabled", http.StatusNotFound)
		return
	}

	userID, err := s.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(document.MaxFileSize) // Use our defined limit
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	assignmentIDStr := r.FormValue("assignment_id")
	documentType := r.FormValue("type")

	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid assignment ID", http.StatusBadRequest)
		return
	}

	// Validate document type
	if documentType != string(document.DocumentTypeSupport) && documentType != string(document.DocumentTypeSubmission) {
		http.Error(w, "Invalid document type", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create upload request
	uploadReq := fileops.FileUploadRequest{
		AssignmentID: uint(assignmentID),
		UserID:       userID,
		Type:         document.DocumentType(documentType),
		FileName:     header.Filename,
		FileContent:  file,
		FileSize:     header.Size,
	}

	// Upload the document
	response, err := fileops.UploadDocument(uploadReq, s.DB)
	if err != nil {
		http.Error(w, response.Message, http.StatusBadRequest)
		return
	}

	// Convert to response format
	docResponse := DocumentResponse{
		ID:           response.Document.ID,
		AssignmentID: response.Document.AssignmentID,
		Type:         string(response.Document.Type),
		FileName:     response.Document.FileName,
		FileType:     response.Document.FileType,
		FileSize:     response.Document.FileSize,
		Version:      response.Document.Version,
		IsOriginal:   response.Document.IsOriginal,
		CreatedAt:    response.Document.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  response.Message,
		"document": docResponse,
	})
}

// GetAssignmentDocumentsHandler retrieves documents for an assignment
func (s *Server) GetAssignmentDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isDocumentFeatureEnabled() {
		http.Error(w, "Document feature not enabled", http.StatusNotFound)
		return
	}

	userID, err := s.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	assignmentIDStr := vars["assignment_id"]

	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid assignment ID", http.StatusBadRequest)
		return
	}

	// Get latest versions only
	documents, err := document.GetLatestVersions(uint(assignmentID), userID, s.DB)
	if err != nil {
		http.Error(w, "Failed to get documents", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var docResponses []DocumentResponse
	for _, doc := range documents {
		docResponses = append(docResponses, DocumentResponse{
			ID:           doc.ID,
			AssignmentID: doc.AssignmentID,
			Type:         string(doc.Type),
			FileName:     doc.FileName,
			FileType:     doc.FileType,
			FileSize:     doc.FileSize,
			Version:      doc.Version,
			IsOriginal:   doc.IsOriginal,
			CreatedAt:    doc.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"documents": docResponses,
	})
}

// DownloadDocumentHandler handles document downloads
func (s *Server) DownloadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isDocumentFeatureEnabled() {
		http.Error(w, "Document feature not enabled", http.StatusNotFound)
		return
	}

	userID, err := s.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	docIDStr := vars["document_id"]

	docID, err := strconv.ParseUint(docIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	file, doc, err := fileops.DownloadDocument(uint(docID), userID, s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.FileName))
	w.Header().Set("Content-Type", doc.FileType)
	w.Header().Set("Content-Length", strconv.FormatInt(doc.FileSize, 10))

	// Copy file content to response
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to send file", http.StatusInternalServerError)
		return
	}
}

// DeleteDocumentHandler handles document deletion
func (s *Server) DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isDocumentFeatureEnabled() {
		http.Error(w, "Document feature not enabled", http.StatusNotFound)
		return
	}

	userID, err := s.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	docIDStr := vars["document_id"]

	docID, err := strconv.ParseUint(docIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	err = fileops.DeleteDocument(uint(docID), userID, s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document deleted successfully",
	})
}

// GetStorageInfoHandler returns user storage statistics
func (s *Server) GetStorageInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isDocumentFeatureEnabled() {
		http.Error(w, "Document feature not enabled", http.StatusNotFound)
		return
	}

	userID, err := s.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	storageInfo, err := fileops.GetUserStorageInfo(userID, s.DB)
	if err != nil {
		http.Error(w, "Failed to get storage info", http.StatusInternalServerError)
		return
	}

	response := StorageInfoResponse{
		TotalSize:       storageInfo.TotalSize,
		DocumentCount:   storageInfo.DocumentCount,
		MaxUserQuota:    document.MaxUserQuota,
		UsagePercentage: float64(storageInfo.TotalSize) / float64(document.MaxUserQuota) * 100,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"storage": response,
	})
}

// Feature flag helper (for safe deployment)
func (s *Server) isDocumentFeatureEnabled() bool {
	// You can make this configurable via environment variable
	// For now, return true to enable the feature
	return true

	// For safer deployment, you could use:
	// return os.Getenv("ENABLE_DOCUMENTS") == "true"
}

// Helper to register document routes
func (s *Server) RegisterDocumentRoutes() {
	// Only register routes if migration was successful
	if s.checkDocumentMigration() {
		s.Router.HandleFunc("/api/documents/upload", s.UploadDocumentHandler).Methods("POST")
		s.Router.HandleFunc("/api/assignments/{assignment_id}/documents", s.GetAssignmentDocumentsHandler).Methods("GET")
		s.Router.HandleFunc("/api/documents/{document_id}/download", s.DownloadDocumentHandler).Methods("GET")
		s.Router.HandleFunc("/api/documents/{document_id}", s.DeleteDocumentHandler).Methods("DELETE")
		s.Router.HandleFunc("/api/user/storage", s.GetStorageInfoHandler).Methods("GET")
	}
}

// Check if document migration is available
func (s *Server) checkDocumentMigration() bool {
	// Check if the documents table exists
	return s.DB.Migrator().HasTable(&document.Document{})
}
