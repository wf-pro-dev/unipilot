package fileops

import (
	"context"
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

func PickFile(ctx context.Context) (string, error) {

	// Open file dialog
	filters := []runtime.FileFilter{
		{
			DisplayName: "Documents",
			Pattern:     "*.pdf;*.doc;*.docx;*.ppt;*.pptx;*.xls;*.xlsx;*.txt;*.md;*.html",
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

func GetFilePath(doc *models.LocalDocument) string {
	userDir, err := utils.GetUserDir()
	if err != nil {
		return ""
	}
	assignmentDir := filepath.Join(userDir, "assignments", doc.AssignmentID)
	documentDir := filepath.Join(assignmentDir, "documents", doc.FileName)
	return filepath.Join(documentDir, doc.FileName)
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
func WriteDocument(document *models.LocalDocument, db *gorm.DB) error {

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(document.FilePath), 0755); err != nil {

		return errors.Wrap(err, errors.FSDirFailed, "Failed to create directory")
	}

	// Write file to disk
	if err := WriteFile(document.FilePath, document.FileContent); err != nil {
		// Clean up database record
		db.Delete(&document)
		return errors.Wrap(err, errors.FSWriteFailed, "Failed to write file")
	}

	// Update HasLocalFile to true after successful write
	document.HasLocalFile = true
	if err := db.Model(&document).Updates(map[string]interface{}{
		"has_local_file": true,
	}).Error; err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Failed to update HasLocalFile")
	}

	return nil
}

// DeleteDocument removes a document and its file
func DeleteDocument(docID, userID string, db *gorm.DB) error {
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
