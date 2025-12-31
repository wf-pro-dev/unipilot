package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// BaseMigrator provides common migration functionality
type BaseMigrator struct {
	models []interface{}
	name   string
}

// NewBaseMigrator creates a new base migrator
func NewBaseMigrator(name string, models []interface{}) *BaseMigrator {
	return &BaseMigrator{
		name:   name,
		models: models,
	}
}

// Migrate runs migrations for all registered models
func (m *BaseMigrator) Migrate(db *gorm.DB) error {
	if len(m.models) == 0 {
		return fmt.Errorf("no models registered for migration")
	}

	log.Printf("[%s] Starting migration for %d models...", m.name, len(m.models))

	err := db.AutoMigrate(m.models...)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Printf("[%s] ✅ Migration completed successfully", m.name)
	return nil
}

// GetModels returns all registered models
func (m *BaseMigrator) GetModels() []interface{} {
	return m.models
}

// ValidateConnection validates the database connection by executing a simple query
func (m *BaseMigrator) ValidateConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// CheckTableExists checks if a table exists in the database
func CheckTableExists(db *gorm.DB, model interface{}) bool {
	return db.Migrator().HasTable(model)
}

// GetTableCount returns the number of tables in the database
func GetTableCount(db *gorm.DB) (int, error) {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Count(&count).Error
	if err != nil {
		// Try PostgreSQL syntax if SQLite fails
		err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Count(&count).Error
		if err != nil {
			return 0, err
		}
	}
	return int(count), nil
}
