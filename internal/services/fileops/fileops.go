package fileops

import (
	"context"
	"io"
	"log"
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

	log.Println("Writing document", document.ID, document.FileSize)

	// Open file content
	fileContent, err := os.Open(document.FilePath)
	if err != nil {
		return errors.Wrap(err, errors.FSOpenFailed, "Failed to open file")
	}
	defer fileContent.Close()

	// Get new file path
	filePath := GetFilePath(document)
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {

		return errors.Wrap(err, errors.FSDirFailed, "Failed to create directory")
	}

	// Write file to disk
	if err := WriteFile(filePath, fileContent); err != nil {
		// Clean up database record
		db.Delete(&document)
		return errors.Wrap(err, errors.FSWriteFailed, "Failed to write file")
	}

	document.FilePath = filePath // Update file path

	log.Println("Document written", document.ID, document.FileSize)

	return nil
}

// writeFile writes content to a file path
func WriteFile(filePath string, content io.Reader) error {

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return errors.Wrap(err, errors.FSDirFailed, "Failed to create directory")
	}

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
