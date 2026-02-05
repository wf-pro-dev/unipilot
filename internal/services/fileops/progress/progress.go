package progress

import (
	"context"
	"sync"
	"time"
)

// ProgressCallback is called when progress updates
type ProgressCallback func(progress *Progress)

// Progress represents a single operation
type Progress struct {
	ID        string
	Total     int64
	Current   int64
	Status    string
	StartTime time.Time
	Error     error
	mu        sync.RWMutex
	callbacks []ProgressCallback
	cancel    context.CancelFunc
}

type ProgressSnapshot struct {
	ID         string    `json:"progress_id"`
	Total      int64     `json:"total"`
	Current    int64     `json:"current"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"start_time"`
	Error      error     `json:"error"`
	Percentage float64   `json:"percentage"`
}

// NewTracker creates a new progress tracker
func NewProgress(id string, total int64, ctx context.Context) *Progress {
	_, cancel := context.WithCancel(ctx)

	return &Progress{

		ID:        id,
		Total:     total,
		Current:   0,
		Status:    "starting",
		StartTime: time.Now(),
		callbacks: make([]ProgressCallback, 0),
		cancel:    cancel,
	}
}

// Update progress
func (t *Progress) Update(current int64) {
	t.mu.Lock()
	t.Current = current
	t.mu.Unlock()

	t.notifyCallbacks()
}

// Increment progress by delta
func (t *Progress) Increment(delta int64) {
	t.mu.Lock()
	t.Current += delta
	t.mu.Unlock()

	t.notifyCallbacks()
}

// SetStatus updates the status message
func (t *Progress) SetStatus(status string) {
	t.mu.Lock()
	t.Status = status
	t.mu.Unlock()

	t.notifyCallbacks()
}

// SetError marks the tracker as failed
func (t *Progress) SetError(err error) {
	t.mu.Lock()
	t.Error = err
	t.Status = "error"
	t.mu.Unlock()

	t.notifyCallbacks()
}

// Complete marks the tracker as finished
func (t *Progress) Complete() {
	t.mu.Lock()
	t.Current = t.Total
	t.Status = "completed"
	t.mu.Unlock()

	t.notifyCallbacks()
}

// Percentage returns current progress as percentage (0-100)
func (t *Progress) Percentage() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Total == 0 {
		return 0
	}
	return float64(t.Current) / float64(t.Total) * 100
}

// IsComplete returns true if progress is 100%
func (t *Progress) IsComplete() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Current >= t.Total && t.Total > 0
}

// Snapshot returns a thread-safe copy of the tracker state
func (t *Progress) Snapshot() ProgressSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return ProgressSnapshot{
		ID:         t.ID,
		Total:      t.Total,
		Current:    t.Current,
		Status:     t.Status,
		StartTime:  t.StartTime,
		Error:      t.Error,
		Percentage: t.Percentage(),
	}
}

// OnProgress registers a callback for progress updates
func (t *Progress) OnProgress(callback ProgressCallback) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.callbacks = append(t.callbacks, callback)
}

func (t *Progress) notifyCallbacks() {
	t.mu.RLock()
	callbacks := make([]ProgressCallback, len(t.callbacks))
	copy(callbacks, t.callbacks)
	t.mu.RUnlock()

	for _, callback := range callbacks {
		callback(t)
	}
}

func StoreCancelFunc(id string, cancel context.CancelFunc) {
	cancelMutex.Lock()
	defer cancelMutex.Unlock()
	cancelFuncs[id] = cancel
}

func RemoveCancelFunc(id string) {
	cancelMutex.Lock()
	defer cancelMutex.Unlock()
	delete(cancelFuncs, id)
}

func GetCancelFunc(id string) (context.CancelFunc, bool) {
	cancelMutex.RLock()
	defer cancelMutex.RUnlock()
	cancel, exists := cancelFuncs[id]
	return cancel, exists
}
