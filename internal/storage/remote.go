package storage

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "public.",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}

	return db, nil
}

