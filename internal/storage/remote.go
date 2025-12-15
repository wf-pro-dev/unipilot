package storage

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"unipilot/internal/secrets"
)

func GetRemoteDB() (*gorm.DB, error) {

	host, err := secrets.GetEnvVar("DB_HOST")
	port, err := secrets.GetEnvVar("DB_PORT")
	user, err := secrets.GetEnvVar("DB_USER")
	password, err := secrets.GetEnvVar("DB_PASSWORD")
	dbname, err := secrets.GetEnvVar("DB_NAME")

	if err != nil {
		return nil, err
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
		Logger: logger.Default.LogMode(logger.Silent), // Disable GORM's default logger
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}

	return db, nil
}
