# Database Service

The database service provides enterprise-grade database initialization and migration functionality for both client (SQLite) and server (PostgreSQL) databases.

## Architecture

The service is organized with clear separation between client and server logic:

```
internal/services/database/
├── types.go      # Types, interfaces, and configuration
├── migrator.go   # Common migration utilities
├── client.go     # Client database (SQLite) operations
├── server.go     # Server database (PostgreSQL) operations
└── README.md     # This file
```

## Features

- **Clear Separation**: Distinct handling for client (SQLite) and server (PostgreSQL) databases
- **Enterprise-Grade**: Well-structured, maintainable code following Go best practices
- **Reusable**: Functions can be used in scripts, CLI tools, and application initialization
- **Validation**: Built-in connection validation and migration checking
- **Flexible Configuration**: Configurable migration behavior

## Usage

### Client Database (SQLite)

#### Basic Initialization

```go
import "unipilot/internal/services/database"

// Initialize with default configuration (auto-migrate enabled)
db, err := database.InitializeClientDB(nil)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

#### Custom Configuration

```go
config := &database.DatabaseConfig{
    Type:         database.DatabaseTypeClient,
    AutoMigrate:  true,
    ValidateConn: true,
}

db, err := database.InitializeClientDB(config)
```

#### Initialize for Specific User

```go
db, err := database.InitializeClientDBWithID(userID, nil)
```

#### Manual Migration

```go
db, _ := utils.GetUserDB()
err := database.MigrateClientDB(db)
```

#### Check Migration Status

```go
needed, missingTables := database.CheckClientMigrationNeeded(db)
if needed {
    log.Printf("Migration needed for: %v", missingTables)
}
```

### Server Database (PostgreSQL)

#### Basic Initialization

```go
import "unipilot/internal/services/database"

// Initialize with default configuration (auto-migrate enabled)
db, err := database.InitializeServerDB(nil)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

#### Custom Configuration

```go
config := &database.DatabaseConfig{
    Type:         database.DatabaseTypeServer,
    AutoMigrate:  true,
    ValidateConn: true,
}

db, err := database.InitializeServerDB(config)
```

#### Manual Migration

```go
db, _ := storage.GetRemoteDB()
err := database.MigrateServerDB(db)
```

#### Check Migration Status

```go
needed, missingTables := database.CheckServerMigrationNeeded(db)
if needed {
    log.Printf("Migration needed for: %v", missingTables)
}
```

## Models

### Client Database Models

The client database uses the following models:
- `LocalCourse`
- `LocalAssignment`
- `LocalSync`
- `LocalDocument`
- `LocalNote`
- `LocalNotification`
- `LocalAiMessage`

### Server Database Models

The server database uses the following models:
- `User`
- `Course`
- `Assignment`
- `Note`
- `Document`
- `DocumentStorage`
- `Follow`
- `Device`

## Server-to-Client Sync

The service includes functions to trigger server-to-client synchronization:

### Trigger Sync After Database Initialization

```go
import "unipilot/internal/services/database"

// Ensure database is initialized and trigger sync
db, err := database.EnsureClientDBAndSync(userID)
if err != nil {
    log.Fatal(err)
}
```

### Trigger Sync on Existing Database

```go
// If database is already initialized, just trigger sync
err := database.EnsureClientDBAndSyncWithExisting(db)
```

### Manual Sync Trigger

```go
// Trigger sync manually (checks if online first)
err := database.TriggerServerToClientSync(db)
```

The sync functions:
- Check if the device is online before attempting sync
- Download all entities (courses, assignments, notes, documents) from server
- Store them in the local database
- Handle errors gracefully (partial sync is better than no sync)

## Integration with Scripts

This service can be used in migration scripts like `scripts/migration/migrate.go`:

```go
package main

import (
    "log"
    "unipilot/internal/services/database"
)

func main() {
    // Server database migration
    db, err := database.InitializeServerDB(nil)
    if err != nil {
        log.Fatalf("Failed to initialize server database: %v", err)
    }
    defer db.Close()
    
    log.Println("✅ Server database migration completed!")
}
```

## Error Handling

All functions return errors wrapped with context using the `unipilot/internal/errors` package. Common error codes:

- `errors.DBConnectionFailed`: Database connection failed
- `errors.DBQueryFailed`: Migration query failed
- `errors.ConfigEnvVarNotFound`: Missing environment variable (server DB)

## Best Practices

1. **Always validate connections** in production environments
2. **Check migration status** before running migrations in scripts
3. **Use appropriate configuration** for your use case (auto-migrate vs manual)
4. **Handle errors properly** - all functions return descriptive errors
5. **Close database connections** when done using them

## Testing

When testing, you can disable auto-migration and validation:

```go
config := &database.DatabaseConfig{
    Type:         database.DatabaseTypeClient,
    AutoMigrate:  false,
    ValidateConn: false,
}
```

## Migration Strategy

The service uses GORM's `AutoMigrate` feature, which:
- Creates missing tables
- Adds missing columns
- Creates missing indexes
- **Does not** delete unused columns or tables (for safety)

For production deployments, consider:
- Running migrations in a controlled environment first
- Using versioned migrations for complex schema changes
- Backing up databases before migrations

