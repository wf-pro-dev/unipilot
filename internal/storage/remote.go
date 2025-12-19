package storage

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"unipilot/internal/secrets"

	"unipilot/internal/errors"
)

func GetRemoteDB() (*gorm.DB, error) {

	host, err := secrets.GetEnvVar("DB_HOST")
	if err != nil {
		return nil, errors.Wrap(err, errors.ConfigEnvVarNotFound, "cannot get the database host")
	}
	port, err := secrets.GetEnvVar("DB_PORT")
	if err != nil {
		return nil, errors.Wrap(err, errors.ConfigEnvVarNotFound, "cannot get the database port")
	}
	user, err := secrets.GetEnvVar("DB_USER")
	if err != nil {
		return nil, errors.Wrap(err, errors.ConfigEnvVarNotFound, "cannot get the database user")
	}
	password, err := secrets.GetEnvVar("DB_PASSWORD")
	if err != nil {
		return nil, errors.Wrap(err, errors.ConfigEnvVarNotFound, "cannot get the database password")
	}
	dbname, err := secrets.GetEnvVar("DB_NAME")
	if err != nil {
		return nil, errors.Wrap(err, errors.ConfigEnvVarNotFound, "cannot get the database name")
	}
	// Updated connection string with SSL
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Configure GORM with silent logger to avoid duplicate logging
	// Application-level logging is handled by our zap logger
	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "public.",
		},
		Logger: logger.Default.LogMode(logger.Error), // Disable GORM's default logger
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "cannot connect to the database")
	}

	return db, nil
}
