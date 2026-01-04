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
	trackers map[string]*Tracker
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewManager creates a new progress manager
func NewManager(ttl time.Duration) *Manager {
	m := &Manager{
		trackers: make(map[string]*Tracker),
		ttl:      ttl,
	}

	// Start cleanup routine
	go m.cleanupLoop()

	return m
}

// GetManager returns singleton progress manager
func GetManager() *Manager {
	once.Do(func() {
		ProgressManager = NewManager(1 * time.Hour)
	})
	return ProgressManager
}

// Create a new tracker and register it
func (m *Manager) Create(id string, total int64) *Tracker {
	tracker := NewTracker(id, total)

	m.mu.Lock()
	m.trackers[id] = tracker
	m.mu.Unlock()

	return tracker
}

// Get retrieves a tracker by ID
func (m *Manager) Get(id string) (*Tracker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tracker, exists := m.trackers[id]
	return tracker, exists
}

// GetSnapshot returns a snapshot of a tracker
func (m *Manager) GetSnapshot(id string) (TrackerSnapshot, bool) {
	tracker, exists := m.Get(id)
	if !exists {
		return TrackerSnapshot{}, false
	}

	return tracker.Snapshot(), true
}

// Remove deletes a tracker
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.trackers, id)
}

// List returns all tracker snapshots
func (m *Manager) List() []TrackerSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := make([]TrackerSnapshot, 0, len(m.trackers))
	for _, tracker := range m.trackers {
		snapshots = append(snapshots, tracker.Snapshot())
	}

	return snapshots
}

// Cancel provides a way to cancel operations
func (m *Manager) Cancel(id string) bool {
	tracker, exists := m.Get(id)
	if !exists {
		return false
	}

	tracker.SetError(context.Canceled)
	return true
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(m.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, tracker := range m.trackers {
			snapshot := tracker.Snapshot()

			// Remove completed or errored trackers after TTL
			if (snapshot.Status == "completed" || snapshot.Error != nil) &&
				now.Sub(snapshot.StartTime) > m.ttl {
				delete(m.trackers, id)
			}
		}
		m.mu.Unlock()
	}
}
