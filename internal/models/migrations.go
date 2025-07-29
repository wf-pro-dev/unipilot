package models

import (
	"unipilot/internal/models/document"

	"gorm.io/gorm"
)

// MigrateDocuments runs the database migrations for document models
func MigrateDocuments(db *gorm.DB) error {
	// Auto-migrate the new document models
	// This is safe because it only adds new tables, doesn't modify existing ones
	err := db.AutoMigrate(
		&document.Document{},
		&document.DocumentStorageInfo{},
	)

	if err != nil {
		return err
	}

	return nil
}

// CheckDocumentMigrationNeeded checks if document migration is needed
func CheckDocumentMigrationNeeded(db *gorm.DB) bool {
	// Check if the documents table exists
	return !db.Migrator().HasTable(&document.Document{})
}
