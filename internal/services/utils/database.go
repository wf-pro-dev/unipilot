package utils

import (
	"fmt"
	"unipilot/internal/storage"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetUserDB() (*gorm.DB, error) {

	// Determine database path
	dbPath, err := GetDBPath()
	if err != nil {
		return nil, err
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		PrepareStmt: true, // Better performance
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s: %w", dbPath, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection

	// Initialize schema (including new document tables)
	if err := storage.InitializeSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func GetUserDBWithID(userID uint) (*gorm.DB, error) {

	dbPath, err := GetDBPathWithID(userID)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		PrepareStmt: true, // Better performance
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s: %w", dbPath, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection

	// Initialize schema (including new document tables)
	if err := storage.InitializeSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil

}
