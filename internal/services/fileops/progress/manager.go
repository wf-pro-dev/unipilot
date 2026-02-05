// pkg/progress/manager.go
package progress

import (
	"context"
	"sync"
	"time"
)

var (
	ProgressManager *Manager
	cancelMutex     sync.RWMutex
	cancelFuncs     = make(map[string]context.CancelFunc)
	once            sync.Once
)

// Manager manages multiple progress trackers
type Manager struct {
	ctx        context.Context
	progresses map[string]*Progress
	mu         sync.RWMutex
	ttl        time.Duration
}

// NewManager creates a new progress manager
func NewManager(ctx context.Context, ttl time.Duration) *Manager {
	m := &Manager{
		ctx:        ctx,
		progresses: make(map[string]*Progress),
		ttl:        ttl,
	}

	// Start cleanup routine
	go m.cleanupLoop()

	return m
}

// GetManager returns singleton progress manager
func GetManager(ctx context.Context) *Manager {
	once.Do(func() {
		ProgressManager = NewManager(ctx, 1*time.Hour)
	})
	return ProgressManager
}

// Create a new tracker and register it
func (m *Manager) Create(id string, total int64) *Progress {
	progress := NewProgress(id, total, m.ctx)

	m.mu.Lock()
	m.progresses[id] = progress
	m.mu.Unlock()

	return progress
}

// Get retrieves a tracker by ID
func (m *Manager) Get(id string) (*Progress, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, exists := m.progresses[id]
	return progress, exists
}

// GetSnapshot returns a snapshot of a tracker
func (m *Manager) GetSnapshot(id string) (ProgressSnapshot, bool) {
	progress, exists := m.Get(id)
	if !exists {
		return ProgressSnapshot{}, false
	}

	return progress.Snapshot(), true
}

// Remove deletes a tracker
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.progresses, id)
}

// List returns all tracker snapshots
func (m *Manager) List() []ProgressSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := make([]ProgressSnapshot, 0, len(m.progresses))
	for _, progress := range m.progresses {
		snapshots = append(snapshots, progress.Snapshot())
	}

	return snapshots
}

// Cancel provides a way to cancel operations
func (m *Manager) Cancel(id string) bool {
	progress, exists := m.Get(id)
	if !exists {
		return false
	}

	progress.SetError(context.Canceled)
	return true
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(m.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, progress := range m.progresses {
			snapshot := progress.Snapshot()

			// Remove completed or errored trackers after TTL
			if (snapshot.Status == "completed" || snapshot.Error != nil) &&
				now.Sub(snapshot.StartTime) > m.ttl {
				delete(m.progresses, id)
			}
		}
		m.mu.Unlock()
	}
}
