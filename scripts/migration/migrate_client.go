package main

import (
	"log"
	"unipilot/internal/services/database"
)

func main() {
	log.Println("Running migrations for unipilot client database...")

	// Initialize client database with migrations
	// This uses the database service which handles all migrations automatically
	db, err := database.InitializeClientDB(nil)
	if err != nil {
		log.Fatalf("Failed to initialize client database: %v", err)
	}

	// Close the database connection when done
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying database connection: %v", err)
	}
	defer sqlDB.Close()

	log.Println("✅ All client database migrations completed successfully!")
}
