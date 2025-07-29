package main

import (
	"fmt"
	"os"

	"unipilot/internal/models/assignment"
	"unipilot/internal/models/document"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 Starting REMOTE database migration for documents...")

	// Read database configuration
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("❌ Failed to read config file: %v\n", err)
		fmt.Println("💡 Make sure .env file exists with DATABASE_URL")
		os.Exit(1)
	}

	// Get database URL
	databaseURL := viper.GetString("DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("❌ DATABASE_URL not found in config")
		fmt.Println("💡 Add DATABASE_URL=your_postgres_url to .env file")
		os.Exit(1)
	}

	fmt.Println("🔗 Connecting to remote database...")

	// Connect to PostgreSQL database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Connected to remote database")

	// Check if migration is needed
	hasDocuments := db.Migrator().HasTable("documents")
	hasStorageInfo := db.Migrator().HasTable("document_storage_infos")

	if hasDocuments && hasStorageInfo {
		fmt.Println("✅ Document tables already exist, checking Assignment table...")
	} else {
		fmt.Println("📋 Running remote database migration...")

		// 1. Create new document tables
		fmt.Println("  → Creating document tables...")
		err = db.AutoMigrate(
			&document.Document{},
			&document.DocumentStorageInfo{},
		)
		if err != nil {
			fmt.Printf("❌ Failed to create document tables: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("  ✅ Document tables created")
	}

	// 2. Update Assignment table to add Documents relationship
	fmt.Println("  → Updating Assignment table...")
	err = db.AutoMigrate(&assignment.Assignment{})
	if err != nil {
		fmt.Printf("❌ Failed to update Assignment table: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✅ Assignment table updated")

	// 3. Verify the migration
	fmt.Println("  → Verifying migration...")

	// Check if tables exist
	tables := []string{"documents", "document_storage_infos", "assignments"}
	for _, tableName := range tables {
		if !db.Migrator().HasTable(tableName) {
			fmt.Printf("❌ Table %s does not exist\n", tableName)
			os.Exit(1)
		}
		fmt.Printf("  ✅ Table %s verified\n", tableName)
	}

	// Get table info for assignments to verify structure
	var count int64
	db.Table("assignments").Count(&count)
	fmt.Printf("  📊 Found %d existing assignments\n", count)

	fmt.Println("")
	fmt.Println("🎉 REMOTE database migration completed successfully!")
	fmt.Println("📄 Tables created/updated:")
	fmt.Println("   - documents (for document metadata)")
	fmt.Println("   - document_storage_infos (for user storage tracking)")
	fmt.Println("🔗 assignments table updated with Documents relationship")
	fmt.Printf("📊 Database ready for %d existing assignments\n", count)
	fmt.Println("")
	fmt.Println("⚠️  This script can now be DELETED as the migration is complete.")
}
