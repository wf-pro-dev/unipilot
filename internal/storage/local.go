package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	dbLock      sync.Mutex
	dbInstances = make(map[uint]*gorm.DB)
)

func GetLocalDB() (*gorm.DB, uint, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	userID, err := GetCurrentUserID()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get current user ID: %w", err)
	}

	// Return cached instance if available
	if db, exists := dbInstances[userID]; exists {
		return db, userID, nil
	}

	// Determine database path
	dbPath, err := getDBPath(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get DB path: %w", err)
	}

	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, 0, fmt.Errorf("failed to create DB directory %s: %w", dbDir, err)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		PrepareStmt: true, // Better performance
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open SQLite database at %s: %w", dbPath, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get SQL DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection

	// Initialize schema (including new document tables)
	if err := InitializeSchema(db); err != nil {
		return nil, 0, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Cache the instance
	dbInstances[userID] = db

	return db, userID, nil
}

func getDBPath(userID uint) (string, error) {
	// Check for custom path in config
	if customPath := viper.GetString("localdb.path"); customPath != "" {
		return filepath.Join(customPath, fmt.Sprintf("user_%d.db", userID)), nil
	}

	// Use OS-specific default location
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	return filepath.Join(
		configDir,
		"acc-homework",
		"data",
		fmt.Sprintf("user_%d.db", userID),
	), nil
}

func InitializeSchema(db *gorm.DB) error {
	// Enable foreign key support for SQLite
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations
	err := db.AutoMigrate(
		&course.LocalCourse{},
		&models.LocalAssignmentType{},
		&models.LocalAssignmentStatus{},
		&assignment.LocalAssignment{},
		&models.LocalUpdate{},
		&document.LocalDocument{},
		&document.LocalDocumentCache{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate schema: %s", err)
	}

	types := []*models.LocalAssignmentType{
		{ID: 1, Name: "HW", Color: "yellow", NotionID: "Vn}Z"},
		{ID: 2, Name: "Exam", Color: "red", NotionID: "oiNS"},
	}

	// Assignment statuses
	statuses := []*models.LocalAssignmentStatus{
		{ID: 1, Name: "Not started", Color: "default", NotionID: "3aa77cf8-c39e-4c7b-b7d2-ab15ae43ff23"},
		{ID: 2, Name: "In progress", Color: "blue", NotionID: "97903420-1e83-4b3a-9eaf-a904354c968b"},
		{ID: 3, Name: "Done", Color: "green", NotionID: "2fef8044-d8d7-4fcf-a3ee-393a1d558e94"},
	}

	for _, t := range types {
		if err := db.Where("id = ?", t.ID).First(&models.LocalAssignmentType{}).Error; err != nil {
			err = db.Create(t).Error
			if err != nil {
				return err
			}
		}
	}

	for _, status := range statuses {
		if err := db.Where("id = ?", status.ID).First(&models.LocalAssignmentStatus{}).Error; err != nil {
			err = db.Create(status).Error
			if err != nil {
				return err
			}
		}
	}

	return nil
}
