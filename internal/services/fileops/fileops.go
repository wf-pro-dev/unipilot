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
func UploadDocument(req FileUploadRequest, db *gorm.DB) (*FileUploadResponse, error) {
	// Validate file type
	if err := ValidateFileType(req.FileName); err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "File type not supported",
		}, fmt.Errorf("unsupported file type")
	}

	// Validate file size
	if req.FileSize > document.MaxFileSize {
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("File size exceeds limit of %d MB", document.MaxFileSize/(1024*1024)),
		}, fmt.Errorf("file too large")
	}

	// Create LocalDocument record
	localDoc := document.LocalDocument{
		AssignmentID: req.AssignmentID,
		UserID:       req.UserID,
		Type:         req.Type,
		FileName:     req.FileName,
		FileType:     GetMimeType(req.FileName),
		FileSize:     req.FileSize,
		Version:      1,
		HasLocalFile: false, // Will be set to true after successful file write
	}

	// Generate file path
	appDataPath, err := document.GetAppDataPath()
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to get app data path",
		}, err
	}

	// Create unique filename with assignment and user info
	fileName := fmt.Sprintf("doc_%d_%d_%s", req.AssignmentID, req.UserID, req.FileName)
	filePath := filepath.Join(appDataPath, "documents", fileName)
	localDoc.FilePath = filePath

	// Check storage quota
	var totalSize int64
	db.Model(&document.LocalDocument{}).
		Where("user_id = ? AND has_local_file = ?", req.UserID, true).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize)

	if totalSize+req.FileSize > document.MaxUserQuota {
		return &FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("Storage quota exceeded. Current: %d MB, Limit: %d MB",
				totalSize/(1024*1024), document.MaxUserQuota/(1024*1024)),
		}, fmt.Errorf("storage quota exceeded")
	}

	// Save to database first
	if err := db.Create(&localDoc).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to save document record",
		}, err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		// Clean up database record
		db.Delete(&localDoc)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to create directory",
		}, err
	}

	// Write file to disk
	if err := writeFile(filePath, req.FileContent); err != nil {
		// Clean up database record
		db.Delete(&localDoc)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to write file",
		}, err
	}

	// Update HasLocalFile to true after successful write
	localDoc.HasLocalFile = true
	db.Save(&localDoc)

	// Update storage cache
	document.UpdateLocalStorageCache(req.UserID, db)

	return &FileUploadResponse{
		LocalDocument: &localDoc,
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
		AssignmentID: existingDoc.AssignmentID,
		UserID:       existingDoc.UserID,
		Type:         existingDoc.Type,
		FileName:     req.FileName,
		FileType:     GetMimeType(req.FileName),
		FileSize:     req.FileSize,
		Version:      existingDoc.Version + 1,
		ParentDocID:  &existingDoc.ID,
		IsOriginal:   false,
		HasLocalFile: false,
	}

	// Generate file path
	appDataPath, err := document.GetAppDataPath()
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to get app data path",
		}, err
	}

	fileName := fmt.Sprintf("doc_%d_%d_v%d_%s", req.AssignmentID, req.UserID, newVersion.Version, req.FileName)
	filePath := filepath.Join(appDataPath, "documents", fileName)
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
	if err := writeFile(filePath, req.FileContent); err != nil {
		db.Delete(&newVersion)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to write file",
		}, err
	}

	// Update HasLocalFile after successful write
	newVersion.HasLocalFile = true
	db.Save(&newVersion)

	// Update storage cache
	document.UpdateLocalStorageCache(req.UserID, db)

	return &FileUploadResponse{
		LocalDocument: &newVersion,
		Success:       true,
		Message:       "New version uploaded successfully",
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
