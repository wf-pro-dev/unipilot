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
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	// Parse multipart form with max memory (32MB in memory, rest on disk)
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Unable to parse multipart form",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}
	defer r.MultipartForm.RemoveAll() // Clean up temp files

	metadata := r.FormValue("metadata")
	if metadata == "" {
		err := errors.New("metadata missing from form")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Metadata missing from form",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}

	// Parse metadata directly into LocalDocument
	var localDoc document.LocalDocument
	if err := json.Unmarshal([]byte(metadata), &localDoc); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid metadata format",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}

	// Create user, assignment  directory
	assignmentDir := fmt.Sprintf("users_data/user_%d/documents/assign_%d", userID, localDoc.RemoteAssignmentID)

	// Generate unique filename
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), localDoc.FileName)
	newKey := fmt.Sprintf("%s/%s", assignmentDir, uniqueFileName)

	if localDoc.HasLocalFile {
		if err := UploadFileToS3(localDoc, newKey, w, r); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error uploading file",
				"tags", []string{"DOCUMENTS", "STORAGE"},
			)
			return
		}
	} else {
		// Copy file in aws S3
		if err := cloudstorage.CopyFile(localDoc.StorageKey, newKey); err != nil {
			server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error copying file",
				"tags", []string{"DOCUMENTS", "STORAGE"},
			)
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
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error saving document metadata",
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		server.LogWarn(r.Context(),
			"Failed to update remote storage info", err,
			"tags", []string{"DOCUMENTS", "STORAGE"},
		)
	}

	// Get document assignment
	var a assignment.Assignment
	if err := db.Where("id = ?", doc.AssignmentID).First(&a).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting assignment",
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Get linked assignments
	linkedAssignments, err := a.GetChildren(db)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error getting linked assignments",
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Marshal document
	doc.User = currentUser // Link the creator data with the document
	dJson, err := json.Marshal(doc)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error marshalling document",
			"tags", []string{"DOCUMENTS", "MARSHALLING"},
		)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"document": doc,
	})

	server.LogInfo(r.Context(), "Document created successfully",
		"document_id", doc.ID,
		"assignment_id", doc.AssignmentID,
		"tags", []string{"DOCUMENTS", "WRITE"},
	)
}

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

// UploadFileToS3 uploads a file to the server
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

// DownloadDocumentHandler stores document metadata remotely
func DownloadDocumentHandler(w http.ResponseWriter, r *http.Request) {

	var docData document.LocalDocument
	if err := json.NewDecoder(r.Body).Decode(&docData); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Invalid request body",
			"tags", []string{"DOCUMENTS", "DOWNLOAD", "REQUEST"},
		)
		return
	}

	// Download from aws S3
	fileReader, err := cloudstorage.DownloadFile(docData.StorageKey)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error downloading file",
			"storage_key", docData.StorageKey,
			"tags", []string{"DOCUMENTS", "DOWNLOAD", "STORAGE"},
		)
		return
	}

	// Set appropriate headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", docData.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(docData.FileSize, 10))
	w.WriteHeader(http.StatusOK)

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

	server.LogInfo(r.Context(), "File streamed successfully",
		"file_name", docData.FileName,
		"bytes", bytesCopied,
		"tags", []string{"DOCUMENTS", "DOWNLOAD"},
	)
}

// GetAssignmentDocumentsHandler retrieves document metadata for an assignment
func GetAssignmentDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	currentUserID := currentUser.ID

	assignmentIDStr := r.URL.Query().Get("assignment_id")
	if assignmentIDStr == "" {
		err := errors.New("assignment ID required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Assignment ID required",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}

	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 32)
	if err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Error converting assignment ID",
			"tags", []string{"DOCUMENTS", "INVALID_ASSIGNMENT_ID"},
		)
		return
	}

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

	server.LogInfo(r.Context(), "Assignment documents retrieved",
		"assignment_id", assignmentID,
		"count", len(docResponses),
		"tags", []string{"DOCUMENTS", "READ"},
	)
}

// DeleteDocumentHandler removes document metadata
func DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	docID := r.URL.Query().Get("document_id")
	if docID == "" {
		err := errors.New("document ID required")
		server.ResponseError(r.Context(), w, err, http.StatusBadRequest, "Document ID required",
			"tags", []string{"DOCUMENTS", "REQUEST"},
		)
		return
	}

	var doc document.Document
	if err := db.Where("local_id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusNotFound, "Document not found",
			"document_id", docID,
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Delete the document on S3
	if err := cloudstorage.DeleteFile(doc.FilePath); err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error deleting document from storage",
			"document_id", docID,
			"tags", []string{"DOCUMENTS", "STORAGE"},
		)
		return
	}

	if err := db.Delete(&doc).Error; err != nil {
		server.ResponseError(r.Context(), w, err, http.StatusInternalServerError, "Error deleting document record",
			"document_id", docID,
			"tags", []string{"DOCUMENTS", "DB"},
		)
		return
	}

	// Update remote storage info for the user
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		server.LogWarn(r.Context(),
			"Failed to update remote storage info", err,
			"tags", []string{"DOCUMENTS", "STORAGE"},
		)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document metadata deleted",
	})

	server.LogInfo(r.Context(), "Document deleted",
		"document_id", docID,
		"tags", []string{"DOCUMENTS", "WRITE"},
	)
}

// UploadDocumentForRAGHandler stores document metadata remotely
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
