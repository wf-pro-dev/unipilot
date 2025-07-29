package document

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"unipilot/internal/models/user"

	"gorm.io/gorm"
)

// DocumentType enum for different document categories
type DocumentType string

const (
	DocumentTypeSupport    DocumentType = "support"
	DocumentTypeSubmission DocumentType = "submission"
)

// Document represents a file attached to an assignment
type Document struct {
	gorm.Model
	AssignmentID uint         `gorm:"not null;index"`
	UserID       uint         `gorm:"not null;index"` // Original uploader
	Type         DocumentType `gorm:"not null;index"`
	FileName     string       `gorm:"not null"`
	FileType     string       `gorm:"not null"` // mime type or extension
	FilePath     string       `gorm:"not null"` // relative to app data directory
	FileSize     int64        `gorm:"not null"` // in bytes
	Version      int          `gorm:"default:1"`
	ParentDocID  *uint        `gorm:"index"`        // For version history
	IsOriginal   bool         `gorm:"default:true"` // For shared assignment tracking

	// Relationships
	User      user.User  `gorm:"foreignKey:UserID;references:ID"`
	ParentDoc *Document  `gorm:"foreignKey:ParentDocID;references:ID"`
	Versions  []Document `gorm:"foreignKey:ParentDocID;references:ID"`
}

// DocumentStorageInfo holds storage statistics
type DocumentStorageInfo struct {
	UserID           uint      `gorm:"primaryKey"`
	TotalSize        int64     `gorm:"default:0"` // Total bytes used by user
	DocumentCount    int       `gorm:"default:0"`
	LastCalculatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`

	User user.User `gorm:"foreignKey:UserID;references:ID"`
}

// Storage limits (in bytes)
const (
	MaxFileSize       = 50 * 1024 * 1024       // 50MB per file
	MaxAssignmentSize = 200 * 1024 * 1024      // 200MB per assignment
	MaxUserQuota      = 2 * 1024 * 1024 * 1024 // 2GB per user
)

// ValidateFileSize checks if file size is within limits
func (d *Document) ValidateFileSize(db *gorm.DB) error {
	// Check individual file size
	if d.FileSize > MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum of %d bytes", d.FileSize, MaxFileSize)
	}

	// Check assignment total size
	var assignmentTotal int64
	err := db.Model(&Document{}).
		Where("assignment_id = ? AND id != ?", d.AssignmentID, d.ID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&assignmentTotal).Error
	if err != nil {
		return fmt.Errorf("failed to calculate assignment storage: %w", err)
	}

	if assignmentTotal+d.FileSize > MaxAssignmentSize {
		return fmt.Errorf("assignment storage would exceed %d bytes limit", MaxAssignmentSize)
	}

	// Check user quota
	var userTotal int64
	err = db.Model(&Document{}).
		Where("user_id = ? AND id != ?", d.UserID, d.ID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&userTotal).Error
	if err != nil {
		return fmt.Errorf("failed to calculate user storage: %w", err)
	}

	if userTotal+d.FileSize > MaxUserQuota {
		return fmt.Errorf("user storage would exceed %d bytes quota", MaxUserQuota)
	}

	return nil
}

// GetAppDataPath returns the application data directory for file storage
func GetAppDataPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	appDataPath := filepath.Join(homeDir, ".unipilot", "documents")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(appDataPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data directory: %w", err)
	}

	return appDataPath, nil
}

// GenerateFilePath creates a unique file path for the document
func (d *Document) GenerateFilePath() (string, error) {
	appDataPath, err := GetAppDataPath()
	if err != nil {
		return "", err
	}

	// Create subdirectories: user_id/assignment_id/document_type/
	subDir := filepath.Join(
		fmt.Sprintf("user_%d", d.UserID),
		fmt.Sprintf("assignment_%d", d.AssignmentID),
		string(d.Type),
	)

	fullDir := filepath.Join(appDataPath, subDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create document directory: %w", err)
	}

	// Generate unique filename with timestamp to avoid conflicts
	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("%d_%s", timestamp, d.FileName)

	// Return relative path for storage in DB
	return filepath.Join(subDir, fileName), nil
}

// GetFullPath returns the absolute path to the document file
func (d *Document) GetFullPath() (string, error) {
	appDataPath, err := GetAppDataPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(appDataPath, d.FilePath), nil
}

// FileExists checks if the document file exists on disk
func (d *Document) FileExists() bool {
	fullPath, err := d.GetFullPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(fullPath)
	return err == nil
}

// CreateNewVersion creates a new version of an existing document
func (d *Document) CreateNewVersion(newFileName string, newFileSize int64, newFilePath string, db *gorm.DB) (*Document, error) {
	// Find the latest version
	var latestVersion int
	err := db.Model(&Document{}).
		Where("assignment_id = ? AND user_id = ? AND type = ? AND (id = ? OR parent_doc_id = ?)",
			d.AssignmentID, d.UserID, d.Type, d.ID, d.ID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&latestVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	// Create new version
	newVersion := &Document{
		AssignmentID: d.AssignmentID,
		UserID:       d.UserID,
		Type:         d.Type,
		FileName:     newFileName,
		FileType:     d.FileType,
		FilePath:     newFilePath,
		FileSize:     newFileSize,
		Version:      latestVersion + 1,
		ParentDocID:  &d.ID,
		IsOriginal:   d.IsOriginal,
	}

	// Validate before creating
	if err := newVersion.ValidateFileSize(db); err != nil {
		return nil, err
	}

	if err := db.Create(newVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to create new version: %w", err)
	}

	return newVersion, nil
}

// GetDocumentsByAssignment retrieves all documents for an assignment
func GetDocumentsByAssignment(assignmentID, userID uint, db *gorm.DB) ([]Document, error) {
	var documents []Document
	err := db.Preload("User").
		Preload("ParentDoc").
		Where("assignment_id = ? AND user_id = ?", assignmentID, userID).
		Order("type ASC, created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	return documents, nil
}

// GetLatestVersions returns only the latest version of each document
func GetLatestVersions(assignmentID, userID uint, db *gorm.DB) ([]Document, error) {
	var documents []Document

	// Get documents that are either original (no parent) or latest versions
	err := db.Preload("User").
		Where(`assignment_id = ? AND user_id = ? AND (
			parent_doc_id IS NULL OR 
			version = (
				SELECT MAX(version) 
				FROM documents d2 
				WHERE d2.parent_doc_id = documents.parent_doc_id OR d2.id = documents.parent_doc_id
			)
		)`, assignmentID, userID).
		Order("type ASC, created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get latest versions: %w", err)
	}

	return documents, nil
}

// UpdateStorageInfo recalculates and updates user storage statistics
func UpdateStorageInfo(userID uint, db *gorm.DB) error {
	var totalSize int64
	var documentCount int64

	// Calculate totals
	err := db.Model(&Document{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(file_size), 0), COUNT(*)").
		Row().Scan(&totalSize, &documentCount)
	if err != nil {
		return fmt.Errorf("failed to calculate storage info: %w", err)
	}

	// Update or create storage info
	storageInfo := &DocumentStorageInfo{
		UserID:           userID,
		TotalSize:        totalSize,
		DocumentCount:    int(documentCount),
		LastCalculatedAt: time.Now(),
	}

	err = db.Save(storageInfo).Error
	if err != nil {
		return fmt.Errorf("failed to update storage info: %w", err)
	}

	return nil
}
