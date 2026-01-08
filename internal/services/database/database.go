package database

import (
	"unipilot/internal/errors"
	"unipilot/internal/models"

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

	db, err := InitializeClient(user.ID, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to initialize client database")
	}

	database.db = db

	return database, nil
}

func (h *Database) SetDB(db *gorm.DB) { h.db = db }

func (h *Database) GetDB() *gorm.DB { return h.db }
