# Document Management Migration Guide

This guide explains how to run the **one-time database migrations** to add document management functionality.

## ⚠️ Important Notes

- These migrations only need to be run **once**
- After running successfully, **delete these migration scripts**
- Make sure to backup your databases before running migrations

---

## 🖥️ Local Database Migration (Desktop App)

**Directory:** `cmd/migrate_local/`

### Prerequisites
- Desktop app has been run at least once (to create initial database)
- User is logged in (so user database exists)

### Run Migration
```bash
# From project root
go run cmd/migrate_local/main.go
```

### What it does:
- ✅ Creates `local_documents` table
- ✅ Creates `local_document_caches` table  
- ✅ Updates `local_assignments` table with Documents relationship
- ✅ Verifies migration success

### Expected Output:
```
🚀 Starting LOCAL database migration for documents...
📂 Migrating database for user ID: 123
💾 Database path: /Users/you/.config/acc-homework/data/user_123.db
✅ Connected to local database
📋 Running local database migration...
  → Creating document tables...
  → Updating LocalAssignment table...
  → Verifying migration...
  ✅ Table local_documents created successfully
  ✅ Table local_document_caches created successfully
🎉 LOCAL database migration completed successfully!
⚠️  This script can now be DELETED as the migration is complete.
```

---

## 🌐 Remote Database Migration (Server)

**Directory:** `cmd/migrate_remote/`

### Prerequisites
- `.env` file with `DATABASE_URL` configured
- Access to PostgreSQL database
- Server dependencies installed (`go mod tidy`)

### Run Migration
```bash
# On your server (or locally with server DB access)
go run cmd/migrate_remote/main.go
```

### What it does:
- ✅ Creates `documents` table
- ✅ Creates `document_storage_infos` table
- ✅ Updates `assignments` table with Documents relationship
- ✅ Verifies migration success

### Expected Output:
```
🚀 Starting REMOTE database migration for documents...
🔗 Connecting to remote database...
✅ Connected to remote database
📋 Running remote database migration...
  → Creating document tables...
  → Updating Assignment table...
  → Verifying migration...
  ✅ Table documents verified
  ✅ Table document_storage_infos verified
  ✅ Table assignments verified
  📊 Found 45 existing assignments
🎉 REMOTE database migration completed successfully!
⚠️  This script can now be DELETED as the migration is complete.
```

---

## 🧹 After Migration

1. **Test the functionality** - Upload a document to verify everything works
2. **Delete migration files:**
   ```bash
   rm -rf cmd/migrate_local/
   rm -rf cmd/migrate_remote/
   rm README_MIGRATION.md
   ```
3. **Commit the changes:**
   ```bash
   git add -A
   git commit -m "chore: remove migration scripts after successful migration"
   ```

---

## 🚨 Troubleshooting

### Local Migration Issues
- **"Database file does not exist"** → Run the desktop app first
- **"Failed to get current user ID"** → Make sure you're logged in
- **Permission errors** → Check file permissions in config directory

### Remote Migration Issues  
- **"DATABASE_URL not found"** → Add to `.env` file
- **Connection failed** → Check database URL and network access
- **Table already exists** → Migration may have already run (check output)

### General Issues
- **"Foreign key constraint failed"** → Some data may be incompatible
- **Migration partially completed** → Check which tables were created, may need manual cleanup

---

## 📋 Migration Checklist

- [ ] Backup databases
- [ ] Run local migration (`go run cmd/migrate_local/main.go`)
- [ ] Run remote migration (`go run cmd/migrate_remote/main.go`)  
- [ ] Test document upload functionality
- [ ] Delete migration directories
- [ ] Commit changes

---

**Need Help?** Check the output messages - they provide specific guidance for each step. 

## Course Code Unique Constraint Migration

### Problem
The original implementation had unique constraints on the `code` field in both `local_courses` (SQLite) and `courses` (PostgreSQL) tables. This caused issues when users tried to create a new course with a code that was previously used by a deleted course, even though the previous course was soft-deleted.

### Solution
The unique constraints have been replaced with composite unique indexes that only apply to non-deleted records:

**Local Database (SQLite):**
```go
Code string `gorm:"index:idx_course_code_active,unique,where:deleted_at IS NULL"`
```

**Remote Database (PostgreSQL):**
```sql
CREATE UNIQUE INDEX idx_courses_user_code_active ON courses(user_id, code) WHERE deleted_at IS NULL;
```

This allows users to:
1. Create courses with unique codes
2. Delete courses (soft delete)
3. Create new courses with the same code as previously deleted courses

### Local Database Migration

The local SQLite database migration is handled automatically by the application:

1. **Automatic Migration** (Recommended)
   - The new code will automatically handle the migration
   - Soft-deleted courses with duplicate codes will be cleaned up
   - New courses can be created with previously used codes

2. **Manual Migration** (If automatic fails)
   ```sql
   -- Drop the old unique constraint (if it exists)
   DROP INDEX IF EXISTS idx_local_courses_code;
   
   -- Create the new composite index
   CREATE UNIQUE INDEX idx_course_code_active ON local_courses(code) WHERE deleted_at IS NULL;
   ```

### Remote Database Migration

For the remote PostgreSQL database, use the provided migration script:

1. **Navigate to the migration directory:**
   ```bash
   cd scripts/migration/remote
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Set environment variables:**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=your_password
   export DB_NAME=unipilot
   export DB_SSLMODE=disable
   ```

4. **Run the migration:**
   ```bash
   go run migrate_remote_courses.go
   ```

### Testing the Migration

You can test the migration using the provided test method:

```go
// This will create a course, delete it, then create another with the same code
err := app.TestCourseCreationWithDuplicateCode("TEST101")
if err != nil {
    log.Printf("Migration test failed: %v", err)
} else {
    log.Println("Migration test passed!")
}
```

### Implementation Details

1. **BeforeCreate Hook**: Added to the `LocalCourse` model to handle code conflicts
2. **CreateCourse Function**: Updated to clean up soft-deleted courses before creating new ones
3. **Migration Function**: Automatically handles existing database schemas
4. **Cleanup Function**: Removes duplicate soft-deleted courses
5. **Remote Migration Script**: Standalone Go script for PostgreSQL migration

### Benefits

- ✅ Users can reuse course codes after deletion
- ✅ Maintains data integrity for active courses
- ✅ Backward compatible with existing data
- ✅ Automatic migration for local databases
- ✅ Standalone migration script for remote databases
- ✅ Manual migration path for existing databases 