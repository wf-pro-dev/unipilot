package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/qdrant/go-client/qdrant"
	"gorm.io/gorm"

	"unipilot/internal/errors"
	cloudstorage "unipilot/internal/services/cloud_storage"
)

// DocumentType enum for different document categories
type DocumentType string

const (
	DocumentTypeSupport    DocumentType = "support"
	DocumentTypeSubmission DocumentType = "submission"
	// SupportedFileTypes defines the allowed file extensions
)

var SupportedFileTypes = map[string]bool{

	// Binary files
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".ppt":  true,
	".pptx": true,
	".xls":  false,
	".xlsx": false,

	// Images
	".png":  false,
	".jpg":  false,
	".jpeg": false,
	".gif":  false,
	".bmp":  false,
	".svg":  false,

	// Text files
	".txt":  true,
	".md":   true,
	".cpp":  true,
	".java": true,
	".py":   true,
	".js":   true,
	".ts":   true,
	".html": true,
	".css":  true,
	".scss": true,
	".sass": true,
	".go":   true,
	".php":  true,
	".sh":   true,
}

type BaseDocument struct {
	Type         DocumentType `gorm:"not null;index" validate:"required,oneof=support submission"`
	FileName     string       `gorm:"not null" validate:"required,min=3,max=150"`
	FilePath     string       // relative to app data directory
	FileSize     int64        `gorm:"not null" validate:"required,min=1"` // in bytes
	StorageKey   *string      `gorm:"unique;default:null"`                // Only for remote storage
	Version      int          `gorm:"default:1" validate:"min=1"`
	ParentDocID  *string      `gorm:"index;default:null"`              // For version history
	IsOriginal   bool         `gorm:"default:true" validate:"boolean"` // For shared assignment tracking
	HasLocalFile bool         `gorm:"default:false" validate:"boolean"`

	AssignmentID string `gorm:"not null;index" validate:"required,min=1"`
}

// Document represents a file attached to an assignment
type Document struct {
	Base
	BaseDocument

	UserID string `gorm:"not null;index" validate:"required"`

	// Relationships
	User       *User       `gorm:"foreignKey:UserID;references:ID" validate:"-"`
	Assignment *Assignment `gorm:"foreignKey:AssignmentID;references:ID" validate:"-"`
	Parent     *Document   `gorm:"foreignKey:ParentDocID;references:ID" validate:"-"`
	Versions   []Document  `gorm:"foreignKey:ParentDocID;references:ID" validate:"-"`
}

// LocalDocument represents a document in the local database
type LocalDocument struct {
	Base
	BaseDocument
	SyncedAt *time.Time `gorm:"default:null"`

	// Local relationships
	Assignment LocalAssignment `gorm:"foreignKey:AssignmentID;references:ID" validate:"-"`
	Parent     *LocalDocument  `gorm:"foreignKey:ParentDocID;references:ID" validate:"-"`
	Versions   []LocalDocument `gorm:"foreignKey:ParentDocID;references:ID" validate:"-"`
}

// Hooks

func (d *Document) BeforeDelete(tx *gorm.DB) error {

	if d.HasLocalFile {
		// Delete the document on cloud
		if err := cloudstorage.DeleteFile(*d.StorageKey); err != nil {
			if errors.HasCode(err, errors.StorageFileNotFound) || errors.HasCode(err, errors.AuthForbidden) {
				return errors.Inherit(err, errors.StorageDeleteFailed)
			}
			return errors.Wrap(err, errors.StorageDeleteFailed, "Failed to delete document from storage")
		}
	}

	client, ok := tx.Get("qdrantClient")
	if !ok {
		return nil
	}
	qdrantClient, ok := client.(*qdrant.Client)
	if !ok {
		return nil
	}

	// Delete the document from Qdrant (non blocking)
	go DeleteDocumentVectors(d, qdrantClient)

	return nil
}

// START: Conversion Functions

func (d *Document) ToLocal() *LocalDocument {

	localDocument := &LocalDocument{
		Base:         d.Base,
		BaseDocument: d.BaseDocument,
	}
	return localDocument
}

// ToRemoteDocument converts local document to remote document format
func (ld *LocalDocument) ToRemote(userID string) *Document {
	baseDocument := ld.BaseDocument
	return &Document{
		Base:         ld.Base,
		BaseDocument: baseDocument,
		UserID:       userID,
	}
}

// END: Conversion Functions

// START: Validation Functions

func (bd *BaseDocument) Validate() error {

	bd.FileName = strings.TrimRight(bd.FileName, " ")
	bd.FileName = strings.TrimLeft(bd.FileName, " ")

	bd.FilePath = strings.TrimRight(bd.FilePath, " ")
	bd.FilePath = strings.TrimLeft(bd.FilePath, " ")

	if bd.StorageKey != nil {
		storageKey := *bd.StorageKey
		storageKey = strings.TrimRight(storageKey, " ")
		storageKey = strings.TrimLeft(storageKey, " ")
		bd.StorageKey = &storageKey
	}

	if err := ValidateFileType(bd.FileName); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Unsupported file type")
	}

	// Validate file size
	if bd.FileSize > MaxFileSize {

		return errors.Wrap(fmt.Errorf("file size exceeds limit of 50MB"), errors.ValidationInvalid, "File size exceeds limit")
	}

	if err := validator.New().Struct(bd); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "BaseDocument validation failed")
	}
	return nil
}

func (d *Document) Validate(db *gorm.DB) error {

	if err := d.BaseDocument.Validate(); err != nil {
		return err
	}
	if err := ValidateFileSize(d, db); err != nil {
		return err
	}
	if err := validator.New().Struct(d); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Document validation failed")
	}
	return nil
}

func (ld *LocalDocument) Validate(db *gorm.DB) error {
	if err := ld.BaseDocument.Validate(); err != nil {
		return err
	}
	if err := ValidateLocalFileSize(ld, db); err != nil {
		return err
	}

	if err := validator.New().Struct(ld); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "LocalDocument validation failed")
	}
	return nil
}

// ValidateFileType checks if the file extension is supported
func ValidateFileType(fileName string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	if _, ok := SupportedFileTypes[ext]; !ok {
		return errors.Wrap(fmt.Errorf("file type %s is not supported", ext), errors.FSFileTypeNotSupported, "File type not supported")
	}
	return nil
}

func ValidateFileTypeRAG(fileName string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	valid, ok := SupportedFileTypes[ext]
	if !ok {
		return errors.Wrap(fmt.Errorf("file type %s is not supported", ext), errors.FSFileTypeNotSupported, "File type not supported for RAG")
	}
	log.Println("valid", valid, "ext", ext)
	if !valid {
		return errors.Wrap(fmt.Errorf("file type %s is not supported for RAG", ext), errors.FSFileTypeNotSupported, "File type not supported for RAG")
	}
	return nil
}

// ValidateFileSize checks if file size is within limits
func ValidateFileSize(doc *Document, db *gorm.DB) error {
	// Check individual file size
	if doc.FileSize > MaxFileSize {
		return errors.NewAppError(errors.ValidationInvalid, "File size exceeds maximum of 50MB", nil)
	}

	// Check assignment total size
	var assignmentTotal int64
	err := db.Model(&Document{}).
		Where("assignment_id = ? AND id != ?", doc.AssignmentID, doc.ID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&assignmentTotal).Error
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	if assignmentTotal+doc.FileSize > MaxAssignmentSize {
		return errors.NewAppError(errors.ValidationInvalid, "Assignment storage would exceed 200MB limit", nil)
	}

	// Check user quota
	var userTotal int64
	err = db.Model(&Document{}).
		Where("user_id = ? AND id != ?", doc.UserID, doc.ID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&userTotal).Error
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	if userTotal+doc.FileSize > MaxUserQuota {
		return errors.NewAppError(errors.ValidationInvalid, "User storage would exceed 2GB quota", nil)
	}

	return nil
}

func ValidateLocalFileSize(doc *LocalDocument, db *gorm.DB) error {

	if doc.FileSize > MaxFileSize {
		return errors.NewAppError(errors.ValidationInvalid, "File size exceeds maximum of 50MB", nil)
	}

	// Check assignment total size
	var assignmentTotal int64
	err := db.Model(&LocalDocument{}).
		Where("assignment_id = ? AND id != ?", doc.AssignmentID, doc.ID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&assignmentTotal).Error
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	if assignmentTotal+doc.FileSize > MaxAssignmentSize {
		return errors.NewAppError(errors.ValidationInvalid, "Assignment storage would exceed 200MB limit", nil)
	}

	// Check user quota
	var userTotal int64
	err = db.Model(&LocalDocument{}).
		Where("id != ?", doc.ID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&userTotal).Error
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	if userTotal+doc.FileSize > MaxUserQuota {
		return errors.NewAppError(errors.ValidationInvalid, "User storage would exceed 2GB quota", nil)
	}

	return nil

}

// END: Validation Functions

// DocumentStorageInfo holds storage statistics
type DocumentStorage struct {
	UserID           string    `gorm:"primaryKey"`
	TotalSize        int64     `gorm:"default:0"` // Total bytes used by user
	DocumentCount    int       `gorm:"default:0"`
	LastCalculatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}

type LocalAssignmentStorage struct {
	AssignmentID     string
	TotalCount       int
	DocumentCount    int
	TotalSize        int64
	Size             int64
	LastCalculatedAt time.Time
}

// Storage limits (in bytes)
const (
	MaxFileSize       = 50 * 1024 * 1024       // 50MB per file
	MaxAssignmentSize = 200 * 1024 * 1024      // 200MB per assignment
	MaxUserQuota      = 2 * 1024 * 1024 * 1024 // 2GB per user
)

func (d *Document) AfterDelete(tx *gorm.DB) error {
	// Commit the transaction
	if err := UpdateStorageInfo(d.UserID, tx); err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

func GetDocument(docID string, db *gorm.DB) (*Document, error) {

	var doc Document
	if err := db.Where("id = ?", docID).First(&doc).Error; err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &doc, nil
}
func GetLDocument(docID string, db *gorm.DB) (*LocalDocument, error) {

	var doc LocalDocument
	if err := db.Where("id = ?", docID).First(&doc).Error; err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &doc, nil
}

// GetAppDataPath returns the application data directory for file storage
func GetAppDataPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to get home directory")
	}

	appDataPath := filepath.Join(homeDir, ".unipilot", "documents")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(appDataPath, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirFailed, "Failed to create app data directory")
	}

	return appDataPath, nil
}

// GenerateFilePath creates a unique file path for the document
func (d *Document) GenerateFilePath() (string, error) {
	appDataPath, err := GetAppDataPath()
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to get app data path")
	}

	// Create subdirectories: user_id/assignment_id/document_type/
	subDir := filepath.Join(
		fmt.Sprintf("user_%d", d.UserID),
		fmt.Sprintf("assignment_%d", d.AssignmentID),
		string(d.Type),
	)

	fullDir := filepath.Join(appDataPath, subDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirFailed, "Failed to create document directory")
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
		return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to get app data path")
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
		return nil, errors.HandleDBReadError(err)
	}

	// Create new version
	baseD := &BaseDocument{
		Type:         d.Type,
		FileName:     d.FileName,
		FilePath:     d.FilePath,
		FileSize:     d.FileSize,
		Version:      latestVersion + 1,
		ParentDocID:  &d.ID,
		IsOriginal:   d.IsOriginal,
		HasLocalFile: d.HasLocalFile,
		AssignmentID: d.AssignmentID,
	}
	newVersion := &Document{
		BaseDocument: *baseD,
		UserID:       d.UserID,
	}

	// Validate before creating
	if err := ValidateFileSize(newVersion, db); err != nil {
		return nil, errors.Inherit(err, errors.FSFileTooLarge)
	}

	if err := db.Create(newVersion).Error; err != nil {
		return nil, errors.HandleDBCreateError(err)
	}

	return newVersion, nil
}

// GetDocumentsByAssignment retrieves all documents for an assignment
func (a *Assignment) GetDocumentsByAssignment(db *gorm.DB) ([]Document, error) {
	var documents []Document
	err := db.Where("assignment_id = ?", a.ID).
		Order("type ASC, created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}

	return documents, nil
}

func (ld *LocalAssignment) GetDocumentsByAssignment(db *gorm.DB) ([]LocalDocument, error) {
	var documents []LocalDocument
	err := db.Where("assignment_id = ?", ld.ID).
		Order("type ASC, created_at DESC").
		Find(&documents).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
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
		return nil, errors.HandleDBReadError(err)
	}

	return documents, nil
}

// GetUserStorageInfo returns storage statistics for a user
func GetUserStorageInfo(userID string, db *gorm.DB) (*DocumentStorage, error) {
	var storageInfo DocumentStorage
	err := db.Where("user_id = ?", userID).First(&storageInfo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new storage info if it doesn't exist
			if err := UpdateStorageInfo(userID, db); err != nil {
				return nil, errors.Wrap(err, errors.DBRecordNotFound, "Document storage info not found")
			}
			// Try again
			err = db.Where("user_id = ?", userID).First(&storageInfo).Error
		}
		if err != nil {
			return nil, errors.Wrap(err, errors.DBQueryFailed, "Failed to get storage info")
		}
	}

	return &storageInfo, nil
}

func GetLocalStorageInfo(db *gorm.DB) (*DocumentStorage, error) {
	var totalSize int64
	var documentCount int64

	err := db.Model(&LocalDocument{}).
		Select("COALESCE(SUM(file_size), 0), COUNT(*)").
		Where("has_local_file = ?", true).
		Row().Scan(&totalSize, &documentCount)
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &DocumentStorage{
		TotalSize:        totalSize,
		DocumentCount:    int(documentCount),
		LastCalculatedAt: time.Now(),
	}, nil
}

func GetLocalAssignmentStorage(assignmentID string, db *gorm.DB) (*LocalAssignmentStorage, error) {
	var size int64
	var documentCount int64

	totalStorage, err := GetLocalStorageInfo(db)
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}

	err = db.Model(&LocalDocument{}).
		Where("assignment_id = ?", assignmentID).
		Select("COALESCE(SUM(file_size), 0), COUNT(*)").
		Where("has_local_file = ?", true).
		Row().Scan(&size, &documentCount)
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}

	return &LocalAssignmentStorage{
		AssignmentID:     assignmentID,
		TotalCount:       totalStorage.DocumentCount,
		DocumentCount:    int(documentCount),
		TotalSize:        totalStorage.TotalSize,
		Size:             size,
		LastCalculatedAt: time.Now(),
	}, nil
}

// UpdateStorageInfo recalculates and updates user storage statistics
func UpdateStorageInfo(userID string, db *gorm.DB) error {
	var totalSize int64
	var documentCount int64

	// Calculate totals
	err := db.Model(&Document{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(file_size), 0), COUNT(*)").
		Row().Scan(&totalSize, &documentCount)
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	// Update or create storage info
	storageInfo := &DocumentStorage{
		UserID:           userID,
		TotalSize:        totalSize,
		DocumentCount:    int(documentCount),
		LastCalculatedAt: time.Now(),
	}

	err = db.Save(storageInfo).Error
	if err != nil {
		return errors.HandleDBWriteError(err)

	}

	return nil
}

func DeleteDocumentVectors(doc *Document, qdrantClient *qdrant.Client) error {
	// Step 2: Delete the document from the Qdrant
	if _, err := qdrantClient.Delete(context.Background(), &qdrant.DeletePoints{
		CollectionName: fmt.Sprintf("unipilot-qdrant-db-%s", doc.AssignmentID),
		Points: qdrant.NewPointsSelectorFilter(
			&qdrant.Filter{
				Must: []*qdrant.Condition{
					qdrant.NewMatch("document_id", doc.ID),
				},
			},
		),
	}); err != nil {
		return errors.WrapServer(
			err,
			errors.QdrantDeletePointsError,
			"Error deleting document from Qdrant",
			fiber.StatusInternalServerError,
		)
	}

	return nil
}
