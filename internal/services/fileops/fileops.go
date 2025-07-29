package fileops

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"unipilot/internal/models/document"

	"gorm.io/gorm"
)

// FileUploadRequest represents a file upload request
type FileUploadRequest struct {
	AssignmentID uint
	UserID       uint
	Type         document.DocumentType
	FileName     string
	FileContent  io.Reader
	FileSize     int64
}

// FileUploadResponse represents the result of a file upload
type FileUploadResponse struct {
	Document *document.Document
	Success  bool
	Message  string
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

// UploadDocument handles the complete file upload process
func UploadDocument(req FileUploadRequest, db *gorm.DB) (*FileUploadResponse, error) {
	// Validate file type
	if err := ValidateFileType(req.FileName); err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Create document record
	doc := &document.Document{
		AssignmentID: req.AssignmentID,
		UserID:       req.UserID,
		Type:         req.Type,
		FileName:     req.FileName,
		FileType:     GetMimeType(req.FileName),
		FileSize:     req.FileSize,
		Version:      1,
		IsOriginal:   true,
	}

	// Validate file size constraints
	if err := doc.ValidateFileSize(db); err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Generate file path
	filePath, err := doc.GenerateFilePath()
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to generate file path: %v", err),
		}, err
	}
	doc.FilePath = filePath

	// Get full path for writing
	fullPath, err := doc.GetFullPath()
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get full path: %v", err),
		}, err
	}

	// Write file to disk
	if err := writeFile(fullPath, req.FileContent); err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to write file: %v", err),
		}, err
	}

	// Save document record to database
	if err := db.Create(doc).Error; err != nil {
		// Clean up file if database insert fails
		os.Remove(fullPath)
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to save document record: %v", err),
		}, err
	}

	// Update user storage info
	if err := document.UpdateStorageInfo(req.UserID, db); err != nil {
		// Log error but don't fail the upload
		fmt.Printf("Warning: Failed to update storage info: %v\n", err)
	}

	return &FileUploadResponse{
		Document: doc,
		Success:  true,
		Message:  "File uploaded successfully",
	}, nil
}

// UploadNewVersion uploads a new version of an existing document
func UploadNewVersion(existingDocID uint, fileName string, fileContent io.Reader, fileSize int64, db *gorm.DB) (*FileUploadResponse, error) {
	// Get existing document
	var existingDoc document.Document
	if err := db.First(&existingDoc, existingDocID).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Existing document not found",
		}, err
	}

	// Validate file type
	if err := ValidateFileType(fileName); err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Create new document for the version
	newDoc := &document.Document{
		AssignmentID: existingDoc.AssignmentID,
		UserID:       existingDoc.UserID,
		Type:         existingDoc.Type,
		FileName:     fileName,
		FileType:     GetMimeType(fileName),
		FileSize:     fileSize,
		IsOriginal:   existingDoc.IsOriginal,
	}

	// Validate file size constraints
	if err := newDoc.ValidateFileSize(db); err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Generate file path
	filePath, err := newDoc.GenerateFilePath()
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to generate file path: %v", err),
		}, err
	}

	// Create new version using existing document's method
	newVersion, err := existingDoc.CreateNewVersion(fileName, fileSize, filePath, db)
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create new version: %v", err),
		}, err
	}

	// Get full path for writing
	fullPath, err := newVersion.GetFullPath()
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get full path: %v", err),
		}, err
	}

	// Write file to disk
	if err := writeFile(fullPath, fileContent); err != nil {
		// Clean up database record if file write fails
		db.Delete(newVersion)
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to write file: %v", err),
		}, err
	}

	// Update user storage info
	if err := document.UpdateStorageInfo(newVersion.UserID, db); err != nil {
		fmt.Printf("Warning: Failed to update storage info: %v\n", err)
	}

	return &FileUploadResponse{
		Document: newVersion,
		Success:  true,
		Message:  "New version uploaded successfully",
	}, nil
}

// DownloadDocument retrieves a document file for download
func DownloadDocument(docID uint, userID uint, db *gorm.DB) (*os.File, *document.Document, error) {
	// Get document record
	var doc document.Document
	if err := db.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		return nil, nil, fmt.Errorf("document not found or access denied")
	}

	// Check if file exists
	if !doc.FileExists() {
		return nil, nil, fmt.Errorf("file not found on disk")
	}

	// Get full path
	fullPath, err := doc.GetFullPath()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file path: %w", err)
	}

	// Open file for reading
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, &doc, nil
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
func writeFile(filePath string, content io.Reader) error {
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
