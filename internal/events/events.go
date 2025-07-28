package events

import (
	"encoding/json"
	"log"

	"unipilot/internal/sse"
)

type Events struct {
	stopChan chan struct{}
}

func NewEvents() *Events {
	return &Events{
		stopChan: make(chan struct{}),
	}
}

// Start now accepts the sseClient as a parameter.
func (h *Events) Start(sseClient *sse.SSE) {
	if sseClient == nil {
		log.Fatal("[EventHandler] Fatal: SSE client is nil.")
		return
	}

	go func() {
		log.Println("[EventHandler] Starting to listen for events...")
		for {
			select {
			case <-h.stopChan:
				log.Println("[EventHandler] Stop signal received, shutting down.")
				return
			// Listen for events from the client's public channel.
			case event, ok := <-sseClient.Events():
				if !ok {
					log.Println("[EventHandler] SSE events channel closed.")
					return
				}
				h.HandleEvent(event)
			// Listen for errors.
			case err, ok := <-sseClient.Errors():
				if !ok {
					log.Println("[EventHandler] SSE errors channel closed.")
					return
				}
				log.Printf("[EventHandler] Received SSE error: %v", err)
			}
		}
	}()
}

// Stop signals the event handling goroutine to terminate.
func (h *Events) Stop() {
	close(h.stopChan)
}

func (h *Events) HandleEvent(event sse.Event) {
	var notification struct {
		Type    string          `json:"type"`
		Entity  string          `json:"entity"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(event.Data, &notification); err != nil {
		log.Printf("[EventHandler] Error parsing notification: %v", err)
		return
	}

	// Route the event based on its entity and type.
	switch notification.Entity {
	case "assignment":
		switch notification.Type {
		case "create":
			h.HandleAssignmentCreate(notification.Data, notification.Message)
		case "update":
			h.HandleAssignmentUpdate(notification.Data, notification.Message)
		case "delete":
			h.HandleAssignmentDelete(notification.Data, notification.Message)
		}
	case "course":
		// Placeholder for future course event handling.
	}
}
