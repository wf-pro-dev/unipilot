package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"
)

type SSEClient struct {
	UserID    uint
	Messages  chan []byte
	Connected bool
	LastActive time.Time
}

type SSEServer struct {
	clients map[uint]*SSEClient
	mu      sync.RWMutex
	db      *gorm.DB
}

func NewSSEServer(db *gorm.DB) *SSEServer {
	return &SSEServer{
		clients: make(map[uint]*SSEClient),
		db:      db,
	}
}

func (s *SSEServer) AddClient(userID uint) *SSEClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := &SSEClient{
		UserID:    userID,
		Messages:  make(chan []byte, 100),
		Connected: true,
	}
	
	PrintLog(fmt.Sprintf("New SSE user id : %d\n",userID))
	s.clients[userID] = client
	return client
}

func (s *SSEServer) RemoveClient(userID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.clients[userID]; ok {
		close(client.Messages)
		delete(s.clients, userID)
	}
}

func (s *SSEServer) SendToUser(userID uint, message []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	s.logActiveClients()	
	if client, ok := s.clients[userID]; ok {
		select {
		case client.Messages <- message:
			PrintLog(fmt.Sprintf("new SSE message for user id : %d", userID))
			return true
		default:
			// Channel full, client might be slow
			return false
		}
	}
	return false
}

func (s *SSEServer) Broadcast(message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		select {
		case client.Messages <- message:
		default:
			// Skip if channel is full
		}
	}
}

func (s *SSEServer) logActiveClients() {
    s.mu.RLock()
    defer s.mu.RUnlock()
    PrintLog(fmt.Sprintf("Active Clients: %v", s.clients))
}


type noTimeoutWriter struct {
    http.ResponseWriter
}

func (w *noTimeoutWriter) Write(p []byte) (int, error) {
    // Disable write timeout
    if conn, _, err := w.ResponseWriter.(http.Hijacker).Hijack(); err == nil {
        conn.SetWriteDeadline(time.Time{})
        conn.Close()
    }
    return w.ResponseWriter.Write(p)
}

func (s *SSEServer) SSEHandler(w http.ResponseWriter, r *http.Request) {
	

	PrintLog(fmt.Sprintf("SSE connection attempt from %s", r.RemoteAddr))

	
	// Get user from context (set by AuthMiddleware)
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")


	// Create a flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Add client to server
	client := s.AddClient(userID)
   	s.logActiveClients()
	
	defer func() {
		PrintLog(fmt.Sprintf("Removing client %d (reason: connection closing)", int(userID)))
		s.RemoveClient(userID)
	}()


	// Send initial connection message
	//fmt.Fprintf(w, "event: connected\ndata: %s\n\n", "SSE connection established")
	//flusher.Flush()

	// Keep connection alive and send messages
	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()
	
	for {
		select {
		case msg := <-client.Messages:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-heartbeatTicker.C:
			// Send heartbeat to keep connection alive
			// Verify client is still active
			/*if time.Since(client.LastActive) > 90*time.Second {
			    PrintLog(fmt.Sprintf("Client %d timed out", userID))
			    return
			}*/
			client.LastActive = time.Now()
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			PrintLog(fmt.Sprintf("Client %d disconnected (context canceled)", userID))
			return
		}
	}
}

type NotificationMessage struct {
	Type    string      `json:"type"`    // "create", "update", "delete"
	Entity  string      `json:"entity"`  // "assignment", "course"
	ID      string      `json:"id"`      // Entity ID
	Message string      `json:"message"` // Human-readable message
	Data    interface{} `json:"data"`    // The actual data
}

func (s *SSEServer) SendNotification(userID uint, msgType, entity, id, message string, data interface{}) {
	notification := NotificationMessage{
		Type:    msgType,
		Entity:  entity,
		ID:      id,
		Message: message,
		Data:    data,
	}

	jsonData, err := json.Marshal(notification)
	if err != nil {
		return
	}

	s.SendToUser(userID, jsonData)
}
