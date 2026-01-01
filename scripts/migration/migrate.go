package main

import (
	"log"
	"unipilot/internal/services/database"
)

func main() {
	log.Println("Running migrations for unipilot server database...")

	// Initialize server database with migrations
	// This uses the database service which handles all migrations automatically
	db, err := database.InitializeServerDB(nil)
	if err != nil {
		log.Fatalf("Failed to initialize server database: %v", err)
	}

	// Close the database connection when done
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying database connection: %v", err)
	}
	defer sqlDB.Close()

	log.Println("✅ All server database migrations completed successfully!")
}

// For uuid type in postgresql, you need to create the extension first:
// CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
