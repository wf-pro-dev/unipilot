package sync

import (
	"log"
	"strconv"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/models"
	"unipilot/internal/network"

	"gorm.io/gorm"
)

type SyncManager struct {
	db *gorm.DB
}

func NewSyncManager(db *gorm.DB) *SyncManager {
	db = db.Debug()
	return &SyncManager{db: db}
}

// ShouldRetry determines if an operation should be retried based on error type
func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a network/connection error
	// You can expand this based on your error types
	return true // For now, retry all errors
}

// GetBackoffDuration calculates exponential backoff
func GetBackoffDuration(retryCount int) time.Duration {
	baseDelay := 1 * time.Second
	maxDelay := 5 * time.Minute

	delay := baseDelay * time.Duration(1<<retryCount)
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// CreateSyncLog creates a sync log entry for failed operations
func (sm *SyncManager) CreateSyncLog(entity models.Entity, entityID uint, action, column, value string, err error) error {
	update := &models.LocalUpdate{
		Entity:      entity,
		EntityID:    entityID,
		Action:      action,
		Column:      column,
		Value:       value,
		RetryCount:  0,
		NextRetryAt: time.Now().Add(GetBackoffDuration(0)),
		LastError:   err.Error(),
	}

	return sm.db.Create(update).Error
}

func (sm *SyncManager) GetSyncLog(entity models.Entity, entityID uint, action, column string) (models.LocalUpdate, error) {
	var update models.LocalUpdate
	err := sm.db.Where("entity = ? AND entity_id = ? AND action = ? AND column = ? AND deleted_at IS NULL", entity, entityID, action, column).First(&update).Error
	return update, err
}

// GetPendingSyncs returns syncs that are ready to be retried
func (sm *SyncManager) GetPendingSyncs() ([]models.LocalUpdate, error) {
	var updates []models.LocalUpdate
	err := sm.db.Where("deleted_at IS NULL AND next_retry_at <= ?", time.Now()).Find(&updates).Error
	return updates, err
}

// MarkSyncAttempted updates retry count and next retry time
func (sm *SyncManager) MarkSyncAttempted(update *models.LocalUpdate, err error) error {
	update.RetryCount++
	update.LastError = err.Error()
	update.NextRetryAt = time.Now().Add(GetBackoffDuration(update.RetryCount))
	return sm.db.Save(update).Error
}

// MarkSyncCompleted marks a sync as successful
func (sm *SyncManager) MarkSyncCompleted(update *models.LocalUpdate) error {
	return sm.db.Delete(update).Error
}

// BackgroundSync runs periodic sync in the background
func (sm *SyncManager) BackgroundSync() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if !network.IsOnline() {
			continue
		}

		if err := sm.ProcessPendingSyncs(); err != nil {
			log.Printf("[SyncManager] Background sync error: %v", err)
		}
	}
}

// ProcessPendingSyncs processes all pending syncs
func (sm *SyncManager) ProcessPendingSyncs() error {
	pendingSyncs, err := sm.GetPendingSyncs()
	if err != nil {
		return err
	}

	// Get remote assignments
	remoteAssignments, err := client.GetAssignments()
	if err != nil {
		log.Println("[Sync] Error getting remote assignments", err)
		return err
	}

	// Get remote courses
	remoteCourses, err := client.GetCourses()
	if err != nil {
		log.Println("[Sync] Error getting remote courses", err)
		return err
	}

	for _, syncLog := range pendingSyncs {
		if err := sm.ProcessSync(syncLog, remoteAssignments, remoteCourses); err != nil {
			log.Printf("[SyncManager] Failed to process sync %d: %v", syncLog.ID, err)
			// Continue with other syncs instead of failing completely
		}
	}

	return nil
}

// ProcessSync processes a single sync operation
func (sm *SyncManager) ProcessSync(syncLog models.LocalUpdate, remoteAssignments, remoteCourses []map[string]string) error {

	switch syncLog.Entity {
	case models.EntityAssignment:
		if err := SyncAssignment(syncLog, findRemoteEntity(remoteAssignments, syncLog.EntityID), sm.db); err != nil {
			sm.MarkSyncAttempted(&syncLog, err)
			return err
		}
	case models.EntityCourse:
		if err := SyncCourse(syncLog, findRemoteEntity(remoteCourses, syncLog.EntityID), sm.db); err != nil {
			sm.MarkSyncAttempted(&syncLog, err)
			return err
		}
	}

	return sm.MarkSyncCompleted(&syncLog)
}

func findRemoteEntity(remoteEntities []map[string]string, localEntityID uint) map[string]string {
	for _, remoteEntity := range remoteEntities {
		if remoteEntity["local_id"] == strconv.Itoa(int(localEntityID)) {
			return remoteEntity
		}
	}
	return nil
}
