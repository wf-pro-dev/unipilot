package main

import (
	"fmt"
	"log"
	"os"

	"unipilot/internal/storage"

	"gorm.io/gorm"
)

type Course struct {
	ID        uint
	Code      string
	Name      string
	UserID    uint
	DeletedAt gorm.DeletedAt
}

func main() {
	db, err := storage.GetRemoteDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Connected to PostgreSQL database successfully")

	// Check if the new index already exists
	var indexExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE indexname = 'idx_courses_user_code_active'
		)
	`).Scan(&indexExists).Error

	if err != nil {
		log.Fatalf("Failed to check if index exists: %v", err)
	}

	if indexExists {
		fmt.Println("✅ Index 'idx_courses_user_code_active' already exists")
		return
	}

	// Check if the old unique constraint exists
	var constraintExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints 
			WHERE constraint_name = 'uni_public_courses_code_user' 
			AND table_name = 'courses'
		)
	`).Scan(&constraintExists).Error

	// Also check for other possible constraint names
	if !constraintExists {
		err = db.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.table_constraints 
				WHERE constraint_type = 'UNIQUE' 
				AND table_name = 'courses'
				AND constraint_name LIKE '%code%'
			)
		`).Scan(&constraintExists).Error
	}

	if err != nil {
		log.Fatalf("Failed to check if constraint exists: %v", err)
	}

	// Start transaction
	tx := db.Begin()
	if err := tx.Error; err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ Transaction rolled back due to panic: %v", r)
		}
	}()

	// Step 1: Drop the old unique constraint if it exists
	if constraintExists {
		fmt.Println("🔄 Dropping old unique constraint...")

		// Try to find the actual constraint name
		var constraintName string
		err = tx.Raw(`
			SELECT constraint_name 
			FROM information_schema.table_constraints 
			WHERE constraint_type = 'UNIQUE' 
			AND table_name = 'courses'
			AND constraint_name LIKE '%code%'
			LIMIT 1
		`).Scan(&constraintName).Error

		if err == nil && constraintName != "" {
			// First, find and drop any dependent objects
			fmt.Printf("🔄 Checking for dependent objects on constraint '%s'...\n", constraintName)

			// Find foreign keys that reference this constraint
			var foreignKeys []string
			err = tx.Raw(`
				SELECT tc.constraint_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu 
					ON tc.constraint_name = kcu.constraint_name
				WHERE tc.constraint_type = 'FOREIGN KEY'
				AND tc.table_name = 'courses'
				AND kcu.column_name IN ('user_id', 'code')
			`).Scan(&foreignKeys).Error

			if err == nil && len(foreignKeys) > 0 {
				fmt.Printf("🔄 Found %d dependent foreign keys, dropping them first...\n", len(foreignKeys))
				for _, fk := range foreignKeys {
					err = tx.Exec(fmt.Sprintf(`ALTER TABLE courses DROP CONSTRAINT IF EXISTS %s CASCADE`, fk)).Error
					if err != nil {
						fmt.Printf("⚠️  Warning: Failed to drop foreign key %s: %v\n", fk, err)
					} else {
						fmt.Printf("✅ Dropped foreign key: %s\n", fk)
					}
				}
			}

			// Now try to drop the constraint with CASCADE
			err = tx.Exec(fmt.Sprintf(`ALTER TABLE courses DROP CONSTRAINT IF EXISTS %s CASCADE`, constraintName)).Error
			if err != nil {
				// If CASCADE fails, check what depends on it
				checkDependentObjects(tx, constraintName)

				// Try without CASCADE
				err = tx.Exec(fmt.Sprintf(`ALTER TABLE courses DROP CONSTRAINT IF EXISTS %s`, constraintName)).Error
				if err != nil {
					tx.Rollback()
					log.Fatalf("Failed to drop old constraint %s: %v\n\n💡 Tip: You may need to manually drop dependent objects first.", constraintName, err)
				}
			}
			fmt.Printf("✅ Old unique constraint '%s' dropped\n", constraintName)
		} else {
			// Fallback to specific constraint names including the user's constraint
			constraintNames := []string{"uni_public_courses_code_user", "courses_code_key", "courses_code_unique", "idx_courses_code"}
			for _, name := range constraintNames {
				// Try with CASCADE first
				err = tx.Exec(fmt.Sprintf(`ALTER TABLE courses DROP CONSTRAINT IF EXISTS %s CASCADE`, name)).Error
				if err != nil {
					// If CASCADE fails, try without CASCADE
					err = tx.Exec(fmt.Sprintf(`ALTER TABLE courses DROP CONSTRAINT IF EXISTS %s`, name)).Error
				}
				if err == nil {
					fmt.Printf("✅ Old unique constraint '%s' dropped\n", name)
					break
				}
			}
		}
	}

	// Step 2: Create the new composite unique index (including user_id)
	fmt.Println("🔄 Creating new composite unique index...")
	err = tx.Exec(`
		CREATE UNIQUE INDEX idx_courses_user_code_active 
		ON courses(user_id, code) 
		WHERE deleted_at IS NULL
	`).Error
	if err != nil {
		tx.Rollback()
		log.Fatalf("Failed to create new index: %v", err)
	}
	fmt.Println("✅ New composite unique index created")

	// Step 3: Clean up soft-deleted courses with duplicate codes
	fmt.Println("🔄 Cleaning up soft-deleted courses with duplicate codes...")

	// Find all soft-deleted courses
	var softDeletedCourses []Course
	err = tx.Unscoped().Where("deleted_at IS NOT NULL").Find(&softDeletedCourses).Error
	if err != nil {
		tx.Rollback()
		log.Fatalf("Failed to query soft-deleted courses: %v", err)
	}

	// Group by code and keep only the most recently deleted one
	codeMap := make(map[string]*Course)
	for i := range softDeletedCourses {
		course := &softDeletedCourses[i]
		if existing, exists := codeMap[course.Code]; exists {
			// Keep the one that was deleted most recently
			if course.DeletedAt.Time.After(existing.DeletedAt.Time) {
				codeMap[course.Code] = course
			}
		} else {
			codeMap[course.Code] = course
		}
	}

	// Permanently delete all soft-deleted courses except the most recent one for each code
	deletedCount := 0
	for _, course := range softDeletedCourses {
		if mostRecent := codeMap[course.Code]; mostRecent != nil && mostRecent.ID != course.ID {
			err := tx.Unscoped().Delete(&Course{}, course.ID).Error
			if err != nil {
				tx.Rollback()
				log.Fatalf("Failed to permanently delete course %d: %v", course.ID, err)
			}
			deletedCount++
		}
	}

	if deletedCount > 0 {
		fmt.Printf("✅ Permanently deleted %d duplicate soft-deleted courses\n", deletedCount)
	} else {
		fmt.Println("✅ No duplicate soft-deleted courses found")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("")
	fmt.Println("📋 Summary:")
	fmt.Println("  - Old unique constraint removed")
	fmt.Println("  - New composite unique index created (user_id, code)")
	fmt.Printf("  - %d duplicate soft-deleted courses cleaned up\n", deletedCount)
	fmt.Println("")
	fmt.Println("🎉 Users can now reuse course codes after deletion!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// checkDependentObjects checks what objects depend on a constraint
func checkDependentObjects(tx *gorm.DB, constraintName string) {
	fmt.Printf("🔍 Checking what depends on constraint '%s'...\n", constraintName)

	// Check for foreign keys
	var foreignKeys []string
	err := tx.Raw(`
		SELECT tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND tc.table_name = 'courses'
		AND kcu.column_name IN ('user_id', 'code')
	`).Scan(&foreignKeys).Error

	if err == nil && len(foreignKeys) > 0 {
		fmt.Printf("📋 Found foreign keys: %v\n", foreignKeys)
	} else {
		fmt.Println("📋 No foreign keys found")
	}

	// Check for indexes
	var indexes []string
	err = tx.Raw(`
		SELECT indexname 
		FROM pg_indexes 
		WHERE tablename = 'courses' 
		AND indexdef LIKE '%user_id%' OR indexdef LIKE '%code%'
	`).Scan(&indexes).Error

	if err == nil && len(indexes) > 0 {
		fmt.Printf("📋 Found indexes: %v\n", indexes)
	} else {
		fmt.Println("📋 No dependent indexes found")
	}
}

