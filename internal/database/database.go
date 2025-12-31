package database

import (
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/services/utils"

	"gorm.io/gorm"
)

// Database provides database operations for the exposed structs
type Database struct {
	db   *gorm.DB
	user *models.User
}

// NewDatabase creates a new database helper
func NewDatabase(user *models.User) (*Database, error) {
	database := &Database{
		user: user,
	}

	gormDB, err := utils.GetUserDB()
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to get user database")
	}

	database.db = gormDB

	return database, nil
}

func (h *Database) SetDB(db *gorm.DB) {
	h.db = db
}
