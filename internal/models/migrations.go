package models

import (
	"unipilot/internal/models/document"

	"gorm.io/gorm"
)

// MigrateDocuments runs the database migrations for document models on REMOTE database
func MigrateDocuments(db *gorm.DB) error {
	// Auto-migrate the new document models for remote storage
	// This should be run on the server database
	err := db.AutoMigrate(
		&document.Document{},
		&document.DocumentStorageInfo{},
	)

	if err != nil {
		return err
	}

	return nil
}

// MigrateLocalDocuments runs the database migrations for local document models
func MigrateLocalDocuments(db *gorm.DB) error {
	// Auto-migrate the local document models for desktop app
	// This is called automatically by InitializeSchema in storage/local.go
	err := db.AutoMigrate(
		&document.LocalDocument{},
		&document.LocalDocumentCache{},
	)

	if err != nil {
		return err
	}

	return nil
}

// CheckDocumentMigrationNeeded checks if document migration is needed
func CheckDocumentMigrationNeeded(db *gorm.DB) bool {
	// Check if the documents table exists on remote database
	return !db.Migrator().HasTable(&document.Document{})
}

// CheckLocalDocumentMigrationNeeded checks if local document migration is needed
func CheckLocalDocumentMigrationNeeded(db *gorm.DB) bool {
	// Check if the local_documents table exists on local database
	return !db.Migrator().HasTable(&document.LocalDocument{})
}
