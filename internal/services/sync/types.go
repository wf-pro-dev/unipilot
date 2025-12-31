package sync

import (
	"time"

	"unipilot/internal/models"

	"gorm.io/gorm"
)

// SyncDirection represents the direction of synchronization
type SyncDirection string

const (
	SyncDirectionClientToServer SyncDirection = "client_to_server" // Push local changes to remote
	SyncDirectionServerToClient SyncDirection = "server_to_client" // Pull remote changes to local
)

// SyncOperation represents a single sync operation
type SyncOperation struct {
	Entity   models.Entity
	EntityID uint
	Action   string // create, update, delete
	Column   string
	Value    string
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Success   bool
	Error     error
	Retryable bool
	NextRetry time.Time
}

// SyncStrategy defines the interface for entity-specific sync strategies
type SyncStrategy interface {
	// SyncCreate handles creating a new entity on the remote server
	SyncCreate(localID uint, db *gorm.DB) (*SyncResult, error)

	// SyncUpdate handles updating an existing entity on the remote server
	SyncUpdate(syncLog models.LocalSync, remoteEntity interface{}, db *gorm.DB) (*SyncResult, error)

	// SyncDelete handles deleting an entity on the remote server
	SyncDelete(syncLog models.LocalSync, remoteEntity interface{}, db *gorm.DB) (*SyncResult, error)

	// GetRemoteEntities fetches all remote entities of this type
	GetRemoteEntities() (interface{}, error)

	// FindRemoteEntity finds a specific remote entity by ID
	FindRemoteEntity(remoteEntities interface{}, remoteID uint) interface{}
}

// SyncConfig holds configuration for sync operations
type SyncConfig struct {
	MaxRetries         int
	BaseBackoff        time.Duration
	MaxBackoff         time.Duration
	BackgroundInterval time.Duration
}

// DefaultSyncConfig returns the default sync configuration
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		MaxRetries:         10,
		BaseBackoff:        1 * time.Second,
		MaxBackoff:         5 * time.Minute,
		BackgroundInterval: 2 * time.Hour,
	}
}
