package fileops

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"unipilot/internal/models/document"
	"unipilot/internal/services/utils"

	"gorm.io/gorm"
)

// FileUploadRequest represents a file upload request
type FileUploadRequest struct {
	AssignmentID       uint
	RemoteAssignmentID uint
	UserID             uint
	Type               document.DocumentType
	FileName           string
	FilePath           string
	FileSize           int64
	FileContent        io.Reader
	StorageKey         string
}

// FileUploadResponse represents the result of a file upload
type FileUploadResponse struct {
	LocalDocument *document.LocalDocument
	Success       bool
	Message       string
}

// SupportedFileTypes defines the allowed file extensions
var SupportedFileTypes = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".ppt":  true,
	".pptx": true,
	".xls":  true,
	".xlsx": true,
	".txt":  true,
	".md":   true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".bmp":  true,
	".svg":  true,
}

// ValidateFileType checks if the file extension is supported
func ValidateFileType(fileName string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	if !SupportedFileTypes[ext] {
		return fmt.Errorf("file type %s is not supported", ext)
	}
	return nil
}

// GetMimeType returns the MIME type for a file extension
func GetMimeType(fileName string) string {
	ext := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

// UploadDocument handles the local file upload process
func WriteDocument(document *document.LocalDocument, fileContent io.Reader, db *gorm.DB) (*FileUploadResponse, error) {

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(document.FilePath), 0755); err != nil {

		return &FileUploadResponse{
			Success: false,
			Message: "Failed to create directory",
		}, err
	}

	// Write file to disk
	if err := WriteFile(document.FilePath, fileContent); err != nil {
		// Clean up database record
		db.Delete(&document)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to write file",
		}, err
	}
	log.Println("Writing file to disk")

	// Update HasLocalFile to true after successful write
	document.HasLocalFile = true
	if err := db.Save(&document).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to update HasLocalFile",
		}, err
	}

	log.Println("Updating HasLocalFile to true")

	// Storage info is now calculated on-demand, no need to update cache

	return &FileUploadResponse{
		LocalDocument: document,
		Success:       true,
		Message:       "Upload successful",
	}, nil
}

// UploadNewVersion creates a new version of an existing document
func UploadNewVersion(existingDocumentID uint, req FileUploadRequest, db *gorm.DB) (*FileUploadResponse, error) {
	// Get existing document
	var existingDoc document.LocalDocument
	if err := db.First(&existingDoc, existingDocumentID).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Original document not found",
		}, err
	}

	// Validate file type
	if err := ValidateFileType(req.FileName); err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "File type not supported",
		}, fmt.Errorf("unsupported file type")
	}

	// Create new version
	newVersion := document.LocalDocument{
		AssignmentID:       existingDoc.AssignmentID,
		RemoteAssignmentID: existingDoc.RemoteAssignmentID,
		UserID:             existingDoc.UserID,
		Type:               existingDoc.Type,
		FileName:           req.FileName,
		FileType:           GetMimeType(req.FileName),
		FileSize:           req.FileSize,
		Version:            existingDoc.Version + 1,
		ParentDocID:        &existingDoc.ID,
		IsOriginal:         false,
		HasLocalFile:       false,
	}

	// Generate file path
	documentDir, err := utils.GetDocumentDir()
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to get app data path",
		}, err
	}

	fileName := fmt.Sprintf("doc_%d_%d_v%d_%s", req.AssignmentID, req.UserID, newVersion.Version, req.FileName)
	filePath := filepath.Join(documentDir, fileName)
	newVersion.FilePath = filePath

	// Save to database
	if err := db.Create(&newVersion).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to save new version",
		}, err
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		db.Delete(&newVersion)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to create directory",
		}, err
	}

	// Write file
	fileContent, err := os.Open(req.FilePath)
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to open file",
		}, err
	}
	defer fileContent.Close()
	if err := WriteFile(filePath, fileContent); err != nil {
		db.Delete(&newVersion)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to write file",
		}, err
	}

	// Update HasLocalFile after successful write
	newVersion.HasLocalFile = true
	db.Save(&newVersion)

	// Storage info is now calculated on-demand, no need to update cache

	return &FileUploadResponse{
		LocalDocument: &newVersion,
		Success:       true,
		Message:       "New version uploaded successfully",
	}, nil
}

// DeleteDocument removes a document and its file
func DeleteDocument(docID uint, userID uint, db *gorm.DB) error {
	// Get document record
	var doc document.Document
	if err := db.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		return fmt.Errorf("document not found or access denied")
	}

	// Get full path
	fullPath, err := doc.GetFullPath()
	if err != nil {
		return fmt.Errorf("failed to get file path: %w", err)
	}

	// Delete file from disk
	if doc.FileExists() {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// Delete all versions if this is the parent document
	if doc.ParentDocID == nil {
		if err := db.Where("parent_doc_id = ?", doc.ID).Delete(&document.Document{}).Error; err != nil {
			return fmt.Errorf("failed to delete document versions: %w", err)
		}
	}

	// Delete document record
	if err := db.Delete(&doc).Error; err != nil {
		return fmt.Errorf("failed to delete document record: %w", err)
	}

	// Update user storage info
	if err := document.UpdateStorageInfo(userID, db); err != nil {
		fmt.Printf("Warning: Failed to update storage info: %v\n", err)
	}

	return nil
}

// writeFile writes content to a file path
func WriteFile(filePath string, content io.Reader) error {
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content to file
	_, err = io.Copy(file, content)
	if err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

// GetUserStorageInfo returns storage statistics for a user
func GetUserStorageInfo(userID uint, db *gorm.DB) (*document.DocumentStorageInfo, error) {
	var storageInfo document.DocumentStorageInfo
	err := db.Where("user_id = ?", userID).First(&storageInfo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new storage info if it doesn't exist
			if err := document.UpdateStorageInfo(userID, db); err != nil {
				return nil, fmt.Errorf("failed to create storage info: %w", err)
			}
			// Try again
			err = db.Where("user_id = ?", userID).First(&storageInfo).Error
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get storage info: %w", err)
		}
	}

	return &storageInfo, nil
}
