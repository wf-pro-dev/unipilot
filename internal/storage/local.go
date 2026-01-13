package storage

import (
	"unipilot/internal/errors"
	"unipilot/internal/models"

	"gorm.io/gorm"
)

func InitializeSchema(db *gorm.DB) error {
	// Run migrations
	err := db.AutoMigrate(
		&models.LocalCourse{},
		&models.LocalAssignment{},
		&models.LocalSync{},
		&models.LocalDocument{},
		&models.LocalNote{},
		&models.LocalAiMessage{},
	)

	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Failed to migrate schema")
	}

	return nil
}
