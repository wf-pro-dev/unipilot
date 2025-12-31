package database

import (
	"gorm.io/gorm"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	// DatabaseTypeClient represents a SQLite client database
	DatabaseTypeClient DatabaseType = "client"
	// DatabaseTypeServer represents a PostgreSQL server database
	DatabaseTypeServer DatabaseType = "server"
)

// Migrator defines the interface for database migration operations
type Migrator interface {
	// Migrate runs all migrations for the database
	Migrate(db *gorm.DB) error
	// GetModels returns all models that should be migrated
	GetModels() []interface{}
	// ValidateConnection validates that the database connection is working
	ValidateConnection(db *gorm.DB) error
}

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	Success      bool
	ModelsCount  int
	ErrorMessage string
}

// DatabaseConfig holds configuration for database initialization
type DatabaseConfig struct {
	Type         DatabaseType
	AutoMigrate  bool
	ValidateConn bool
}

// DefaultClientConfig returns default configuration for client database
func DefaultClientConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:         DatabaseTypeClient,
		AutoMigrate:  true,
		ValidateConn: true,
	}
}

// DefaultServerConfig returns default configuration for server database
func DefaultServerConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:         DatabaseTypeServer,
		AutoMigrate:  true,
		ValidateConn: true,
	}
}
