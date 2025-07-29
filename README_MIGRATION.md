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