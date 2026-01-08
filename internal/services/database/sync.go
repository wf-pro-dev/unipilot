package database

import (
	"log"

	"unipilot/internal/errors"
	"unipilot/internal/network"
	syncservice "unipilot/internal/services/sync"

	"gorm.io/gorm"
)

// TriggerServerToClientSync triggers a server-to-client sync if the user is online
// This downloads all data from the server and stores it in the local database
// Returns nil if offline (sync will happen when online) or if sync succeeds
func TriggerServerToClientSync(db *gorm.DB) error {
	if !network.IsOnline() {
		log.Println("[ClientDB] Offline, skipping server-to-client sync (will sync when online)")
		return nil
	}

	log.Println("[ClientDB] Triggering server-to-client sync...")

	// Create migrator to download data from server
	migrator := syncservice.NewMigrator(db)

	// Migrate all entities from server to client
	if err := migrator.MigrateAll(); err != nil {
		// Log error but don't fail - partial sync is better than no sync
		log.Printf("[ClientDB] ⚠️  Server-to-client sync completed with errors: %v", err)
		return errors.Wrap(err, errors.SyncFailed, "Server-to-client sync had errors")
	}

	log.Println("[ClientDB] ✅ Server-to-client sync completed successfully")
	return nil
}
