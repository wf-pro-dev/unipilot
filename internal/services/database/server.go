package database

import (
	"fmt"
	"log"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/storage"

	"gorm.io/gorm"
)

// ServerMigrator handles migrations for PostgreSQL server databases
type ServerMigrator struct {
	*BaseMigrator
}

// NewServerMigrator creates a new server database migrator
func NewServerMigrator() *ServerMigrator {
	models := []interface{}{
		&models.User{},
		&models.Course{},
		&models.Assignment{},
		&models.Note{},
		&models.Document{},
		&models.DocumentStorage{},
		&models.Follow{},
		&models.Device{},
		&models.CourseLinkRequest{},
	}

	return &ServerMigrator{
		BaseMigrator: NewBaseMigrator("ServerDB", models),
	}
}

// InitializeServerDB initializes a server database connection and runs migrations
// This is the main entry point for server database setup
func InitializeServerDB(config *DatabaseConfig) (*gorm.DB, error) {
	if config == nil {
		config = DefaultServerConfig()
	}

	log.Println("[ServerDB] Initializing server database...")

	// Get database connection
	db, err := storage.GetRemoteDB()
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to connect to server database")
	}

	// Validate connection if requested
	if config.ValidateConn {
		if err := validateServerConnection(db); err != nil {
			return nil, errors.Wrap(err, errors.DBConnectionFailed, "Server database connection validation failed")
		}
	}

	// Run migrations if requested
	if config.AutoMigrate {
		migrator := NewServerMigrator()
		if err := migrator.Migrate(db); err != nil {
			return nil, errors.Wrap(err, errors.DBQueryFailed, "Server database migration failed")
		}

		// Run additional model-specific migrations
		if err := runAdditionalServerMigrations(db); err != nil {
			return nil, errors.Wrap(err, errors.DBQueryFailed, "Additional server migrations failed")
		}
	}

	log.Println("[ServerDB] ✅ Server database initialized successfully")
	return db, nil
}

// MigrateServerDB runs migrations on an existing server database connection
func MigrateServerDB(db *gorm.DB) error {
	migrator := NewServerMigrator()
	if err := migrator.Migrate(db); err != nil {
		return err
	}

	// Run additional model-specific migrations
	return runAdditionalServerMigrations(db)
}

// validateServerConnection validates the server database connection
func validateServerConnection(db *gorm.DB) error {
	migrator := NewServerMigrator()
	return migrator.ValidateConnection(db)
}

// GetServerModels returns all models used in server database migrations
func GetServerModels() []interface{} {
	migrator := NewServerMigrator()
	return migrator.GetModels()
}

// CheckServerMigrationNeeded checks if server database migration is needed
func CheckServerMigrationNeeded(db *gorm.DB) (bool, []string) {
	migrator := NewServerMigrator()
	missingTables := []string{}

	for _, model := range migrator.GetModels() {
		if !CheckTableExists(db, model) {
			missingTables = append(missingTables, fmt.Sprintf("%T", model))
		}
	}

	// Check additional migrations
	if models.CheckDocumentMigrationNeeded(db) {
		missingTables = append(missingTables, "Document/DocumentStorage")
	}
	if models.CheckFollowMigrationNeeded(db) {
		missingTables = append(missingTables, "Follow")
	}

	return len(missingTables) > 0, missingTables
}

// runAdditionalServerMigrations runs model-specific migrations that may have custom logic
func runAdditionalServerMigrations(db *gorm.DB) error {
	// Migrate documents if needed
	if models.CheckDocumentMigrationNeeded(db) {
		log.Println("[ServerDB] Running document migrations...")
		if err := models.MigrateDocuments(db); err != nil {
			return fmt.Errorf("document migration failed: %w", err)
		}
	}

	// Migrate follows if needed
	if models.CheckFollowMigrationNeeded(db) {
		log.Println("[ServerDB] Running follow migrations...")
		if err := models.MigrateFollows(db); err != nil {
			return fmt.Errorf("follow migration failed: %w", err)
		}
	}

	return nil
}
