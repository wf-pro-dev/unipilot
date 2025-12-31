# Sync Service

Enterprise-grade synchronization service for client-server and server-client data synchronization.

## Architecture

The sync service is organized into clear, focused modules:

### Core Components

- **`types.go`** - Type definitions, interfaces, and configuration
- **`manager.go`** - Main sync manager with retry logic and background processing
- **`client_sync.go`** - Client-to-server sync operations (push local changes)
- **`server_sync.go`** - Server-to-client sync operations (pull remote data)
- **`utils.go`** - Helper functions for finding remote entities
- **`backoff.go`** - Retry and exponential backoff logic

## Usage

### Client-to-Server Sync (Push Local Changes)

```go
import syncservice "unipilot/internal/services/sync"

// Create a sync manager
manager := syncservice.NewManager(db)

// Create a sync log for failed operations
err := manager.CreateSyncLog(
    models.EntityAssignment,
    assignmentID,
    "create",
    "title",
    "New Assignment",
    err,
)

// Process pending syncs
err := manager.ProcessPendingSyncs()

// Start background sync (runs periodically)
manager.StartBackgroundSync()
```

### Server-to-Client Sync (Pull Remote Data)

```go
import syncservice "unipilot/internal/services/sync"

// Create a migrator
migrator := syncservice.NewMigrator(db)

// Migrate specific entities
err := migrator.MigrateCourses()
err := migrator.MigrateAssignments()
err := migrator.MigrateNotes()
err := migrator.MigrateDocuments()

// Or migrate all entities at once
err := migrator.MigrateAll()
```

## Features

### Retry Logic
- Exponential backoff with configurable base and max delays
- Maximum retry count protection
- Automatic retry for network errors
- Background processing of failed syncs

### Sync Operations
- **Create**: Push new local entities to remote server
- **Update**: Push local changes to remote server
- **Delete**: Sync deletions to remote server
- **Conflict Resolution**: Checks remote `UpdatedAt` to prevent overwriting newer data

### Supported Entities
- Assignments
- Courses
- Notes
- Documents
- Users

## Configuration

```go
config := &syncservice.SyncConfig{
    MaxRetries:        10,
    BaseBackoff:       1 * time.Second,
    MaxBackoff:        5 * time.Minute,
    BackgroundInterval: 2 * time.Hour,
}

manager := syncservice.NewManagerWithConfig(db, config)
```

## Error Handling

The sync service uses the application's error handling system:
- Network errors are automatically retried
- Sync failures are logged to `LocalSync` table
- Non-retryable errors are logged but don't stop other syncs

## Migration from Old Sync Package

The old `internal/sync` package has been replaced with this service. Key changes:

- `sync.NewSyncManager()` → `syncservice.NewManager()`
- `sync.MigrateCourses()` → `syncservice.NewMigrator(db).MigrateCourses()`
- `sync.BackgroundSync()` → `syncservice.NewManager(db).StartBackgroundSync()`

