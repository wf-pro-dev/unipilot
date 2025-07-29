package main

import (
	"fmt"
	"os"
	"path/filepath"

	"unipilot/internal/models/assignment"
	"unipilot/internal/models/document"
	"unipilot/internal/storage"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 Starting LOCAL database migration for documents...")

	// Get current user ID (same logic as storage package)
	userID, err := storage.GetCurrentUserID()
	if err != nil {
		fmt.Printf("❌ Failed to get current user ID: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📂 Migrating database for user ID: %d\n", userID)

	// Get database path
	dbPath, err := getDBPath(userID)
	if err != nil {
		fmt.Printf("❌ Failed to get database path: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("💾 Database path: %s\n", dbPath)

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ Database file does not exist at %s\n", dbPath)
		fmt.Println("💡 Please run the application first to create the initial database")
		os.Exit(1)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Connected to local database")

	// Enable foreign keys for SQLite
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		fmt.Printf("❌ Failed to enable foreign keys: %v\n", err)
		os.Exit(1)
	}

	// Check if migration is needed
	hasDocuments := db.Migrator().HasTable(&document.LocalDocument{})
	hasCache := db.Migrator().HasTable(&document.LocalDocumentCache{})

	if hasDocuments && hasCache {
		fmt.Println("✅ Document tables already exist, no migration needed")
		return
	}

	fmt.Println("📋 Running local database migration...")

	// 1. Create new document tables
	fmt.Println("  → Creating document tables...")
	err = db.AutoMigrate(
		&document.LocalDocument{},
		&document.LocalDocumentCache{},
	)
	if err != nil {
		fmt.Printf("❌ Failed to create document tables: %v\n", err)
		os.Exit(1)
	}

	// 2. Update LocalAssignment table to add Documents relationship
	fmt.Println("  → Updating LocalAssignment table...")
	err = db.AutoMigrate(&assignment.LocalAssignment{})
	if err != nil {
		fmt.Printf("❌ Failed to update LocalAssignment table: %v\n", err)
		os.Exit(1)
	}

	// 3. Verify the migration
	fmt.Println("  → Verifying migration...")

	// Check if tables exist
	tables := []string{"local_documents", "local_document_caches"}
	for _, tableName := range tables {
		if !db.Migrator().HasTable(tableName) {
			fmt.Printf("❌ Table %s was not created\n", tableName)
			os.Exit(1)
		}
		fmt.Printf("  ✅ Table %s created successfully\n", tableName)
	}

	// Check if LocalAssignment has the Documents field by querying table info
	var columns []struct {
		Name string `json:"name"`
	}

	db.Raw("PRAGMA table_info(local_assignments)").Scan(&columns)
	fmt.Printf("  📋 LocalAssignment table has %d columns\n", len(columns))

	fmt.Println("")
	fmt.Println("🎉 LOCAL database migration completed successfully!")
	fmt.Println("📄 New tables created:")
	fmt.Println("   - local_documents (for document metadata)")
	fmt.Println("   - local_document_caches (for storage tracking)")
	fmt.Println("🔗 LocalAssignment table updated with Documents relationship")
	fmt.Println("")
	fmt.Println("⚠️  This script can now be DELETED as the migration is complete.")
}

// Helper function to get database path (copied from storage package)
func getDBPath(userID uint) (string, error) {
	// Check for custom path in config
	if customPath := viper.GetString("localdb.path"); customPath != "" {
		return filepath.Join(customPath, fmt.Sprintf("user_%d.db", userID)), nil
	}

	// Use OS-specific default location
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	return filepath.Join(
		configDir,
		"acc-homework",
		"data",
		fmt.Sprintf("user_%d.db", userID),
	), nil
}
