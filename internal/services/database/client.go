package database

import (
	"fmt"
	"log"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/services/utils"

	"gorm.io/gorm"
)

// ClientMigrator handles migrations for SQLite client databases
type ClientMigrator struct {
	*BaseMigrator
}

// NewClientMigrator creates a new client database migrator
func NewClientMigrator() *ClientMigrator {
	models := []interface{}{
		&models.LocalCourse{},
		&models.LocalAssignment{},
		&models.LocalSync{},
		&models.LocalDocument{},
		&models.LocalNote{},
		&models.LocalNotification{},
		&models.LocalAiMessage{},
	}

	return &ClientMigrator{
		BaseMigrator: NewBaseMigrator("ClientDB", models),
	}
}

// InitializeClientDBWithID initializes a client database for a specific user ID
func InitializeClient(userID uint, config *DatabaseConfig) (*gorm.DB, error) {
	if config == nil {
		config = DefaultClientConfig()
	}

	log.Printf("[ClientDB] Initializing client database for user ID: %d...", userID)

	// Get database connection for specific user
	db, err := utils.GetUserDBWithID(userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to connect to client database")
	}

	// Validate connection if requested
	if config.ValidateConn {
		if err := validateClientConnection(db); err != nil {
			return nil, errors.Wrap(err, errors.DBConnectionFailed, "Client database connection validation failed")
		}
	}

	// Note: GetUserDBWithID already runs InitializeSchema, so we don't need to migrate again
	// But we can validate that migrations were successful
	if config.AutoMigrate {
		migrator := NewClientMigrator()
		// Verify all tables exist
		for _, model := range migrator.GetModels() {
			log.Printf("[ClientDB] Checking table for model %T", model)
			if !CheckTableExists(db, model) {
				log.Printf("[ClientDB] Warning: Table for model %T does not exist, running migration...", model)
				if err := migrator.Migrate(db); err != nil {
					return nil, errors.Wrap(err, errors.DBQueryFailed, "Client database migration failed")
				}
				break
			}
		}
	}

	log.Printf("[ClientDB] ✅ Client database initialized successfully for user ID: %d", userID)
	return db, nil
}

// MigrateClientDB runs migrations on an existing client database connection
func MigrateClientDB(db *gorm.DB) error {
	migrator := NewClientMigrator()
	return migrator.Migrate(db)
}

// validateClientConnection validates the client database connection
func validateClientConnection(db *gorm.DB) error {
	migrator := NewClientMigrator()
	return migrator.ValidateConnection(db)
}

// GetClientModels returns all models used in client database migrations
func GetClientModels() []interface{} {
	migrator := NewClientMigrator()
	return migrator.GetModels()
}

// CheckClientMigrationNeeded checks if client database migration is needed
func CheckClientMigrationNeeded(db *gorm.DB) (bool, []string) {
	migrator := NewClientMigrator()
	missingTables := []string{}

	for _, model := range migrator.GetModels() {
		if !CheckTableExists(db, model) {
			missingTables = append(missingTables, fmt.Sprintf("%T", model))
		}
	}

	return len(missingTables) > 0, missingTables
}

// EnsureClientDBInitializedWithUser ensures the client database is initialized for a specific user
func EnsureClientDBInitialized(userID uint) (*gorm.DB, error) {
	log.Printf("[ClientDB] Ensuring client database is initialized for user ID: %d...", userID)

	// Initialize database for the user (creates DB file and runs migrations if needed)
	db, err := InitializeClient(userID, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to ensure client database is initialized")
	}

	log.Printf("[ClientDB] ✅ Client database ready for user ID: %d", userID)
	return db, nil
}
