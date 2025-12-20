package notifications

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"unipilot/internal/errors"
	"unipilot/internal/models/course"
	"unipilot/internal/models/notifications"
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

	// Initialize any daemon-specific configurations
	log.Printf("[EventHandler] Initialized daemon for user %d", userID)
	return nil
}

// StartEventHandler initializes and starts listening to SSE events using existing SSE connection
func (eh *EventHandler) StartEventHandler(existingSSE *sse.SSE) error {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	if eh.isRunning {
		log.Printf("[EventHandler] Event handler is already running")
		return nil
	}

	if existingSSE == nil {
		return errors.NewAppError(errors.ValidationInvalid, "Existing SSE client is nil", nil)
	}

	log.Printf("[EventHandler] Starting event handler for user %d using existing SSE connection", eh.userID)

	// Use the existing SSE client instead of creating a new one
	eh.sseClient = existingSSE

	// Start event processing in a separate goroutine
	go eh.processEvents()

	eh.isRunning = true
	log.Printf("[EventHandler] Event handler started successfully using existing SSE connection")
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

	log.Printf("[EventHandler] Stopping event handler for user %d", eh.userID)

	// Cancel context to stop all goroutines
	if eh.cancel != nil {
		eh.cancel()
	}

	// Note: We don't stop the SSE client here since it's managed by the main app
	// We just stop our event processing

	eh.isRunning = false
	log.Printf("[EventHandler] Event handler stopped successfully")
	return nil
}

// IsEventHandlerRunning checks if the event handler is currently active
func (eh *EventHandler) IsEventHandlerRunning() bool {
	eh.mu.RLock()
	defer eh.mu.RUnlock()
	return eh.isRunning
}

// HandleFollowNotification processes follow events from SSE
func (eh *EventHandler) HandleFollowNotification(notification notifications.LocalNotification) {
	log.Printf("[EventHandler] Processing follow notification: %s", notification.Message)
	log.Printf("[EventHandler] Data: %s", string(notification.Data))

	// Set the notification type and ensure it's marked as unread
	notification.Type = notifications.NotificationFollow
	notification.Read = false
	notification.ExpiresAt = &time.Time{}

	if err := eh.db.Create(&notification).Error; err != nil {
		wrappedErr := errors.HandleDBCreateError(err)
		log.Printf("[EventHandler] Error saving notification: %v", wrappedErr)
		return
	}

	if err := beeep.Notify(notification.Title, notification.Message, ""); err != nil {
		log.Printf("[EventHandler] Error sending system notification: %v", err)
	} else {
		log.Printf("[EventHandler] Sent follow notification: %s", notification.Title)
	}
}

// HandleSyncNotification processes sync events for notes/assignments
func (eh *EventHandler) HandleSyncNotification(notification notifications.LocalNotification) {
	log.Printf("[EventHandler] Processing sync notification: %s", notification.Message)
	log.Printf("[EventHandler] Data: %s", string(notification.Data))

	notification.Type = notifications.NotificationSync
	notification.Read = false
	notification.ExpiresAt = &time.Time{}

	if err := eh.db.Create(&notification).Error; err != nil {
		log.Printf("[EventHandler] Error parsing sync data: %v", err)
		return
	}

	if err := beeep.Notify(notification.Title, notification.Message, ""); err != nil {
		log.Printf("[EventHandler] Error sending system notification: %v", err)
	} else {
		log.Printf("[EventHandler] Sent sync notification: %s", notification.Title)
	}
}

func (eh *EventHandler) HandleLinkNotification(notification notifications.LocalNotification) {
	log.Printf("[EventHandler] Processing link notification: %s", notification.Message)
	log.Printf("[EventHandler] Data: %s", string(notification.Data))

	notification.Type = notifications.NotificationLink
	notification.Read = false
	notification.ExpiresAt = &time.Time{}

	// Parse the data
	var data struct {
		CourseCode string `json:"course_code"`
		LinkID     uint   `json:"link_id"`
	}
	if err := json.Unmarshal([]byte(notification.Data), &data); err != nil {
		log.Printf("[EventHandler] Error parsing link data: %v", err)
		return
	}

	// Update  the course by code
	if err := eh.db.Model(&course.LocalCourse{Code: data.CourseCode}).Update("link_id", data.LinkID).Error; err != nil {
		log.Printf("[EventHandler] Link data: %+v", data)
	}

	if err := beeep.Notify(notification.Title, notification.Message, ""); err != nil {
		log.Printf("[EventHandler] Error sending system notification: %v", err)
	} else {
		log.Printf("[EventHandler] Sent assignment notification: %s", notification.Title)
	}
}

// HandleAssignmentNotification processes real-time assignment events
func (eh *EventHandler) HandleAssignmentNotification(notification notifications.LocalNotification) {
	log.Printf("[EventHandler] Processing assignment notification: %s", notification.Message)
	log.Printf("[EventHandler] Data: %s", string(notification.Data))

	notification.Type = notifications.NotificationAssignmentUpdate
	notification.Read = false
	notification.ExpiresAt = &time.Time{}

	if err := eh.db.Create(&notification).Error; err != nil {
		log.Printf("[EventHandler] Error parsing assignment data: %v", err)
		return
	}

	if err := beeep.Notify(notification.Title, notification.Message, ""); err != nil {
		log.Printf("[EventHandler] Error sending system notification: %v", err)
	} else {
		log.Printf("[EventHandler] Sent assignment notification: %s", notification.Title)
	}
}

// handleDocumentNotification processes real-time document events
func (eh *EventHandler) handleDocumentNotification(notification notifications.LocalNotification) {
	log.Printf("[EventHandler] Processing document notification: %s", notification.Message)
	log.Printf("[EventHandler] Data: %s", string(notification.Data))

	notification.Type = notifications.NotificationDocumentUpdate
	notification.Read = false
	notification.ExpiresAt = &time.Time{}

	if err := eh.db.Create(&notification).Error; err != nil {
		log.Printf("[EventHandler] Error parsing document data: %v", err)
		return
	}

	if err := beeep.Notify(notification.Title, notification.Message, ""); err != nil {
		log.Printf("[EventHandler] Error sending system notification: %v", err)
	} else {
		log.Printf("[EventHandler] Sent document notification: %s", notification.Title)
	}
}

// handleNoteNotification processes real-time note events
func (eh *EventHandler) handleNoteNotification(notification notifications.LocalNotification) {
	log.Printf("[EventHandler] Processing document notification: %s", notification.Message)
	log.Printf("[EventHandler] Data: %s", string(notification.Data))

	notification.Type = notifications.NotificationNoteUpdate
	notification.Read = false
	notification.ExpiresAt = &time.Time{}

	if err := eh.db.Create(&notification).Error; err != nil {
		log.Printf("[EventHandler] Error parsing note data: %v", err)
		return
	}

	if err := beeep.Notify(notification.Title, notification.Message, ""); err != nil {
		log.Printf("[EventHandler] Error sending system notification: %v", err)
	} else {
		log.Printf("[EventHandler] Sent note notification: %s", notification.Title)
	}
}

// processEvents processes incoming SSE events
func (eh *EventHandler) processEvents() {
	log.Printf("[EventHandler] Starting event processing loop")

	for {
		select {
		case <-eh.ctx.Done():
			log.Printf("[EventHandler] Event processing stopped (context cancelled)")
			return
		case event, ok := <-eh.sseClient.Events():
			if !ok {
				log.Printf("[EventHandler] SSE events channel closed")
				return
			}

			// Parse the event and route to appropriate handler
			eh.routeEvent(event)
		case err, ok := <-eh.sseClient.Errors():
			if !ok {
				log.Printf("[EventHandler] SSE errors channel closed")
				return
			}
			log.Printf("[EventHandler] SSE error: %v", err)
		}
	}
}

// routeEvent routes SSE events to appropriate handlers
func (eh *EventHandler) routeEvent(event sse.Event) {

	var notification notifications.LocalNotification
	if err := json.Unmarshal(event.Data, &notification); err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse notification event")
		log.Printf("[EventHandler] Error parsing notification: %v", wrappedErr)
		return
	}

	log.Printf("[EventHandler] Notification: %+v", notification)

	// Route based on entity type
	switch notification.Type {
	case notifications.NotificationFollow:
		eh.HandleFollowNotification(notification)
	case notifications.NotificationSync:
		eh.HandleSyncNotification(notification)
	case notifications.NotificationLink:
		eh.HandleLinkNotification(notification)
	case notifications.NotificationAssignmentUpdate:
		eh.HandleAssignmentNotification(notification)
	case notifications.NotificationDocumentUpdate:
		eh.handleDocumentNotification(notification)
	case notifications.NotificationNoteUpdate:
		eh.handleNoteNotification(notification)

	default:
		log.Printf("[EventHandler] Unknown entity type: %s", notification.Entity)
	}
}

// GetNotificationStatus returns current notification system status
func (eh *EventHandler) GetNotificationStatus() map[string]interface{} {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	status := map[string]interface{}{
		"event_handler_running": eh.isRunning,
		"user_id":               eh.userID,
		"timestamp":             time.Now().Unix(),
	}

	// Add SSE connection status if available
	if eh.sseClient != nil {
		status["sse_client_available"] = true
		// Note: We can't directly check if SSE is connected since it's managed by the main app
		// The main app should provide this information
	} else {
		status["sse_client_available"] = false
	}

	// Add context status
	if eh.ctx != nil {
		select {
		case <-eh.ctx.Done():
			status["context_cancelled"] = true
		default:
			status["context_cancelled"] = false
		}
	} else {
		status["context_cancelled"] = true
	}

	// Add error status if any recent errors occurred
	// This could be expanded to track recent errors
	status["recent_errors"] = []string{} // Placeholder for error tracking

	return status
}
