package utils

import (
	"unipilot/internal/storage"

	"unipilot/internal/errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetUserDB() (*gorm.DB, error) {

	// Determine database path
	dbPath, err := GetDBPath()
	if err != nil {
		return nil, errors.Wrap(err, errors.FSFileNotFound, "Failed to get database path")
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		PrepareStmt: true, // Better performance
		Logger:      logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to open SQLite database")
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to set connection pool")
	}
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection

	// Initialize schema (including new document tables)
	if err := storage.InitializeSchema(db); err != nil {
		return nil, errors.Wrap(err, errors.DBQueryFailed, "Failed to initialize schema")
	}

	return db, nil
}

func GetUserDBWithID(userID uint) (*gorm.DB, error) {

	dbPath, err := GetDBPathWithID(userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.FSFileNotFound, "Failed to get database path")
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		PrepareStmt: true, // Better performance
		Logger:      logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to open SQLite database")
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to set connection pool")
	}
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection

	// Initialize schema (including new document tables)
	if err := storage.InitializeSchema(db); err != nil {
		return nil, errors.Wrap(err, errors.DBQueryFailed, "Failed to initialize schema")
	}

	return db, nil

}
