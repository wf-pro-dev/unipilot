package sync

import (
	"log"
	"time"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/network"

	"gorm.io/gorm"
)

// Manager manages synchronization operations between client and server
type Manager struct {
	db     *gorm.DB
	config *SyncConfig
}

// NewManager creates a new sync manager instance
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:     db,
		config: DefaultSyncConfig(),
	}
}

// NewManagerWithConfig creates a new sync manager with custom configuration
func NewManagerWithConfig(db *gorm.DB, config *SyncConfig) *Manager {
	return &Manager{
		db:     db,
		config: config,
	}
}

// CreateSyncLog creates a sync log entry for failed operations
func (m *Manager) CreateSyncLog(entity models.Entity, entityID uint, action, column, value string, err error) error {
	update := &models.LocalSync{
		Entity:      entity,
		EntityID:    entityID,
		Action:      action,
		Column:      column,
		Value:       value,
		RetryCount:  0,
		NextRetryAt: time.Now().Add(GetBackoffDuration(0)),
		LastError:   err.Error(),
	}

	if err := m.db.Create(update).Error; err != nil {
		return errors.HandleDBCreateError(err)
	}
	return nil
}

// GetSyncLog retrieves a sync log entry
func (m *Manager) GetSyncLog(entity models.Entity, entityID uint, action, column string) (models.LocalSync, error) {
	var update models.LocalSync
	err := m.db.Where("entity = ? AND entity_id = ? AND action = ? AND column = ? AND deleted_at IS NULL",
		entity, entityID, action, column).First(&update).Error
	if err != nil {
		return update, errors.HandleDBReadError(err)
	}
	return update, nil
}

// GetPendingSyncs returns syncs that are ready to be retried
func (m *Manager) GetPendingSyncs() ([]models.LocalSync, error) {
	var updates []models.LocalSync
	err := m.db.Where("deleted_at IS NULL AND next_retry_at <= ?", time.Now()).Find(&updates).Error
	if err != nil {
		return updates, errors.HandleDBReadError(err)
	}
	return updates, nil
}

// MarkSyncAttempted updates retry count and next retry time
func (m *Manager) MarkSyncAttempted(update *models.LocalSync, err error) error {
	update.RetryCount++
	update.LastError = err.Error()
	backoffDuration := CalculateBackoffDuration(update.RetryCount, m.config.BaseBackoff, m.config.MaxBackoff)
	update.NextRetryAt = time.Now().Add(backoffDuration)

	if err := m.db.Save(update).Error; err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// MarkSyncCompleted marks a sync as successful and removes it from the queue
func (m *Manager) MarkSyncCompleted(update *models.LocalSync) error {
	if err := m.db.Delete(update).Error; err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// Undo removes a sync log entry for an entity
func (m *Manager) Undo(entity models.Entity, entityID uint) error {
	if err := m.db.Delete(&models.LocalSync{Entity: entity, EntityID: entityID}).Error; err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// StartBackgroundSync starts a background goroutine that periodically processes pending syncs
func (m *Manager) StartBackgroundSync() {
	go m.backgroundSync()
}

// backgroundSync runs periodic sync in the background
func (m *Manager) backgroundSync() {
	ticker := time.NewTicker(m.config.BackgroundInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !network.IsOnline() {
			continue
		}

		if err := m.ProcessPendingSyncs(); err != nil {
			log.Printf("[SyncManager] Background sync error: %v", err)
		}
	}
}

// ProcessPendingSyncs processes all pending syncs
func (m *Manager) ProcessPendingSyncs() error {
	pendingSyncs, err := m.GetPendingSyncs()
	if err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to get pending syncs")
	}

	if len(pendingSyncs) == 0 {
		return nil
	}

	// Get remote data once for all syncs
	remoteData, err := m.fetchRemoteData()
	if err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to fetch remote data")
	}

	// Process each sync
	for _, syncLog := range pendingSyncs {
		if err := m.processSync(syncLog, remoteData); err != nil {
			log.Printf("[SyncManager] Failed to process sync %d: %v", syncLog.ID, err)
			// Continue with other syncs instead of failing completely
		}
	}

	return nil
}

// RemoteData holds all remote entities needed for sync operations
type RemoteData struct {
	Assignments []models.Assignment
	Courses     []models.Course
	Notes       []models.Note
	Documents   []models.Document
}

// fetchRemoteData fetches all remote entities needed for sync operations
func (m *Manager) fetchRemoteData() (*RemoteData, error) {
	// TODO: Implement parallel fetching for better performance
	// For now, fetch sequentially

	data := &RemoteData{}

	// Fetch assignments
	assignments, err := fetchRemoteAssignments()
	if err != nil {
		log.Printf("[SyncManager] Warning: Failed to fetch remote assignments: %v", err)
	} else {
		data.Assignments = assignments
	}

	// Fetch courses
	courses, err := fetchRemoteCourses()
	if err != nil {
		log.Printf("[SyncManager] Warning: Failed to fetch remote courses: %v", err)
	} else {
		data.Courses = courses
	}

	// Fetch notes
	notes, err := fetchRemoteNotes()
	if err != nil {
		log.Printf("[SyncManager] Warning: Failed to fetch remote notes: %v", err)
	} else {
		data.Notes = notes
	}

	// Fetch documents
	documents, err := fetchRemoteDocuments()
	if err != nil {
		log.Printf("[SyncManager] Warning: Failed to fetch remote documents: %v", err)
	} else {
		data.Documents = documents
	}

	return data, nil
}

// processSync processes a single sync operation
func (m *Manager) processSync(syncLog models.LocalSync, remoteData *RemoteData) error {
	var err error

	switch syncLog.Entity {
	case models.EntityAssignment:
		err = m.syncAssignment(syncLog, remoteData.Assignments)
	case models.EntityCourse:
		err = m.syncCourse(syncLog, remoteData.Courses)
	case models.EntityNote:
		err = m.syncNote(syncLog, remoteData.Notes)
	case models.EntityDocument:
		err = m.syncDocument(syncLog, remoteData.Documents)
	case models.EntityUser:
		err = m.syncUser(syncLog)
	default:
		log.Printf("[SyncManager] Unknown entity type: %s", syncLog.Entity)
		return nil
	}

	if err != nil {
		if ShouldRetry(err) && syncLog.RetryCount < m.config.MaxRetries {
			return m.MarkSyncAttempted(&syncLog, err)
		}
		// Max retries exceeded or non-retryable error
		log.Printf("[SyncManager] Sync failed permanently for %s %d: %v", syncLog.Entity, syncLog.EntityID, err)
		return err
	}

	return m.MarkSyncCompleted(&syncLog)
}
