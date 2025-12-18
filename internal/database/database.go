package database

import (
	"unipilot/internal/errors"
	"unipilot/internal/models/user"
	"unipilot/internal/services/utils"

	"gorm.io/gorm"
)

// Database provides database operations for the exposed structs
type Database struct {
	db   *gorm.DB
	user *user.User
}

// NewDatabase creates a new database helper
func NewDatabase(user *user.User) (*Database, error) {
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
