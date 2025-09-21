package database

import (
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
func NewDatabase(user *user.User) *Database {
	database := &Database{
		user: user,
	}

	gormDB, err := utils.GetUserDB()
	if err != nil {
		return nil
	}

	database.db = gormDB

	return database
}
