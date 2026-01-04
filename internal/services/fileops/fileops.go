package fileops

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/services/utils"
)

// FileUploadRequest represents a file upload request
type FileUploadRequest struct {
	UploadID           string
	AssignmentID       uint
	RemoteAssignmentID uint
	UserID             uint
	Type               models.DocumentType
	FileName           string
	FilePath           string
	FileSize           int64
	FileContent        io.Reader
	StorageKey         string
}

// FileUploadResponse represents the result of a file upload
type FileUploadResponse struct {
	LocalDocument *models.LocalDocument
	Success       bool
	Message       string
}

func PickFile(ctx context.Context) (string, error) {

	// Open file dialog
	filters := []runtime.FileFilter{
		{
			DisplayName: "Documents",
			Pattern:     "*.pdf;*.doc;*.docx;*.ppt;*.pptx;*.xls;*.xlsx;*.txt;*.md",
		},
		{
			DisplayName: "Images",
			Pattern:     "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.svg",
		},
		{
			DisplayName: "All Files",
			Pattern:     "*",
		},
	}

	filePath, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title:   "Select Document to Upload",
		Filters: filters,
	})

	if err != nil {
		return "", errors.Wrap(err, errors.FSOpenFailed, "Failed to open file dialog")
	}

	return filePath, nil
}

func GetFileName(filePath string) string {
	return filepath.Base(filePath)
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

// WriteDocument writes a document to the local file system
func WriteDocument(document *models.LocalDocument, fileContent io.Reader, db *gorm.DB) (*FileUploadResponse, error) {

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(document.FilePath), 0755); err != nil {

		return &FileUploadResponse{
			Success: false,
			Message: "Failed to create directory",
		}, errors.Wrap(err, errors.FSDirFailed, "Failed to create directory")
	}

	// Write file to disk
	if err := WriteFile(document.FilePath, fileContent); err != nil {
		// Clean up database record
		db.Delete(&document)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to write file",
		}, errors.Wrap(err, errors.FSWriteFailed, "Failed to write file")
	}

	// Update HasLocalFile to true after successful write
	document.HasLocalFile = true
	if err := db.Save(&document).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to update HasLocalFile",
		}, errors.Wrap(err, errors.DBQueryFailed, "Failed to update HasLocalFile")
	}

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
	var existingDoc models.LocalDocument
	if err := db.First(&existingDoc, existingDocumentID).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Original document not found",
		}, errors.Wrap(err, errors.DBRecordNotFound, "Original document not found")
	}

	// Create new version
	newVersion := models.LocalDocument{
		BaseDocument: models.BaseDocument{
			AssignmentID: existingDoc.AssignmentID,
			Type:         existingDoc.Type,
			FileName:     req.FileName,
			FileSize:     req.FileSize,
			Version:      existingDoc.Version + 1,
			ParentDocID:  &existingDoc.ID,
			IsOriginal:   false,
			HasLocalFile: false,
		},
		RemoteAssignmentID: existingDoc.RemoteAssignmentID,
	}

	// Generate file path
	documentDir, err := utils.GetDocumentDir()
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to get app data path",
		}, errors.Wrap(err, errors.FSFileNotFound, "Failed to get app data path")
	}

	fileName := fmt.Sprintf("doc_%d_%d_v%d_%s", req.AssignmentID, req.UserID, newVersion.Version, req.FileName)
	filePath := filepath.Join(documentDir, fileName)
	newVersion.FilePath = filePath

	// Save to database
	if err := db.Create(&newVersion).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to save new version",
		}, errors.Wrap(err, errors.DBQueryFailed, "Failed to save new version")
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		db.Delete(&newVersion)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to create directory",
		}, errors.Wrap(err, errors.FSDirFailed, "Failed to create directory")
	}

	// Write file
	fileContent, err := os.Open(req.FilePath)
	if err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to open file",
		}, errors.Wrap(err, errors.FSOpenFailed, "Failed to open file")
	}
	defer fileContent.Close()
	if err := WriteFile(filePath, fileContent); err != nil {
		db.Delete(&newVersion)
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to write file",
		}, errors.Wrap(err, errors.FSWriteFailed, "Failed to write file")
	}

	// Update HasLocalFile after successful write
	newVersion.HasLocalFile = true
	if err := db.Save(&newVersion).Error; err != nil {
		return &FileUploadResponse{
			Success: false,
			Message: "Failed to update HasLocalFile",
		}, errors.Wrap(err, errors.DBQueryFailed, "Failed to update HasLocalFile")
	}

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
	var doc models.Document
	if err := db.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		return errors.Wrap(err, errors.DBRecordNotFound, "Document not found or access denied")
	}

	// Get full path
	fullPath, err := doc.GetFullPath()
	if err != nil {
		return errors.Wrap(err, errors.FSFileNotFound, "Failed to get file path")
	}

	// Delete file from disk
	if doc.FileExists() {
		if err := os.Remove(fullPath); err != nil {
			return errors.Wrap(err, errors.FSDeleteFailed, "Failed to delete file")
		}
	}

	// Delete all versions if this is the parent document
	if doc.ParentDocID == nil {
		if err := db.Where("parent_doc_id = ?", doc.ID).Delete(&models.Document{}).Error; err != nil {
			return errors.Wrap(err, errors.DBRecordNotFound, "Failed to delete document versions")
		}
	}

	// Delete document record
	if err := db.Delete(&doc).Error; err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Failed to delete document record")
	}

	// Update user storage info
	if err := models.UpdateStorageInfo(userID, db); err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Failed to update storage info")
	}

	return nil
}

// writeFile writes content to a file path
func WriteFile(filePath string, content io.Reader) error {
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, errors.FSCreateFailed, "Failed to create file")
	}
	defer file.Close()

	// Copy content to file
	_, err = io.Copy(file, content)
	if err != nil {
		return errors.Wrap(err, errors.FSWriteFailed, "Failed to write file content")
	}

	return nil
}
