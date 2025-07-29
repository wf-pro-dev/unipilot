package document

import (
	"time"

	"gorm.io/gorm"
)

// LocalDocument represents a document in the local database
// This is for local operations and caching remote metadata
type LocalDocument struct {
	gorm.Model
	RemoteID     uint         `gorm:"index"` // Reference to remote document ID
	AssignmentID uint         `gorm:"not null;index"`
	UserID       uint         `gorm:"not null;index"` // Original uploader
	Type         DocumentType `gorm:"not null;index"`
	FileName     string       `gorm:"not null"`
	FileType     string       `gorm:"not null"`
	FilePath     string       // Local file path (only if we have the file)
	FileSize     int64        `gorm:"not null"`
	Version      int          `gorm:"default:1"`
	ParentDocID  *uint        `gorm:"index"`
	IsOriginal   bool         `gorm:"default:true"`
	HasLocalFile bool         `gorm:"default:false"` // Do we have the actual file locally?
	LastSyncAt   *time.Time   // When we last synced metadata from remote

	// Local relationships
	ParentDoc *LocalDocument  `gorm:"foreignKey:ParentDocID;references:ID"`
	Versions  []LocalDocument `gorm:"foreignKey:ParentDocID;references:ID"`
}

// LocalDocumentCache represents cached document storage info
type LocalDocumentCache struct {
	gorm.Model
	UserID           uint       `gorm:"not null;unique"`
	TotalSize        int64      `gorm:"default:0"`
	DocumentCount    int        `gorm:"default:0"`
	LastCalculatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	LastSyncAt       *time.Time // When we last synced from remote
}

// ToRemoteDocument converts local document to remote document format
func (ld *LocalDocument) ToRemoteDocument() *Document {
	return &Document{
		Model:        ld.Model,
		AssignmentID: ld.AssignmentID,
		UserID:       ld.UserID,
		Type:         ld.Type,
		FileName:     ld.FileName,
		FileType:     ld.FileType,
		FilePath:     ld.FilePath,
		FileSize:     ld.FileSize,
		Version:      ld.Version,
		ParentDocID:  ld.ParentDocID,
		IsOriginal:   ld.IsOriginal,
	}
}

// FromRemoteDocument creates/updates local document from remote data
func (ld *LocalDocument) FromRemoteDocument(rd *Document, hasLocalFile bool) {
	ld.RemoteID = rd.ID
	ld.AssignmentID = rd.AssignmentID
	ld.UserID = rd.UserID
	ld.Type = rd.Type
	ld.FileName = rd.FileName
	ld.FileType = rd.FileType
	ld.FileSize = rd.FileSize
	ld.Version = rd.Version
	ld.ParentDocID = rd.ParentDocID
	ld.IsOriginal = rd.IsOriginal
	ld.HasLocalFile = hasLocalFile
	now := time.Now()
	ld.LastSyncAt = &now
}

// GetLocalDocumentsByAssignment retrieves all local documents for an assignment
func GetLocalDocumentsByAssignment(assignmentID uint, db *gorm.DB) ([]LocalDocument, error) {
	var documents []LocalDocument
	err := db.Where("assignment_id = ?", assignmentID).
		Order("type ASC, created_at DESC").
		Find(&documents).Error

	return documents, err
}

// SyncRemoteMetadata updates local document cache from remote metadata
func SyncRemoteMetadata(assignmentID uint, remoteDocuments []Document, db *gorm.DB) error {
	for _, remoteDoc := range remoteDocuments {
		var localDoc LocalDocument

		// Check if we already have this document locally
		err := db.Where("remote_id = ?", remoteDoc.ID).First(&localDoc).Error

		if err == gorm.ErrRecordNotFound {
			// Create new local record
			localDoc = LocalDocument{}
			localDoc.FromRemoteDocument(&remoteDoc, false) // Assume no local file unless we uploaded it

			// Check if this is our own document (we might have the file)
			// This would need to be determined by checking if file exists locally

			if err := db.Create(&localDoc).Error; err != nil {
				return err
			}
		} else if err == nil {
			// Update existing record
			localDoc.FromRemoteDocument(&remoteDoc, localDoc.HasLocalFile)
			if err := db.Save(&localDoc).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
