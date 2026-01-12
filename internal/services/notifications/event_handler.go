package notifications

import (
	"context"
	"log"
	"sync"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"

	"github.com/gen2brain/beeep"
	"gorm.io/gorm"
)

// EventHandler handles real-time notification events
type EventHandler struct {
	sseClient *sse.SSE
	db        *gorm.DB
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
	mu        sync.RWMutex
	userID    uint
}

// NewEventHandler creates a new event handler
func NewEventHandler(userID uint) (*EventHandler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	db, err := utils.GetUserDBWithID(userID)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to get user database")
	}
	return &EventHandler{
		userID:    userID,
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		isRunning: false,
	}, nil
}

// InitializeForDaemon sets up the event handler to run as a daemon
func (eh *EventHandler) InitializeForDaemon(userID uint) error {
	// Set the user ID for this daemon instance
	eh.userID = userID

	return nil
}

// StartEventHandler initializes and starts listening to SSE events using existing SSE connection
func (eh *EventHandler) StartEventHandler(existingSSE *sse.SSE) error {
	// Thread safe lock
	eh.mu.Lock()
	defer eh.mu.Unlock()

	if eh.isRunning {
		log.Printf("[EventHandler] Event handler is already running")
		return nil
	}

	if existingSSE == nil {
		return errors.NewAppError(errors.ValidationInvalid, "Existing SSE client is nil", nil)
	}
	// Use the existing SSE client instead of creating a new one
	eh.sseClient = existingSSE

	// Start event processing in a separate goroutine
	go eh.processEvents()

	eh.isRunning = true
	return nil
}

// StopEventHandler gracefully stops the event handler
func (eh *EventHandler) StopEventHandler() error {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	if !eh.isRunning {
		log.Printf("[EventHandler] Event handler is not running")
		return nil
	}
	// Cancel context to stop all goroutines
	if eh.cancel != nil {
		eh.cancel()
	}

	// Note: We don't stop the SSE client here since it's managed by the main app
	eh.isRunning = false
	return nil
}

// IsEventHandlerRunning checks if the event handler is currently active
func (eh *EventHandler) IsEventHandlerRunning() bool {
	eh.mu.RLock()
	defer eh.mu.RUnlock()
	return eh.isRunning
}

// processEvents processes incoming SSE events
func (eh *EventHandler) processEvents() {
	for {
		select {
		case <-eh.ctx.Done():
			return
		case event, ok := <-eh.sseClient.Events():
			if !ok {
				return
			}

			// Parse the event and route to appropriate handler
			eh.routeEvent(event)
		case err, ok := <-eh.sseClient.Errors():
			log.Printf("[EventHandler] SSE error: %v", err)
			if !ok {
				return
			}
		}
	}
}

// routeEvent routes SSE events to appropriate handlers
func (eh *EventHandler) routeEvent(message models.Message) {

	log.Printf("[EventHandler] Notification: %+v", message)

	if err := beeep.Notify(message.Title, message.Message, ""); err != nil {
		log.Printf("[EventHandler] Error sending system notification: %v", err)
	}

	return
}
