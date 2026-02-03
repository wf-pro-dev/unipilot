package models

import (
	"unipilot/internal/errors"

	"gorm.io/gorm"
)

// MigrateDocuments runs the database migrations for document models on REMOTE database
func MigrateDocuments(db *gorm.DB) error {
	// Auto-migrate the new document models for remote storage
	// This should be run on the server database
	err := db.AutoMigrate(
		&Document{},
		&DocumentStorage{},
	)

	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Failed to migrate document models")
	}

	return nil
}

// MigrateLocalDocuments runs the database migrations for local document models
func MigrateLocalDocuments(db *gorm.DB) error {
	// Auto-migrate the local document models for desktop app
	// This is called automatically by InitializeSchema in storage/local.go
	err := db.AutoMigrate(
		&LocalDocument{},
	)

	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Failed to migrate local document models")
	}

	return nil
}

// CheckDocumentMigrationNeeded checks if document migration is needed
func CheckDocumentMigrationNeeded(db *gorm.DB) bool {
	// Check if the documents table exists on remote database
	return !db.Migrator().HasTable(&Document{})
}

// CheckLocalDocumentMigrationNeeded checks if local document migration is needed
func CheckLocalDocumentMigrationNeeded(db *gorm.DB) bool {
	// Check if the local_documents table exists on local database
	return !db.Migrator().HasTable(&LocalDocument{})
}

// MigrateFollows runs the database migrations for follow models
func MigrateFollows(db *gorm.DB) error {
	// Auto-migrate the follow models
	err := db.AutoMigrate(
		&Friendship{},
	)

	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Failed to migrate follow models")
	}

	return nil
}

// CheckFollowMigrationNeeded checks if follow migration is needed
func CheckFollowMigrationNeeded(db *gorm.DB) bool {
	// Check if the follows table exists
	return !db.Migrator().HasTable(&Friendship{})
}
